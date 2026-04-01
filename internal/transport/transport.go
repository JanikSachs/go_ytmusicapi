// Package transport handles HTTP communication with the YouTube Music API.
package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	YTMDomain  = "https://music.youtube.com"
	YTMBaseAPI = YTMDomain + "/youtubei/v1/"
	YTMParams  = "?alt=json"
	UserAgent  = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:88.0) Gecko/20100101 Firefox/88.0"
)

// Middleware wraps an http.RoundTripper, enabling request/response interception.
type Middleware func(http.RoundTripper) http.RoundTripper

// RetryPolicy configures automatic retry behaviour for failed requests.
// Only transient errors (network failures and HTTP 5xx responses) trigger retries.
type RetryPolicy struct {
	MaxAttempts int           // total number of attempts; 0 or 1 means no retry
	WaitBetween time.Duration // pause between attempts
}

// Client wraps an http.Client and holds base headers/cookies.
type Client struct {
	http        *http.Client
	headers     map[string]string
	retryPolicy RetryPolicy
}

// ClientOption configures a transport Client.
type ClientOption func(*Client)

// WithTimeout sets the HTTP request timeout (default: 30s).
func WithTimeout(d time.Duration) ClientOption {
	return func(c *Client) { c.http.Timeout = d }
}

// WithHTTPClient replaces the underlying http.Client entirely.
func WithHTTPClient(h *http.Client) ClientOption {
	return func(c *Client) { c.http = h }
}

// WithUserAgent overrides the default User-Agent header.
func WithUserAgent(ua string) ClientOption {
	return func(c *Client) { c.headers["User-Agent"] = ua }
}

// WithExtraHeaders merges additional headers into every request.
func WithExtraHeaders(h map[string]string) ClientOption {
	return func(c *Client) {
		for k, v := range h {
			c.headers[k] = v
		}
	}
}

// WithRetry configures the retry policy for transient failures.
func WithRetry(p RetryPolicy) ClientOption {
	return func(c *Client) { c.retryPolicy = p }
}

// WithMiddleware wraps the underlying RoundTripper with one or more middleware.
// Middleware is applied in the order provided.
func WithMiddleware(mw ...Middleware) ClientOption {
	return func(c *Client) {
		rt := c.http.Transport
		if rt == nil {
			rt = http.DefaultTransport
		}
		for _, m := range mw {
			rt = m(rt)
		}
		c.http.Transport = rt
	}
}

// NewClient creates a Client with the provided base headers and options.
func NewClient(headers map[string]string, opts ...ClientOption) *Client {
	h := make(map[string]string, len(headers))
	for k, v := range headers {
		h[k] = v
	}
	c := &Client{
		http:    &http.Client{Timeout: 30 * time.Second},
		headers: h,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Post sends a POST request to endpoint with JSON body and returns decoded JSON.
// It retries according to the configured RetryPolicy.
func (c *Client) Post(ctx context.Context, endpoint string, params string, body map[string]any, extraHeaders map[string]string) (map[string]any, error) {
	maxAttempts := 1
	if c.retryPolicy.MaxAttempts > 1 {
		maxAttempts = c.retryPolicy.MaxAttempts
	}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 && c.retryPolicy.WaitBetween > 0 {
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("transport: retry wait: %w", ctx.Err())
			case <-time.After(c.retryPolicy.WaitBetween):
			}
		}

		result, err := c.doPost(ctx, endpoint, params, body, extraHeaders)
		if err == nil {
			return result, nil
		}
		lastErr = err

		// Do not retry on client errors (4xx) or parse errors.
		var parseErr *ParseError
		if errors.As(err, &parseErr) {
			break
		}
		var serverErr *ServerError
		if errors.As(err, &serverErr) && serverErr.StatusCode < 500 {
			break
		}
	}
	return nil, lastErr
}

func (c *Client) doPost(ctx context.Context, endpoint string, params string, body map[string]any, extraHeaders map[string]string) (map[string]any, error) {
	url := YTMBaseAPI + endpoint + YTMParams + params

	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("transport: marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("transport: create request: %w", err)
	}

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}
	req.AddCookie(&http.Cookie{Name: "SOCS", Value: "CAI"})

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("transport: execute request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("transport: read response body: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, &ParseError{
			Message: fmt.Sprintf("decode response JSON: %v", err),
			Cause:   err,
		}
	}

	if resp.StatusCode >= 400 {
		msg := fmt.Sprintf("server returned HTTP %d: %s", resp.StatusCode, resp.Status)
		if errObj, ok := result["error"].(map[string]any); ok {
			if errMsg, ok := errObj["message"].(string); ok {
				msg += ": " + errMsg
			}
		}
		return nil, &ServerError{Message: msg, StatusCode: resp.StatusCode}
	}

	return result, nil
}

// Get sends a GET request and returns the raw response body.
func (c *Client) Get(ctx context.Context, url string, extraHeaders map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("transport: create GET request: %w", err)
	}
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}
	req.AddCookie(&http.Cookie{Name: "SOCS", Value: "CAI"})

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("transport: execute GET request: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("transport: read GET response: %w", err)
	}
	return data, nil
}

// ServerError is returned when the API responds with an HTTP error status.
type ServerError struct {
	Message    string
	StatusCode int
}

func (e *ServerError) Error() string { return e.Message }

// ParseError is returned when the response body cannot be decoded as valid JSON.
type ParseError struct {
	Message string
	Cause   error
}

func (e *ParseError) Error() string { return "transport: " + e.Message }
func (e *ParseError) Unwrap() error { return e.Cause }
