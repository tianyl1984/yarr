package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"
)

const cookieName = "auth"

// SetSession stores a signed session cookie identifying the logged in user.
// The value is `username:hmac(username, secret)` so it cannot be forged
// without knowing the server secret.
func SetSession(rw http.ResponseWriter, username, secret string) {
	http.SetCookie(rw, &http.Cookie{
		Name:     cookieName,
		Value:    username + ":" + sign(username, secret),
		MaxAge:   604800, // 1 week
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// SessionUser returns the logged in username if the request carries a valid
// session cookie, or ("", false) otherwise.
func SessionUser(req *http.Request, secret string) (string, bool) {
	cookie, _ := req.Cookie(cookieName)
	if cookie == nil {
		return "", false
	}
	username, mac, found := strings.Cut(cookie.Value, ":")
	if !found || username == "" {
		return "", false
	}
	if !stringsEqual(mac, sign(username, secret)) {
		return "", false
	}
	return username, true
}

// Logout clears the session cookie.
func Logout(rw http.ResponseWriter) {
	http.SetCookie(rw, &http.Cookie{
		Name:   cookieName,
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})
}

func stringsEqual(p1, p2 string) bool {
	return subtle.ConstantTimeCompare([]byte(p1), []byte(p2)) == 1
}

func sign(msg, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(msg))
	return hex.EncodeToString(mac.Sum(nil))
}
