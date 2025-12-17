package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

// OAuthProvider OAuth 提供商接口
type OAuthProvider interface {
	// GetAuthURL 获取授权 URL
	GetAuthURL(state string) string

	// ExchangeCode 交换 code 获取用户信息
	ExchangeCode(code string) (*OAuthUserInfo, error)
}

// OAuthUserInfo OAuth 用户信息
type OAuthUserInfo struct {
	Provider string // 提供商
	Subject  string // 唯一标识（OIDC sub / 微信 unionid）
	Nickname string // 昵称（可选）
	Avatar   string // 头像（可选，不存储）
}

// WeChatProvider 微信网站应用提供商
// 文档：https://developers.weixin.qq.com/doc/oplatform/Website_App/WeChat_Login/Wechat_Login.html
type WeChatProvider struct {
	appID       string
	appSecret   string
	redirectURI string
}

// NewWeChatProvider 创建微信提供商
// 安全说明：
// 1. appID 和 appSecret 从环境变量读取
// 2. 生产环境必须设置 WECHAT_APP_ID 和 WECHAT_APP_SECRET
func NewWeChatProvider() *WeChatProvider {
	return &WeChatProvider{
		appID:       os.Getenv("WECHAT_APP_ID"),
		appSecret:   os.Getenv("WECHAT_APP_SECRET"),
		redirectURI: os.Getenv("WECHAT_REDIRECT_URI"), // 例如：https://yourdomain.com/api/oauth/wechat/callback
	}
}

// GetAuthURL 获取微信授权 URL
// 返回二维码扫码登录地址
func (p *WeChatProvider) GetAuthURL(state string) string {
	params := url.Values{}
	params.Add("appid", p.appID)
	params.Add("redirect_uri", p.redirectURI)
	params.Add("response_type", "code")
	params.Add("scope", "snsapi_login") // 网站应用使用 snsapi_login
	params.Add("state", state)

	return "https://open.weixin.qq.com/connect/qrconnect?" + params.Encode() + "#wechat_redirect"
}

// ExchangeCode 交换 code 获取用户信息
// 步骤：
// 1. 用 code 换取 access_token
// 2. 用 access_token 获取用户信息（包含 unionid）
func (p *WeChatProvider) ExchangeCode(code string) (*OAuthUserInfo, error) {
	// 1. 获取 access_token
	accessToken, openID, err := p.getAccessToken(code)
	if err != nil {
		return nil, err
	}

	// 2. 获取用户信息
	userInfo, err := p.getUserInfo(accessToken, openID)
	if err != nil {
		return nil, err
	}

	return userInfo, nil
}

// getAccessToken 获取 access_token
func (p *WeChatProvider) getAccessToken(code string) (accessToken, openID string, err error) {
	params := url.Values{}
	params.Add("appid", p.appID)
	params.Add("secret", p.appSecret)
	params.Add("code", code)
	params.Add("grant_type", "authorization_code")

	url := "https://api.weixin.qq.com/sns/oauth2/access_token?" + params.Encode()

	resp, err := http.Get(url)
	if err != nil {
		return "", "", fmt.Errorf("failed to get access token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		AccessToken string `json:"access_token"`
		OpenID      string `json:"openid"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", "", fmt.Errorf("failed to parse response: %w", err)
	}

	if result.ErrCode != 0 {
		return "", "", fmt.Errorf("wechat api error: %s", result.ErrMsg)
	}

	return result.AccessToken, result.OpenID, nil
}

// getUserInfo 获取用户信息
func (p *WeChatProvider) getUserInfo(accessToken, openID string) (*OAuthUserInfo, error) {
	params := url.Values{}
	params.Add("access_token", accessToken)
	params.Add("openid", openID)

	url := "https://api.weixin.qq.com/sns/userinfo?" + params.Encode()

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		UnionID  string `json:"unionid"` // 优先使用 unionid
		OpenID   string `json:"openid"`  // 备用
		Nickname string `json:"nickname"`
		ErrCode  int    `json:"errcode"`
		ErrMsg   string `json:"errmsg"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat api error: %s", result.ErrMsg)
	}

	// 优先使用 unionid，如果没有则使用 openid
	subject := result.UnionID
	if subject == "" {
		subject = result.OpenID
	}

	return &OAuthUserInfo{
		Provider: "wechat",
		Subject:  subject,
		Nickname: result.Nickname,
	}, nil
}

// OIDCProvider OIDC 通用提供商（Google、Apple 等）
// TODO: 实现 OIDC 标准流程
type OIDCProvider struct {
	clientID     string
	clientSecret string
	redirectURI  string
	issuer       string // OIDC issuer URL
}

// NewOIDCProvider 创建 OIDC 提供商
func NewOIDCProvider(provider string) *OIDCProvider {
	// 根据 provider 加载不同的配置
	// 例如：google, apple, github 等
	return &OIDCProvider{
		clientID:     os.Getenv(fmt.Sprintf("%s_CLIENT_ID", provider)),
		clientSecret: os.Getenv(fmt.Sprintf("%s_CLIENT_SECRET", provider)),
		redirectURI:  os.Getenv(fmt.Sprintf("%s_REDIRECT_URI", provider)),
		issuer:       os.Getenv(fmt.Sprintf("%s_ISSUER", provider)),
	}
}

// GetAuthURL 获取 OIDC 授权 URL
func (p *OIDCProvider) GetAuthURL(state string) string {
	// TODO: 实现 OIDC discovery 和授权 URL 生成
	// 1. 从 issuer/.well-known/openid-configuration 获取 authorization_endpoint
	// 2. 构造授权 URL
	return "TODO: implement OIDC GetAuthURL"
}

// ExchangeCode 交换 code 获取用户信息
func (p *OIDCProvider) ExchangeCode(code string) (*OAuthUserInfo, error) {
	// TODO: 实现 OIDC token 交换和用户信息获取
	// 1. 用 code 换取 id_token 和 access_token
	// 2. 验证 id_token
	// 3. 解析 id_token 获取 sub（subject）
	return nil, errors.New("TODO: implement OIDC ExchangeCode")
}

// GetProvider 获取 OAuth 提供商
func GetProvider(provider string) (OAuthProvider, error) {
	switch provider {
	case "wechat":
		return NewWeChatProvider(), nil
	case "google", "apple", "github":
		return NewOIDCProvider(provider), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}
