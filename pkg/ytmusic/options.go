package ytmusic

import (
	"net/http"
	"time"

	"github.com/JanikSachs/go_ytmusicapi/internal/auth"
	"github.com/JanikSachs/go_ytmusicapi/internal/transport"
)

// Option configures the Client.
type Option func(*config)

type config struct {
	language        string
	location        string
	userID          string
	httpClient      *http.Client
	authCookie      string
	authOrigin      string
	oauthToken      string
	browserAuthFile string
	session         *auth.Session
	headers         map[string]string
	retryPolicy     transport.RetryPolicy
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

// WithTimeout sets the HTTP request timeout, overriding the default 30s.
// If WithHTTPClient is also used, the timeout is applied to that client.
func WithTimeout(d time.Duration) Option {
	return func(c *config) {
		if c.httpClient == nil {
			c.httpClient = &http.Client{}
		}
		c.httpClient.Timeout = d
	}
}

// WithHeaders adds extra HTTP headers to every request.
func WithHeaders(h map[string]string) Option {
	return func(c *config) { c.headers = h }
}

// WithSession injects a pre-built auth.Session directly, bypassing the
// WithBrowserAuth / WithOAuthToken / WithBrowserAuthFile options.
func WithSession(s *auth.Session) Option {
	return func(c *config) { c.session = s }
}

// WithBrowserAuth configures browser cookie authentication.
// cookieStr is the raw Cookie header; origin is the Origin header value.
func WithBrowserAuth(cookieStr, origin string) Option {
	return func(c *config) {
		c.authCookie = cookieStr
		c.authOrigin = origin
	}
}

// WithBrowserAuthFile loads browser cookie authentication from a file.
// The file should contain a single line with the raw Cookie header value.
func WithBrowserAuthFile(path, origin string) Option {
	return func(c *config) {
		c.browserAuthFile = path
		c.authOrigin = origin
	}
}

// WithOAuthToken configures OAuth Bearer token authentication.
func WithOAuthToken(token string) Option {
	return func(c *config) { c.oauthToken = token }
}

// WithRetry configures simple retry behaviour for transient errors (network
// failures and HTTP 5xx responses). Client errors (4xx) are never retried.
func WithRetry(maxAttempts int, waitBetween time.Duration) Option {
	return func(c *config) {
		c.retryPolicy = transport.RetryPolicy{
			MaxAttempts: maxAttempts,
			WaitBetween: waitBetween,
		}
	}
}
