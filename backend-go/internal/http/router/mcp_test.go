package router

import (
	"net/http/httptest"
	"neverbroke/internal/repository/postgres"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMCPRequiresAuthentication(t *testing.T) {
	r := Setup(postgres.NewPgStore(nil), nil)
	for _, token := range []string{"", "Bearer invalid"} {
		for _, method := range []string{"GET", "POST", "DELETE"} {
			req := httptest.NewRequest(method, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
			req.Header.Set("Authorization", token)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			require.Equal(t, 401, w.Code)
		}
	}
}
