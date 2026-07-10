package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ssoClient is reused for all requests to the SSO service.
var ssoClient = &http.Client{Timeout: 10 * time.Second}

// UserInfo is the subset of the GitHub user payload returned by the SSO
// `/userinfo` endpoint that we care about.
type UserInfo struct {
	Login string `json:"login"`
}

// LoginURL builds the SSO login entry point that, after a successful GitHub
// authorization, redirects the browser back to `callback?token=<token>`.
func LoginURL(authURL, callback string) string {
	base := strings.TrimRight(authURL, "/")
	return fmt.Sprintf("%s/login?callback=%s", base, url.QueryEscape(callback))
}

// FetchUserInfo exchanges a one-time SSO token for the authenticated user's
// info by calling `<authURL>/userinfo?token=<token>`.
func FetchUserInfo(authURL, token string) (*UserInfo, error) {
	base := strings.TrimRight(authURL, "/")
	endpoint := fmt.Sprintf("%s/userinfo?token=%s", base, url.QueryEscape(token))

	resp, err := ssoClient.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sso userinfo failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var info UserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, err
	}
	if info.Login == "" {
		return nil, fmt.Errorf("sso userinfo missing login")
	}
	return &info, nil
}
