package ytmusic_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
