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

// makeSearchResponseWithContinuation returns a search response that includes a
// continuation token in the musicShelfRenderer, simulating a multi-page result.
func makeSearchResponseWithContinuation(filter, token string) map[string]any {
	shelf := map[string]any{
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
										map[string]any{"text": "Page 1 Song"},
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
											"videoId": "page1VideoId",
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
	}
	if token != "" {
		shelf["continuations"] = []any{
			map[string]any{
				"nextContinuationData": map[string]any{
					"continuation": token,
				},
			},
		}
	}
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
											"musicShelfRenderer": shelf,
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

// makeContinuationResponse returns a response that mimics the YouTube Music API
// pagination follow-up, using the continuationContents key.
func makeContinuationResponse(nextToken string) map[string]any {
	shelf := map[string]any{
		"contents": []any{
			map[string]any{
				"musicResponsiveListItemRenderer": map[string]any{
					"flexColumns": []any{
						map[string]any{
							"musicResponsiveListItemFlexColumnRenderer": map[string]any{
								"text": map[string]any{
									"runs": []any{
										map[string]any{"text": "Page 2 Song"},
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
											"videoId": "page2VideoId",
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
	}
	if nextToken != "" {
		shelf["continuations"] = []any{
			map[string]any{
				"nextContinuationData": map[string]any{
					"continuation": nextToken,
				},
			},
		}
	}
	return map[string]any{
		"continuationContents": map[string]any{
			"musicShelfContinuation": shelf,
		},
	}
}

func TestSearchPage_NoContinuation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(makeSearchResponseWithContinuation("songs", ""))
	}))
	defer srv.Close()

	client, err := ytmusic.NewClient(ytmusic.WithHTTPClient(
		&http.Client{Transport: &mockRoundTripper{srv: srv, inner: http.DefaultTransport}},
	))
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}

	page, err := client.SearchPage(context.Background(), "test query", ytmusic.SearchOptions{Filter: "songs"})
	if err != nil {
		t.Fatalf("SearchPage() error: %v", err)
	}
	if len(page.Results) == 0 {
		t.Error("expected at least one result")
	}
	if page.HasMore {
		t.Error("expected HasMore=false when no continuation token present")
	}
	if page.Continuation != "" {
		t.Errorf("expected empty continuation, got %q", page.Continuation)
	}
}

func TestSearchPage_WithContinuation(t *testing.T) {
	const expectedToken = "CBQQAA==testtoken"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(makeSearchResponseWithContinuation("songs", expectedToken))
	}))
	defer srv.Close()

	client, err := ytmusic.NewClient(ytmusic.WithHTTPClient(
		&http.Client{Transport: &mockRoundTripper{srv: srv, inner: http.DefaultTransport}},
	))
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}

	page, err := client.SearchPage(context.Background(), "test query")
	if err != nil {
		t.Fatalf("SearchPage() error: %v", err)
	}
	if !page.HasMore {
		t.Error("expected HasMore=true when continuation token is present")
	}
	if page.Continuation != expectedToken {
		t.Errorf("expected continuation %q, got %q", expectedToken, page.Continuation)
	}
}

func TestSearchNext_FetchesNextPage(t *testing.T) {
	const firstPageToken = "CBQQAA==testtoken"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Return a page 2 response with no further continuation.
		json.NewEncoder(w).Encode(makeContinuationResponse(""))
	}))
	defer srv.Close()

	client, err := ytmusic.NewClient(ytmusic.WithHTTPClient(
		&http.Client{Transport: &mockRoundTripper{srv: srv, inner: http.DefaultTransport}},
	))
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}

	page, err := client.SearchNext(context.Background(), firstPageToken)
	if err != nil {
		t.Fatalf("SearchNext() error: %v", err)
	}
	if len(page.Results) == 0 {
		t.Error("expected at least one result on page 2")
	}
	if page.Results[0].Title != "Page 2 Song" {
		t.Errorf("expected title 'Page 2 Song', got %q", page.Results[0].Title)
	}
	if page.HasMore {
		t.Error("expected HasMore=false on last page")
	}
}

func TestSearchNext_MultiStep(t *testing.T) {
	const page1Token = "token_page1"
	const page2Token = "token_page2"

	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		callCount++
		switch callCount {
		case 1:
			// First search call returns page 1 with a continuation token.
			json.NewEncoder(w).Encode(makeSearchResponseWithContinuation("songs", page1Token))
		case 2:
			// Continuation call returns page 2 with another continuation token.
			json.NewEncoder(w).Encode(makeContinuationResponse(page2Token))
		default:
			// Page 3 – final page, no continuation.
			json.NewEncoder(w).Encode(makeContinuationResponse(""))
		}
	}))
	defer srv.Close()

	client, err := ytmusic.NewClient(ytmusic.WithHTTPClient(
		&http.Client{Transport: &mockRoundTripper{srv: srv, inner: http.DefaultTransport}},
	))
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}

	// Page 1.
	page1, err := client.SearchPage(context.Background(), "test query")
	if err != nil {
		t.Fatalf("SearchPage() error: %v", err)
	}
	if !page1.HasMore {
		t.Error("page 1: expected HasMore=true")
	}

	// Page 2.
	page2, err := client.SearchNext(context.Background(), page1.Continuation)
	if err != nil {
		t.Fatalf("SearchNext(page1) error: %v", err)
	}
	if !page2.HasMore {
		t.Error("page 2: expected HasMore=true")
	}

	// Page 3 (last).
	page3, err := client.SearchNext(context.Background(), page2.Continuation)
	if err != nil {
		t.Fatalf("SearchNext(page2) error: %v", err)
	}
	if page3.HasMore {
		t.Error("page 3: expected HasMore=false")
	}
	if page3.Continuation != "" {
		t.Errorf("page 3: expected empty continuation, got %q", page3.Continuation)
	}
}

func TestSearchNext_EmptyToken(t *testing.T) {
	client, err := ytmusic.NewClient()
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}

	page, err := client.SearchNext(context.Background(), "")
	if err != nil {
		t.Fatalf("SearchNext(\"\") unexpected error: %v", err)
	}
	if page.HasMore {
		t.Error("expected HasMore=false for empty continuation token")
	}
	if len(page.Results) != 0 {
		t.Errorf("expected no results for empty continuation token, got %d", len(page.Results))
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
