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

// ---------------------------------------------------------------------------
// helper: make a test server returning the given body
// ---------------------------------------------------------------------------

func makeTestClient(t *testing.T, body map[string]any) (*ytmusic.Client, *httptest.Server) {
t.Helper()
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(body)
}))
mockTransport := &mockRoundTripper{srv: srv, inner: http.DefaultTransport}
client, err := ytmusic.NewClient(ytmusic.WithHTTPClient(&http.Client{Transport: mockTransport}))
if err != nil {
t.Fatalf("NewClient() error: %v", err)
}
return client, srv
}

// ---------------------------------------------------------------------------
// GetArtist tests
// ---------------------------------------------------------------------------

func TestGetArtist_Empty(t *testing.T) {
client, srv := makeTestClient(t, map[string]any{})
defer srv.Close()

artist, err := client.GetArtist(context.Background(), "UCsomeArtistId")
if err != nil {
t.Fatalf("GetArtist() unexpected error: %v", err)
}
if artist == nil {
t.Fatal("expected non-nil artist")
}
}

func TestGetArtist_EmptyID(t *testing.T) {
client, err := ytmusic.NewClient()
if err != nil {
t.Fatalf("NewClient() error: %v", err)
}
_, err = client.GetArtist(context.Background(), "")
if err == nil {
t.Error("expected error for empty browseId")
}
}

// ---------------------------------------------------------------------------
// GetAlbum tests
// ---------------------------------------------------------------------------

func makeAlbumResponse() map[string]any {
return map[string]any{
"contents": map[string]any{
"twoColumnBrowseResultsRenderer": map[string]any{
"tabs": []any{
map[string]any{
"tabRenderer": map[string]any{
"content": map[string]any{
"sectionListRenderer": map[string]any{
"contents": []any{
map[string]any{
"musicResponsiveHeaderRenderer": map[string]any{
"title": map[string]any{
"runs": []any{map[string]any{"text": "Test Album"}},
},
"subtitle": map[string]any{
"runs": []any{
map[string]any{"text": "Album"},
map[string]any{"text": " • "},
map[string]any{"text": "2023"},
},
},
"secondSubtitle": map[string]any{
"runs": []any{
map[string]any{"text": "10 songs"},
map[string]any{"text": " • "},
map[string]any{"text": "45 minutes"},
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
"secondaryContents": map[string]any{
"sectionListRenderer": map[string]any{
"contents": []any{
map[string]any{
"musicShelfRenderer": map[string]any{
"contents": []any{},
},
},
},
},
},
},
},
}
}

func TestGetAlbum_Basic(t *testing.T) {
client, srv := makeTestClient(t, makeAlbumResponse())
defer srv.Close()

album, err := client.GetAlbum(context.Background(), "MPREb_testAlbum123")
if err != nil {
t.Fatalf("GetAlbum() unexpected error: %v", err)
}
if album.Title != "Test Album" {
t.Errorf("Title = %q, want %q", album.Title, "Test Album")
}
if album.Year != "2023" {
t.Errorf("Year = %q, want %q", album.Year, "2023")
}
if album.TrackCount != 10 {
t.Errorf("TrackCount = %d, want %d", album.TrackCount, 10)
}
if album.Tracks == nil {
t.Error("Tracks must not be nil")
}
}

func TestGetAlbum_InvalidBrowseID(t *testing.T) {
client, err := ytmusic.NewClient()
if err != nil {
t.Fatalf("NewClient() error: %v", err)
}
_, err = client.GetAlbum(context.Background(), "UCsomething")
if err == nil {
t.Error("expected error for non-MPRE browseId")
}
}

func TestGetAlbum_EmptyID(t *testing.T) {
client, err := ytmusic.NewClient()
if err != nil {
t.Fatalf("NewClient() error: %v", err)
}
_, err = client.GetAlbum(context.Background(), "")
if err == nil {
t.Error("expected error for empty browseId")
}
}

// ---------------------------------------------------------------------------
// GetPlaylist tests
// ---------------------------------------------------------------------------

func makePlaylistResponse() map[string]any {
return map[string]any{
"contents": map[string]any{
"twoColumnBrowseResultsRenderer": map[string]any{
"tabs": []any{
map[string]any{
"tabRenderer": map[string]any{
"content": map[string]any{
"sectionListRenderer": map[string]any{
"contents": []any{
map[string]any{
"musicResponsiveHeaderRenderer": map[string]any{
"title": map[string]any{
"runs": []any{map[string]any{"text": "My Playlist"}},
},
"subtitle": map[string]any{
"runs": []any{
map[string]any{"text": "Playlist"},
map[string]any{"text": " • "},
map[string]any{"text": "2024"},
},
},
"secondSubtitle": map[string]any{
"runs": []any{
map[string]any{"text": "5 tracks"},
map[string]any{"text": " • "},
map[string]any{"text": "20 minutes"},
},
},
"straplineTextOne": map[string]any{
"runs": []any{map[string]any{"text": "Author Name"}},
},
},
},
},
},
},
},
},
},
"secondaryContents": map[string]any{
"sectionListRenderer": map[string]any{
"contents": []any{
map[string]any{
"musicPlaylistShelfRenderer": map[string]any{
"contents": []any{},
},
},
},
},
},
},
},
}
}

func TestGetPlaylist_Basic(t *testing.T) {
client, srv := makeTestClient(t, makePlaylistResponse())
defer srv.Close()

playlist, err := client.GetPlaylist(context.Background(), "PLtest123")
if err != nil {
t.Fatalf("GetPlaylist() unexpected error: %v", err)
}
if playlist.Title != "My Playlist" {
t.Errorf("Title = %q, want %q", playlist.Title, "My Playlist")
}
if playlist.Year != "2024" {
t.Errorf("Year = %q, want %q", playlist.Year, "2024")
}
if playlist.Author == nil || playlist.Author.Name != "Author Name" {
t.Errorf("Author = %v, want 'Author Name'", playlist.Author)
}
if playlist.Tracks == nil {
t.Error("Tracks must not be nil")
}
}

func TestGetPlaylist_EmptyID(t *testing.T) {
client, err := ytmusic.NewClient()
if err != nil {
t.Fatalf("NewClient() error: %v", err)
}
_, err = client.GetPlaylist(context.Background(), "")
if err == nil {
t.Error("expected error for empty playlistId")
}
}

// ---------------------------------------------------------------------------
// GetSong tests
// ---------------------------------------------------------------------------

func makeSongResponse(videoID, title string) map[string]any {
return map[string]any{
"videoDetails": map[string]any{
"videoId":    videoID,
"title":      title,
"viewCount":  "1000000",
},
}
}

func TestGetSong_Basic(t *testing.T) {
client, srv := makeTestClient(t, makeSongResponse("dQw4w9WgXcQ", "Never Gonna Give You Up"))
defer srv.Close()

song, err := client.GetSong(context.Background(), "dQw4w9WgXcQ")
if err != nil {
t.Fatalf("GetSong() unexpected error: %v", err)
}
if song.VideoID != "dQw4w9WgXcQ" {
t.Errorf("VideoID = %q, want %q", song.VideoID, "dQw4w9WgXcQ")
}
if song.Title != "Never Gonna Give You Up" {
t.Errorf("Title = %q, want %q", song.Title, "Never Gonna Give You Up")
}
if song.Views != "1000000" {
t.Errorf("Views = %q, want %q", song.Views, "1000000")
}
}

func TestGetSong_EmptyID(t *testing.T) {
client, err := ytmusic.NewClient()
if err != nil {
t.Fatalf("NewClient() error: %v", err)
}
_, err = client.GetSong(context.Background(), "")
if err == nil {
t.Error("expected error for empty videoId")
}
}
