package mcpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

type authTransport string

func (a authTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("Authorization", string(a))
	return http.DefaultTransport.RoundTrip(r)
}
func connect(t *testing.T, endpoint, token string) *mcp.ClientSession {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	c := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil)
	session, err := c.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: endpoint, HTTPClient: &http.Client{Transport: authTransport(token)}}, nil)
	require.NoError(t, err)
	t.Cleanup(func() { session.Close() })
	return session
}

func TestHTTPTools(t *testing.T) {
	var mu sync.Mutex
	calls := 0
	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls++
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/accounts":
			json.NewEncoder(w).Encode(map[string]any{"code": 0, "data": r.Header.Get("Authorization")})
		case "/api/v1/transactions/daily":
			require.Equal(t, "2026-10-30", r.URL.Query().Get("date"))
			// Existing REST APIs encode business errors in HTTP 200 responses.
			w.Write([]byte(`{"code":400,"message":"invalid request"}`))
		case "/api/v1/transactions":
			require.Equal(t, "POST", r.Method)
			var body map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			require.Equal(t, "account-1", body["accountId"])
			json.NewEncoder(w).Encode(map[string]any{"code": 0, "data": body})
		default:
			t.Errorf("unexpected endpoint %s", r.URL)
			w.WriteHeader(404)
		}
	})
	server := httptest.NewServer(NewHandler(api))
	defer server.Close()
	alice := connect(t, server.URL, "Bearer alice")
	bob := connect(t, server.URL, "Bearer bob")
	ctx := context.Background()
	list, err := alice.ListTools(ctx, nil)
	require.NoError(t, err)
	require.Len(t, list.Tools, 12)
	for _, tool := range list.Tools {
		require.NotNil(t, tool.InputSchema)
		require.NotEmpty(t, tool.Description)
	}
	for _, user := range []struct {
		s     *mcp.ClientSession
		token string
	}{{alice, "alice"}, {bob, "bob"}, {alice, "alice"}} {
		res, err := user.s.CallTool(ctx, &mcp.CallToolParams{Name: "list_accounts", Arguments: map[string]any{}})
		require.NoError(t, err)
		require.False(t, res.IsError)
		require.Contains(t, res.Content[0].(*mcp.TextContent).Text, "Bearer "+user.token)
		require.NotNil(t, res.StructuredContent)
	}
	res, err := alice.CallTool(ctx, &mcp.CallToolParams{Name: "get_daily_transactions", Arguments: map[string]any{"date": "2026-10-30"}})
	require.NoError(t, err)
	require.True(t, res.IsError)
	res, err = alice.CallTool(ctx, &mcp.CallToolParams{Name: "create_transaction", Arguments: map[string]any{"type": "expense", "amount": 12.5, "accountId": "account-1", "occurredAt": "2026-10-30T12:00:00+08:00"}})
	require.NoError(t, err)
	require.False(t, res.IsError)
	mu.Lock()
	before := calls
	mu.Unlock()
	for _, tc := range []struct {
		name string
		args map[string]any
	}{
		{"get_daily_transactions", map[string]any{"date": "bad"}},
		{"get_monthly_budget", map[string]any{}},
		{"set_monthly_budget", map[string]any{"month": "2026-10", "totalBudget": -1}},
		{"forecast_cashflow", map[string]any{"startDate": "2026-10-30", "days": 10000}},
		{"forecast_cashflow", map[string]any{"startDate": "2026-10-30", "days": 0}},
		{"create_transaction", map[string]any{"type": "expense", "amount": 0, "occurredAt": "bad"}},
		{"unknown_tool", map[string]any{}},
	} {
		res, err := alice.CallTool(ctx, &mcp.CallToolParams{Name: tc.name, Arguments: tc.args})
		require.True(t, err != nil || res.IsError, tc.name)
	}
	mu.Lock()
	require.Equal(t, before, calls)
	mu.Unlock()
}

func TestOriginAndBodyLimit(t *testing.T) {
	handler := NewHandler(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Fatal("unexpected API call") }))
	for _, origin := range []string{"https://evil.example", "null", "http://localhost:3000/evil"} {
		r := httptest.NewRequest("POST", "http://localhost:8080/mcp", strings.NewReader(`{}`))
		r.Header.Set("Origin", origin)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		require.Equal(t, 403, w.Code)
	}
	r := httptest.NewRequest("POST", "http://localhost:8080/mcp", strings.NewReader(strings.Repeat(" ", 1<<20)+`{}`))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json, text/event-stream")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	require.Equal(t, 413, w.Code)
}
