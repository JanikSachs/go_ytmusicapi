package transport_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JanikSachs/go_ytmusicapi/internal/transport"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// buildClient creates a transport.Client pointed at srv, redirecting all requests to it.
func buildClient(srv *httptest.Server, opts ...transport.ClientOption) *transport.Client {
	base := map[string]string{"Content-Type": "application/json"}
	redirect := transport.WithMiddleware(func(next http.RoundTripper) http.RoundTripper {
		return roundTripFunc(func(req *http.Request) (*http.Response, error) {
			req2 := req.Clone(req.Context())
			req2.URL.Scheme = "http"
			req2.URL.Host = srv.Listener.Addr().String()
			return http.DefaultTransport.RoundTrip(req2)
		})
	})
	return transport.NewClient(base, append([]transport.ClientOption{redirect}, opts...)...)
}

func TestPost_Success(t *testing.T) {
	want := map[string]any{"ok": true}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	c := buildClient(srv)
	got, err := c.Post(context.Background(), "search", "", map[string]any{"q": "test"}, nil)
	if err != nil {
		t.Fatalf("Post() unexpected error: %v", err)
	}
	if got["ok"] != true {
		t.Errorf("Post() got %v, want ok=true", got)
	}
}

func TestPost_ServerError_4xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "invalid credentials"},
		})
	}))
	defer srv.Close()

	c := buildClient(srv)
	_, err := c.Post(context.Background(), "search", "", map[string]any{}, nil)
	if err == nil {
		t.Fatal("Post() expected error, got nil")
	}

	var serverErr *transport.ServerError
	if !errors.As(err, &serverErr) {
		t.Fatalf("expected *transport.ServerError, got %T: %v", err, err)
	}
	if serverErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, want %d", serverErr.StatusCode, http.StatusUnauthorized)
	}
}

func TestPost_ServerError_5xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{})
	}))
	defer srv.Close()

	c := buildClient(srv)
	_, err := c.Post(context.Background(), "search", "", map[string]any{}, nil)
	if err == nil {
		t.Fatal("Post() expected error, got nil")
	}

	var serverErr *transport.ServerError
	if !errors.As(err, &serverErr) {
		t.Fatalf("expected *transport.ServerError, got %T: %v", err, err)
	}
	if serverErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", serverErr.StatusCode, http.StatusInternalServerError)
	}
}

func TestPost_ParseError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not valid json {{{{"))
	}))
	defer srv.Close()

	c := buildClient(srv)
	_, err := c.Post(context.Background(), "search", "", map[string]any{}, nil)
	if err == nil {
		t.Fatal("Post() expected parse error, got nil")
	}

	var parseErr *transport.ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("expected *transport.ParseError, got %T: %v", err, err)
	}
}

func TestPost_ExtraHeaders(t *testing.T) {
	var gotHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Custom")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{})
	}))
	defer srv.Close()

	c := buildClient(srv)
	_, err := c.Post(context.Background(), "search", "", map[string]any{},
		map[string]string{"X-Custom": "test-value"})
	if err != nil {
		t.Fatalf("Post() unexpected error: %v", err)
	}
	if gotHeader != "test-value" {
		t.Errorf("X-Custom header = %q, want %q", gotHeader, "test-value")
	}
}

func TestPost_Retry_OnServerError5xx(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		w.Header().Set("Content-Type", "application/json")
		if n < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]any{})
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer srv.Close()

	policy := transport.RetryPolicy{MaxAttempts: 3, WaitBetween: 0}
	c := buildClient(srv, transport.WithRetry(policy))

	got, err := c.Post(context.Background(), "search", "", map[string]any{}, nil)
	if err != nil {
		t.Fatalf("Post() unexpected error after retry: %v", err)
	}
	if got["ok"] != true {
		t.Errorf("Post() got %v, want ok=true", got)
	}
	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestPost_Retry_NoRetryOn4xx(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{})
	}))
	defer srv.Close()

	policy := transport.RetryPolicy{MaxAttempts: 3, WaitBetween: 0}
	c := buildClient(srv, transport.WithRetry(policy))

	_, err := c.Post(context.Background(), "search", "", map[string]any{}, nil)
	if err == nil {
		t.Fatal("Post() expected error, got nil")
	}
	if atomic.LoadInt32(&attempts) != 1 {
		t.Errorf("expected 1 attempt (no retry on 4xx), got %d", attempts)
	}
}

func TestPost_Retry_NoRetryOnParseError(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.Write([]byte("not json"))
	}))
	defer srv.Close()

	policy := transport.RetryPolicy{MaxAttempts: 3, WaitBetween: 0}
	c := buildClient(srv, transport.WithRetry(policy))

	_, err := c.Post(context.Background(), "search", "", map[string]any{}, nil)
	if err == nil {
		t.Fatal("Post() expected error, got nil")
	}
	if atomic.LoadInt32(&attempts) != 1 {
		t.Errorf("expected 1 attempt (no retry on parse error), got %d", attempts)
	}
}

func TestPost_Retry_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{})
	}))
	defer srv.Close()

	policy := transport.RetryPolicy{MaxAttempts: 5, WaitBetween: 100 * time.Millisecond}
	c := buildClient(srv, transport.WithRetry(policy))

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := c.Post(ctx, "search", "", map[string]any{}, nil)
	if err == nil {
		t.Fatal("Post() expected error, got nil")
	}
}

func TestMiddleware(t *testing.T) {
	var middlewareCalled bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{})
	}))
	defer srv.Close()

	mw := func(next http.RoundTripper) http.RoundTripper {
		return roundTripFunc(func(req *http.Request) (*http.Response, error) {
			middlewareCalled = true
			req2 := req.Clone(req.Context())
			req2.URL.Scheme = "http"
			req2.URL.Host = srv.Listener.Addr().String()
			return next.RoundTrip(req2)
		})
	}

	base := map[string]string{"Content-Type": "application/json"}
	c := transport.NewClient(base, transport.WithMiddleware(mw))

	_, err := c.Post(context.Background(), "search", "", map[string]any{}, nil)
	if err != nil {
		t.Fatalf("Post() unexpected error: %v", err)
	}
	if !middlewareCalled {
		t.Error("middleware was not called")
	}
}

func TestWithTimeout(t *testing.T) {
	// Server that sleeps longer than the timeout.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{})
	}))
	defer srv.Close()

	c := buildClient(srv, transport.WithTimeout(50*time.Millisecond))
	_, err := c.Post(context.Background(), "search", "", map[string]any{}, nil)
	if err == nil {
		t.Fatal("Post() expected timeout error, got nil")
	}
}

func TestWithUserAgent(t *testing.T) {
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{})
	}))
	defer srv.Close()

	c := buildClient(srv, transport.WithUserAgent("custom-agent/1.0"))
	_, err := c.Post(context.Background(), "search", "", map[string]any{}, nil)
	if err != nil {
		t.Fatalf("Post() unexpected error: %v", err)
	}
	if gotUA != "custom-agent/1.0" {
		t.Errorf("User-Agent = %q, want %q", gotUA, "custom-agent/1.0")
	}
}
