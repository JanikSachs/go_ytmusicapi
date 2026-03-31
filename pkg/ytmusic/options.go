package ytmusic

import (
	"net/http"
	"time"
)

// Option configures the Client.
type Option func(*config)

type config struct {
	language   string
	location   string
	userID     string
	httpClient *http.Client
	authCookie string
	authOrigin string
	oauthToken string
}

func defaultConfig() *config {
	return &config{
		language:   "en",
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// WithLanguage sets the response language (default: "en").
func WithLanguage(lang string) Option {
	return func(c *config) { c.language = lang }
}

// WithLocation sets the geographic location for results.
func WithLocation(loc string) Option {
	return func(c *config) { c.location = loc }
}

// WithUserID sets the brand account user ID for requests.
func WithUserID(id string) Option {
	return func(c *config) { c.userID = id }
}

// WithHTTPClient replaces the default HTTP client.
func WithHTTPClient(h *http.Client) Option {
	return func(c *config) { c.httpClient = h }
}

// WithBrowserAuth configures browser cookie authentication.
// cookieStr is the raw Cookie header; origin is the Origin header value.
func WithBrowserAuth(cookieStr, origin string) Option {
	return func(c *config) {
		c.authCookie = cookieStr
		c.authOrigin = origin
	}
}

// WithOAuthToken configures OAuth Bearer token authentication.
func WithOAuthToken(token string) Option {
	return func(c *config) { c.oauthToken = token }
}
