package ytmusic_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JanikSachs/go_ytmusicapi/pkg/ytmusic"
)

func makeSearchResponse(filter string) map[string]any {
	return map[string]any{
		"contents": map[string]any{
			"tabbedSearchResultsRenderer": map[string]any{
				"tabs": []any{
					map[string]any{
						"tabRenderer": map[string]any{
							"content": map[string]any{
								"sectionListRenderer": map[string]any{
									"contents": []any{
										map[string]any{
											"musicShelfRenderer": map[string]any{
												"title": map[string]any{
													"runs": []any{
														map[string]any{"text": "Songs"},
													},
												},
												"contents": []any{
													map[string]any{
														"musicResponsiveListItemRenderer": map[string]any{
															"flexColumns": []any{
																map[string]any{
																	"musicResponsiveListItemFlexColumnRenderer": map[string]any{
																		"text": map[string]any{
																			"runs": []any{
																				map[string]any{"text": "Test Song"},
																			},
																		},
																	},
																},
															},
															"overlay": map[string]any{
																"musicItemThumbnailOverlayRenderer": map[string]any{
																	"content": map[string]any{
																		"musicPlayButtonRenderer": map[string]any{
																			"playNavigationEndpoint": map[string]any{
																				"watchEndpoint": map[string]any{
																					"videoId": "testVideoId123",
																				},
																			},
																			"watchEndpoint": map[string]any{
																				"watchEndpointMusicSupportedConfigs": map[string]any{
																					"watchEndpointMusicConfig": map[string]any{
																						"musicVideoType": "MUSIC_VIDEO_TYPE_ATV",
																					},
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func TestSearch_Basic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(makeSearchResponse("songs"))
	}))
	defer srv.Close()

	mockTransport := &mockRoundTripper{srv: srv, inner: http.DefaultTransport}
	mockClient := &http.Client{Transport: mockTransport}

	client, err := ytmusic.NewClient(ytmusic.WithHTTPClient(mockClient))
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}

	results, err := client.Search(context.Background(), "test query", ytmusic.SearchOptions{
		Filter: "songs",
		Limit:  5,
	})
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}

	if len(results) == 0 {
		t.Error("expected at least one result")
	}
	if len(results) > 0 && results[0].Title != "Test Song" {
		t.Errorf("expected title 'Test Song', got %q", results[0].Title)
	}
}

func TestSearch_InvalidFilter(t *testing.T) {
	client, err := ytmusic.NewClient()
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}
	_, err = client.Search(context.Background(), "query", ytmusic.SearchOptions{
		Filter: "invalid_filter",
	})
	if err == nil {
		t.Error("expected error for invalid filter")
	}
}

func TestSearch_InvalidScope(t *testing.T) {
	client, err := ytmusic.NewClient()
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}
	_, err = client.Search(context.Background(), "query", ytmusic.SearchOptions{
		Scope: "invalid_scope",
	})
	if err == nil {
		t.Error("expected error for invalid scope")
	}
}

func TestNewClient_Default(t *testing.T) {
	client, err := ytmusic.NewClient()
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}
	if client == nil {
		t.Error("expected non-nil client")
	}
}

func TestNewClient_WithOptions(t *testing.T) {
	client, err := ytmusic.NewClient(
		ytmusic.WithLanguage("de"),
		ytmusic.WithLocation("DE"),
	)
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}
	if client == nil {
		t.Error("expected non-nil client")
	}
}

func TestNewClient_WithTimeout(t *testing.T) {
	client, err := ytmusic.NewClient(ytmusic.WithTimeout(5 * time.Second))
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}
	if client == nil {
		t.Error("expected non-nil client")
	}
}

func TestNewClient_WithHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Test-Header") != "hello" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(makeSearchResponse("songs"))
	}))
	defer srv.Close()

	mockTransport := &mockRoundTripper{srv: srv, inner: http.DefaultTransport}
	mockClient := &http.Client{Transport: mockTransport}

	client, err := ytmusic.NewClient(
		ytmusic.WithHTTPClient(mockClient),
		ytmusic.WithHeaders(map[string]string{"X-Test-Header": "hello"}),
	)
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}

	_, err = client.Search(context.Background(), "test")
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}
}

func TestNewClient_WithOAuthToken(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(makeSearchResponse("songs"))
	}))
	defer srv.Close()

	mockTransport := &mockRoundTripper{srv: srv, inner: http.DefaultTransport}
	mockClient := &http.Client{Transport: mockTransport}

	client, err := ytmusic.NewClient(
		ytmusic.WithHTTPClient(mockClient),
		ytmusic.WithOAuthToken("ya29.test"),
	)
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}

	_, err = client.Search(context.Background(), "test")
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}
	if gotAuth != "Bearer ya29.test" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer ya29.test")
	}
}

func TestNewClient_WithBrowserAuthFile(t *testing.T) {
	cookieStr := `__Secure-3PAPISID=abcdef1234567890/XYZ; HSID=hsidvalue`
	dir := t.TempDir()
	path := filepath.Join(dir, "cookies.txt")
	if err := os.WriteFile(path, []byte(cookieStr), 0600); err != nil {
		t.Fatalf("write cookie file: %v", err)
	}

	client, err := ytmusic.NewClient(ytmusic.WithBrowserAuthFile(path, "https://music.youtube.com"))
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}
	if client == nil {
		t.Error("expected non-nil client")
	}
}

func TestNewClient_WithBrowserAuthFile_Missing(t *testing.T) {
	_, err := ytmusic.NewClient(ytmusic.WithBrowserAuthFile("/nonexistent/cookies.txt", "https://music.youtube.com"))
	if err == nil {
		t.Fatal("NewClient() expected error for missing cookie file, got nil")
	}
}

func TestNewClient_WithRetry(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		w.Header().Set("Content-Type", "application/json")
		if n < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]any{})
			return
		}
		json.NewEncoder(w).Encode(makeSearchResponse("songs"))
	}))
	defer srv.Close()

	mockTransport := &mockRoundTripper{srv: srv, inner: http.DefaultTransport}
	mockClient := &http.Client{Transport: mockTransport}

	client, err := ytmusic.NewClient(
		ytmusic.WithHTTPClient(mockClient),
		ytmusic.WithRetry(3, 0),
	)
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}

	_, err = client.Search(context.Background(), "test")
	if err != nil {
		t.Fatalf("Search() unexpected error: %v", err)
	}
	if atomic.LoadInt32(&attempts) != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}

type mockRoundTripper struct {
	srv   *httptest.Server
	inner http.RoundTripper
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := req.Clone(req.Context())
	req2.URL.Scheme = "http"
	req2.URL.Host = m.srv.Listener.Addr().String()
	return m.inner.RoundTrip(req2)
}
