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
	AuthURL string
	Secret  string
	Public  []string
}

func (m *Middleware) Handler(c *router.Context) {
	for _, path := range m.Public {
		if strings.HasPrefix(c.Req.URL.Path, path) {
			c.Next()
			return
		}
	}

	if _, ok := SessionUser(c.Req, m.Secret); ok {
		c.Next()
		return
	}

	// Only direct browser navigation can be redirected server-side (here the
	// Host header is the browser's). In this fork that path is unused — the
	// SPA is served separately and only reaches us via XHR under /api.
	if c.Req.Method == http.MethodGet && acceptsHTML(c.Req) {
		c.Redirect(LoginURL(m.AuthURL, m.callbackURL(c)))
		return
	}

	// API/XHR: hand the SPA the SSO base URL and let it build the login URL
	// from window.location.origin. Behind the frontend proxy the backend
	// can't see the browser's real scheme/host/port (nginx drops the port,
	// Vite rewrites Host to its own), so the callback must be built client-side.
	c.Out.Header().Set("X-Auth-Url", m.AuthURL)
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

	return scheme + "://" + host + "/api/auth/callback"
}

func acceptsHTML(req *http.Request) bool {
	return strings.Contains(req.Header.Get("Accept"), "text/html")
}
