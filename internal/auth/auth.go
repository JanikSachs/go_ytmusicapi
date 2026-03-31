// Package auth handles authentication for the YouTube Music API.
package auth

import (
	"crypto/sha1"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Type enumerates authentication modes.
type Type int

const (
	Unauthorized Type = iota
	Browser           // cookie-based browser auth
	OAuth             // OAuth2 token auth
)

// Session holds authentication state and produces request headers.
type Session struct {
	authType  Type
	sapisid   string
	origin    string
	cookieStr string
	token     string // Bearer token for OAuth
}

// NewUnauthenticated returns a Session with no credentials.
func NewUnauthenticated() *Session {
	return &Session{authType: Unauthorized}
}

// NewBrowserAuth returns a Session using browser cookie authentication.
// cookieStr should be the raw cookie header value.
func NewBrowserAuth(cookieStr, origin string) (*Session, error) {
	sapisid, err := sapisidFromCookie(cookieStr)
	if err != nil {
		return nil, fmt.Errorf("auth: parse SAPISID from cookie: %w", err)
	}
	return &Session{
		authType:  Browser,
		sapisid:   sapisid,
		origin:    origin,
		cookieStr: cookieStr,
	}, nil
}

// NewOAuthAuth returns a Session using an OAuth Bearer token.
func NewOAuthAuth(bearerToken string) *Session {
	return &Session{authType: OAuth, token: bearerToken}
}

// AuthType returns the authentication type.
func (s *Session) AuthType() Type { return s.authType }

// Headers returns the authentication-specific headers to add to each request.
func (s *Session) Headers() map[string]string {
	headers := make(map[string]string)
	switch s.authType {
	case Browser:
		headers["Cookie"] = s.cookieStr
		headers["Authorization"] = getAuthorization(s.sapisid + " " + s.origin)
		headers["X-Origin"] = s.origin
	case OAuth:
		headers["Authorization"] = "Bearer " + s.token
		headers["X-Goog-Request-Time"] = strconv.FormatInt(time.Now().Unix(), 10)
	}
	return headers
}

// sapisidFromCookie extracts the __Secure-3PAPISID value from a raw cookie string.
func sapisidFromCookie(raw string) (string, error) {
	raw = strings.ReplaceAll(raw, `"`, "")
	header := http.Header{}
	header.Add("Cookie", raw)
	req := &http.Request{Header: header}
	cookie, err := req.Cookie("__Secure-3PAPISID")
	if err != nil {
		return "", fmt.Errorf("cookie __Secure-3PAPISID not found: %w", err)
	}
	return cookie.Value, nil
}

// getAuthorization generates a SAPISIDHASH from sapisid+origin.
func getAuthorization(auth string) string {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	h := sha1.New()
	h.Write([]byte(ts + " " + auth))
	return fmt.Sprintf("SAPISIDHASH %s_%x", ts, h.Sum(nil))
}
