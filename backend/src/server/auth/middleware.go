package auth

import (
	"net/http"
	"strings"

	"github.com/nkanaev/yarr/src/server/router"
)

// Middleware guards every request behind the cf-worker-auth SSO service.
// Requests without a valid session cookie are redirected to the SSO login
// page (for browser navigation) or rejected with 401 (for API calls).
type Middleware struct {
	BasePath string
	AuthURL  string
	Secret   string
	Public   []string
}

func (m *Middleware) Handler(c *router.Context) {
	for _, path := range m.Public {
		if strings.HasPrefix(c.Req.URL.Path, m.BasePath+path) {
			c.Next()
			return
		}
	}

	if _, ok := SessionUser(c.Req, m.Secret); ok {
		c.Next()
		return
	}

	loginURL := LoginURL(m.AuthURL, m.callbackURL(c))

	// Only redirect top-level browser navigation to the SSO login page.
	// API/XHR calls get a plain 401, but we expose the SSO login URL in a
	// header so the (statically-served) SPA can send the browser there
	// itself — a bare reload would just re-serve index.html and loop.
	if c.Req.Method == http.MethodGet && acceptsHTML(c.Req) {
		c.Redirect(loginURL)
		return
	}

	c.Out.Header().Set("X-Auth-Login-Url", loginURL)
	c.Out.WriteHeader(http.StatusUnauthorized)
}

// callbackURL builds this server's externally reachable /auth/callback URL,
// honouring the proxy-provided scheme/host when present.
func (m *Middleware) callbackURL(c *router.Context) string {
	scheme := "http"
	if proto := c.Req.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	} else if c.Req.TLS != nil {
		scheme = "https"
	}

	host := c.Req.Host
	if forwarded := c.Req.Header.Get("X-Forwarded-Host"); forwarded != "" {
		host = forwarded
	}

	return scheme + "://" + host + m.BasePath + "/api/auth/callback"
}

func acceptsHTML(req *http.Request) bool {
	return strings.Contains(req.Header.Get("Accept"), "text/html")
}
