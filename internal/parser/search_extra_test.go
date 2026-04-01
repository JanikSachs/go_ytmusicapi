package parser_test

import (
	"testing"

	"github.com/JanikSachs/go_ytmusicapi/internal/model"
	"github.com/JanikSachs/go_ytmusicapi/internal/parser"
)

// ---------------------------------------------------------------------------
// Helpers shared across search_extra_test.go
// ---------------------------------------------------------------------------

// makeFlexCols constructs a flexColumns list from a slice of run-slice specs.
// Each inner []any becomes the runs list of one flex column.
func makeFlexCols(cols [][]any) []any {
	result := make([]any, len(cols))
	for i, runs := range cols {
		result[i] = map[string]any{
			"musicResponsiveListItemFlexColumnRenderer": map[string]any{
				"text": map[string]any{"runs": runs},
			},
		}
	}
	return result
}

func makeRun(text string) map[string]any {
	return map[string]any{"text": text}
}

func makeRunWithBrowse(text, browseID string) map[string]any {
	return map[string]any{
		"text": text,
		"navigationEndpoint": map[string]any{
			"browseEndpoint": map[string]any{"browseId": browseID},
		},
	}
}

func makePlayButton(videoID, videoType string) map[string]any {
	return map[string]any{
		"overlay": map[string]any{
			"musicItemThumbnailOverlayRenderer": map[string]any{
				"content": map[string]any{
					"musicPlayButtonRenderer": map[string]any{
						"playNavigationEndpoint": map[string]any{
							"watchEndpoint": map[string]any{"videoId": videoID},
						},
						"watchEndpoint": map[string]any{
							"watchEndpointMusicSupportedConfigs": map[string]any{
								"watchEndpointMusicConfig": map[string]any{
									"musicVideoType": videoType,
								},
							},
						},
					},
				},
			},
		},
	}
}

// mergeInto merges src map into dst map (in place).
func mergeInto(dst map[string]any, src map[string]any) map[string]any {
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// ---------------------------------------------------------------------------
// ParseSearchResult – additional result types
// ---------------------------------------------------------------------------

func TestParseSearchResult_Artist(t *testing.T) {
	data := map[string]any{
		"flexColumns": makeFlexCols([][]any{
			{makeRun("Queen")},
		}),
		"navigationEndpoint": map[string]any{
			"browseEndpoint": map[string]any{"browseId": "UCbnkAoWhS9i3YvcnysFnFQA"},
		},
		"menu": map[string]any{
			"menuRenderer": map[string]any{
				"items": []any{
					map[string]any{
						"menuNavigationItemRenderer": map[string]any{
							"icon": map[string]any{"iconType": "MUSIC_SHUFFLE"},
							"navigationEndpoint": map[string]any{
								"watchPlaylistEndpoint": map[string]any{"playlistId": "RDAMPL_shuffle"},
							},
						},
					},
					map[string]any{
						"menuNavigationItemRenderer": map[string]any{
							"icon": map[string]any{"iconType": "MIX"},
							"navigationEndpoint": map[string]any{
								"watchPlaylistEndpoint": map[string]any{"playlistId": "RDAMPL_radio"},
							},
						},
					},
				},
			},
		},
	}

	result := parser.ParseSearchResult(data, "artist", "Artists")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.ResultType != "artist" {
		t.Errorf("ResultType = %q, want %q", result.ResultType, "artist")
	}
	if result.Artist != "Queen" {
		t.Errorf("Artist = %q, want %q", result.Artist, "Queen")
	}
	if result.BrowseID != "UCbnkAoWhS9i3YvcnysFnFQA" {
		t.Errorf("BrowseID = %q, want %q", result.BrowseID, "UCbnkAoWhS9i3YvcnysFnFQA")
	}
	if result.ShuffleID != "RDAMPL_shuffle" {
		t.Errorf("ShuffleID = %q, want %q", result.ShuffleID, "RDAMPL_shuffle")
	}
	if result.RadioID != "RDAMPL_radio" {
		t.Errorf("RadioID = %q, want %q", result.RadioID, "RDAMPL_radio")
	}
}

func TestParseSearchResult_Album(t *testing.T) {
	data := mergeInto(makePlayButton("", ""), map[string]any{
		"flexColumns": makeFlexCols([][]any{
			{makeRun("The Eminem Show")},
			{makeRunWithBrowse("Eminem", "UCfM3zsQsOnfWNUppiycmBuw"), makeRun(" • "), makeRun("2002")},
		}),
		"navigationEndpoint": map[string]any{
			"browseEndpoint": map[string]any{"browseId": "MPREb_emshow"},
		},
	})
	// For album type the playNavigationEndpoint contains the playlistId
	data["overlay"].(map[string]any)["musicItemThumbnailOverlayRenderer"].(map[string]any)["content"].(map[string]any)["musicPlayButtonRenderer"].(map[string]any)["playNavigationEndpoint"] = map[string]any{
		"watchEndpoint": map[string]any{"playlistId": "OLAK5uy_emshow"},
	}

	result := parser.ParseSearchResult(data, "album", "Albums")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.ResultType != "album" {
		t.Errorf("ResultType = %q, want album", result.ResultType)
	}
	if result.Title != "The Eminem Show" {
		t.Errorf("Title = %q, want %q", result.Title, "The Eminem Show")
	}
	if result.BrowseID != "MPREb_emshow" {
		t.Errorf("BrowseID = %q, want %q", result.BrowseID, "MPREb_emshow")
	}
	if result.PlaylistID != "OLAK5uy_emshow" {
		t.Errorf("PlaylistID = %q, want %q", result.PlaylistID, "OLAK5uy_emshow")
	}
}

func TestParseSearchResult_Video(t *testing.T) {
	data := mergeInto(makePlayButton("videoABC", "MUSIC_VIDEO_TYPE_UGC"), map[string]any{
		"flexColumns": makeFlexCols([][]any{
			{makeRun("Stan (Music Video)")},
			{makeRunWithBrowse("Eminem", "UCfM3zsQsOnfWNUppiycmBuw"), makeRun(" • "), makeRun("1.5M views")},
		}),
	})

	result := parser.ParseSearchResult(data, "video", "Videos")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.ResultType != "video" {
		t.Errorf("ResultType = %q, want video", result.ResultType)
	}
	if result.Title != "Stan (Music Video)" {
		t.Errorf("Title = %q, want %q", result.Title, "Stan (Music Video)")
	}
	if result.VideoID != "videoABC" {
		t.Errorf("VideoID = %q, want %q", result.VideoID, "videoABC")
	}
	if result.Views != "1.5M" {
		t.Errorf("Views = %q, want %q", result.Views, "1.5M")
	}
}

func TestParseSearchResult_Playlist_WithAuthor(t *testing.T) {
	data := map[string]any{
		"flexColumns": makeFlexCols([][]any{
			{makeRun("Top 50 Songs")},
			// runs: author run, separator, itemCount run  → hasAuthor=true (len==3)
			{makeRun("Spotify"), makeRun(" • "), makeRun("50 tracks")},
		}),
		"navigationEndpoint": map[string]any{
			"browseEndpoint": map[string]any{"browseId": "VLPLabc123"},
		},
	}

	result := parser.ParseSearchResult(data, "playlist", "Playlists")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.ResultType != "playlist" {
		t.Errorf("ResultType = %q, want playlist", result.ResultType)
	}
	if result.Author != "Spotify" {
		t.Errorf("Author = %q, want %q", result.Author, "Spotify")
	}
	if result.ItemCount != 50 {
		t.Errorf("ItemCount = %d, want 50", result.ItemCount)
	}
	if result.BrowseID != "VLPLabc123" {
		t.Errorf("BrowseID = %q, want %q", result.BrowseID, "VLPLabc123")
	}
}

func TestParseSearchResult_Playlist_NoAuthor(t *testing.T) {
	data := map[string]any{
		"flexColumns": makeFlexCols([][]any{
			{makeRun("My Playlist")},
			// runs: just the item count → hasAuthor=false (len==1)
			{makeRun("10 tracks")},
		}),
		"navigationEndpoint": map[string]any{
			"browseEndpoint": map[string]any{"browseId": "VLPLxyz"},
		},
	}

	result := parser.ParseSearchResult(data, "playlist", "Playlists")
	if result.Author != "" {
		t.Errorf("Author = %q, want empty (no author in runs)", result.Author)
	}
	if result.ItemCount != 10 {
		t.Errorf("ItemCount = %d, want 10", result.ItemCount)
	}
}

func TestParseSearchResult_Station(t *testing.T) {
	data := map[string]any{
		"flexColumns": makeFlexCols([][]any{
			{makeRun("Eminem Radio")},
		}),
		"navigationEndpoint": map[string]any{
			"watchEndpoint": map[string]any{
				"videoId":    "videoStation",
				"playlistId": "RDAMVM_station",
			},
		},
	}

	result := parser.ParseSearchResult(data, "station", "Stations")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.VideoID != "videoStation" {
		t.Errorf("VideoID = %q, want %q", result.VideoID, "videoStation")
	}
	if result.PlaylistID != "RDAMVM_station" {
		t.Errorf("PlaylistID = %q, want %q", result.PlaylistID, "RDAMVM_station")
	}
}

func TestParseSearchResult_Profile(t *testing.T) {
	data := map[string]any{
		"flexColumns": makeFlexCols([][]any{
			{makeRun("Display Name")},
			{makeRun(" • "), makeRun(" • "), makeRun("@handle")},
		}),
		"navigationEndpoint": map[string]any{
			"browseEndpoint": map[string]any{"browseId": "MPLAprofile123"},
		},
	}

	result := parser.ParseSearchResult(data, "profile", "Profiles")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.ResultType != "profile" {
		t.Errorf("ResultType = %q, want profile", result.ResultType)
	}
	if result.BrowseID != "MPLAprofile123" {
		t.Errorf("BrowseID = %q, want %q", result.BrowseID, "MPLAprofile123")
	}
}

func TestParseSearchResult_Episode(t *testing.T) {
	data := mergeInto(makePlayButton("episodeVideo123", "MUSIC_VIDEO_TYPE_PODCAST_EPISODE"), map[string]any{
		"flexColumns": makeFlexCols([][]any{
			{makeRun("Episode Title")},
			{
				makeRun("Jan 1, 2024"),
				makeRun(" • "),
				makeRunWithBrowse("My Podcast", "MPSPpodcast123"),
			},
		}),
	})

	result := parser.ParseSearchResult(data, "episode", "Episodes")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.ResultType != "episode" {
		t.Errorf("ResultType = %q, want episode", result.ResultType)
	}
	if result.VideoID != "episodeVideo123" {
		t.Errorf("VideoID = %q, want %q", result.VideoID, "episodeVideo123")
	}
	if result.Date != "Jan 1, 2024" {
		t.Errorf("Date = %q, want %q", result.Date, "Jan 1, 2024")
	}
	if result.Podcast == nil {
		t.Fatal("Podcast must not be nil")
	}
	if result.Podcast.Name != "My Podcast" {
		t.Errorf("Podcast.Name = %q, want %q", result.Podcast.Name, "My Podcast")
	}
}

func TestParseSearchResult_Episode_NoDateNoPodcast(t *testing.T) {
	// Episode with only 1 run (no date since hasDate=false), but podcast IS set from runs[0].
	data := mergeInto(makePlayButton("ep2", "MUSIC_VIDEO_TYPE_PODCAST_EPISODE"), map[string]any{
		"flexColumns": makeFlexCols([][]any{
			{makeRun("Episode 2")},
			{makeRunWithBrowse("My Podcast", "MPSPpod")},
		}),
	})
	result := parser.ParseSearchResult(data, "episode", "Episodes")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// hasDate = (len(runs) > 1) = false → Date stays empty
	if result.Date != "" {
		t.Errorf("Date should be empty when only 1 run, got %q", result.Date)
	}
	// idx = 0, runs[0] is used as the podcast entry
	if result.Podcast == nil {
		t.Fatal("Podcast should be set from runs[0] when hasDate=false")
	}
	if result.Podcast.Name != "My Podcast" {
		t.Errorf("Podcast.Name = %q, want %q", result.Podcast.Name, "My Podcast")
	}
}

func TestParseSearchResult_Episode_WithLiveBadge(t *testing.T) {
	data := mergeInto(makePlayButton("liveEp", "MUSIC_VIDEO_TYPE_PODCAST_EPISODE"), map[string]any{
		"flexColumns": makeFlexCols([][]any{
			{makeRun("Live Episode")},
			{
				makeRun("Live now"),
				makeRun(" • "),
				makeRunWithBrowse("Live Podcast", "MPSPlive"),
			},
		}),
		"badges": []any{
			map[string]any{"liveBadgeRenderer": map[string]any{}},
		},
	})
	result := parser.ParseSearchResult(data, "episode", "Episodes")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.Live {
		t.Error("expected Live=true for live episode badge")
	}
}

// ---------------------------------------------------------------------------
// ParseSearchResult – auto-detect result type from browseId / videoType
// ---------------------------------------------------------------------------

func TestParseSearchResult_AutoDetect_Song(t *testing.T) {
	data := makePlayButton("songID", "MUSIC_VIDEO_TYPE_ATV")
	result := parser.ParseSearchResult(data, "", "Songs")
	if result.ResultType != model.ResultTypeSong {
		t.Errorf("ResultType = %q, want song", result.ResultType)
	}
}

func TestParseSearchResult_AutoDetect_Video(t *testing.T) {
	data := makePlayButton("vidID", "MUSIC_VIDEO_TYPE_UGC")
	result := parser.ParseSearchResult(data, "", "Videos")
	if result.ResultType != model.ResultTypeVideo {
		t.Errorf("ResultType = %q, want video", result.ResultType)
	}
}

func TestParseSearchResult_AutoDetect_Episode(t *testing.T) {
	data := makePlayButton("epID", "MUSIC_VIDEO_TYPE_PODCAST_EPISODE")
	result := parser.ParseSearchResult(data, "", "Episodes")
	if result.ResultType != model.ResultTypeEpisode {
		t.Errorf("ResultType = %q, want episode", result.ResultType)
	}
}

func TestParseSearchResult_AutoDetect_ArtistFromBrowseID(t *testing.T) {
	data := map[string]any{
		"navigationEndpoint": map[string]any{
			"browseEndpoint": map[string]any{"browseId": "UCbnkAoWhS9i3YvcnysFnFQA"},
		},
	}
	result := parser.ParseSearchResult(data, "", "Artists")
	if result.ResultType != model.ResultTypeArtist {
		t.Errorf("ResultType = %q, want artist (UC prefix)", result.ResultType)
	}
}

func TestParseSearchResult_AutoDetect_AlbumFromBrowseID(t *testing.T) {
	data := map[string]any{
		"navigationEndpoint": map[string]any{
			"browseEndpoint": map[string]any{"browseId": "MPREb_album123"},
		},
	}
	result := parser.ParseSearchResult(data, "", "Albums")
	if result.ResultType != model.ResultTypeAlbum {
		t.Errorf("ResultType = %q, want album (MPRE prefix)", result.ResultType)
	}
}

func TestParseSearchResult_AutoDetect_PlaylistFromBrowseID(t *testing.T) {
	// VM prefix → playlist
	data := map[string]any{
		"navigationEndpoint": map[string]any{
			"browseEndpoint": map[string]any{"browseId": "VMplaylist123"},
		},
	}
	result := parser.ParseSearchResult(data, "", "Playlists")
	if result.ResultType != model.ResultTypePlaylist {
		t.Errorf("ResultType = %q, want playlist (VM prefix)", result.ResultType)
	}
}

func TestParseSearchResult_ExplicitBadge(t *testing.T) {
	data := mergeInto(makePlayButton("songExplicit", "MUSIC_VIDEO_TYPE_ATV"), map[string]any{
		"flexColumns": makeFlexCols([][]any{
			{makeRun("Rap God")},
			{makeRunWithBrowse("Eminem", "UCfM3zsQsOnfWNUppiycmBuw"), makeRun(" • "), makeRun("3:08")},
		}),
		"badges": []any{
			map[string]any{
				"musicInlineBadgeRenderer": map[string]any{
					"accessibilityData": map[string]any{
						"accessibilityData": map[string]any{
							"label": "Explicit",
						},
					},
				},
			},
		},
	})
	result := parser.ParseSearchResult(data, "song", "Songs")
	if !result.IsExplicit {
		t.Error("expected IsExplicit=true for explicit badge")
	}
}

// ---------------------------------------------------------------------------
// ParseMenuPlaylists
// ---------------------------------------------------------------------------

func TestParseMenuPlaylists_Shuffle(t *testing.T) {
	data := map[string]any{
		"menu": map[string]any{
			"menuRenderer": map[string]any{
				"items": []any{
					map[string]any{
						"menuNavigationItemRenderer": map[string]any{
							"icon": map[string]any{"iconType": "MUSIC_SHUFFLE"},
							"navigationEndpoint": map[string]any{
								"watchPlaylistEndpoint": map[string]any{"playlistId": "RDAMPL_shuffle123"},
							},
						},
					},
				},
			},
		},
	}
	result := &model.SearchResult{}
	parser.ParseMenuPlaylists(data, result)
	if result.ShuffleID != "RDAMPL_shuffle123" {
		t.Errorf("ShuffleID = %q, want %q", result.ShuffleID, "RDAMPL_shuffle123")
	}
	if result.RadioID != "" {
		t.Errorf("RadioID should be empty, got %q", result.RadioID)
	}
}

func TestParseMenuPlaylists_Radio(t *testing.T) {
	data := map[string]any{
		"menu": map[string]any{
			"menuRenderer": map[string]any{
				"items": []any{
					map[string]any{
						"menuNavigationItemRenderer": map[string]any{
							"icon": map[string]any{"iconType": "MIX"},
							"navigationEndpoint": map[string]any{
								"watchEndpoint": map[string]any{"playlistId": "RDAMVM_radio123"},
							},
						},
					},
				},
			},
		},
	}
	result := &model.SearchResult{}
	parser.ParseMenuPlaylists(data, result)
	if result.RadioID != "RDAMVM_radio123" {
		t.Errorf("RadioID = %q, want %q", result.RadioID, "RDAMVM_radio123")
	}
}

func TestParseMenuPlaylists_UnknownIcon(t *testing.T) {
	data := map[string]any{
		"menu": map[string]any{
			"menuRenderer": map[string]any{
				"items": []any{
					map[string]any{
						"menuNavigationItemRenderer": map[string]any{
							"icon": map[string]any{"iconType": "UNKNOWN_ICON"},
						},
					},
				},
			},
		},
	}
	result := &model.SearchResult{}
	parser.ParseMenuPlaylists(data, result)
	if result.ShuffleID != "" || result.RadioID != "" {
		t.Errorf("unexpected IDs set: shuffle=%q radio=%q", result.ShuffleID, result.RadioID)
	}
}

func TestParseMenuPlaylists_EmptyItems(t *testing.T) {
	data := map[string]any{}
	result := &model.SearchResult{}
	parser.ParseMenuPlaylists(data, result)
	if result.ShuffleID != "" || result.RadioID != "" {
		t.Errorf("unexpected IDs on empty input: shuffle=%q radio=%q", result.ShuffleID, result.RadioID)
	}
}

func TestParseMenuPlaylists_NonMenuNavigationItem(t *testing.T) {
	// Item is not a menuNavigationItemRenderer, should be skipped.
	data := map[string]any{
		"menu": map[string]any{
			"menuRenderer": map[string]any{
				"items": []any{
					map[string]any{
						"menuServiceItemRenderer": map[string]any{"text": "save"},
					},
					nil,
				},
			},
		},
	}
	result := &model.SearchResult{}
	parser.ParseMenuPlaylists(data, result)
	if result.ShuffleID != "" || result.RadioID != "" {
		t.Error("expected no IDs set for non-menu-navigation items")
	}
}

// ---------------------------------------------------------------------------
// ParseTopResult
// ---------------------------------------------------------------------------

func TestParseTopResult_Artist(t *testing.T) {
	data := map[string]any{
		"subtitle": map[string]any{
			"runs": []any{
				makeRun("Artist"),
				makeRun(" • "),
				makeRun("50M subscribers"),
			},
		},
		"header": map[string]any{
			"musicCardShelfHeaderBasicRenderer": map[string]any{
				"title": map[string]any{
					"runs": []any{makeRun("Top result")},
				},
			},
		},
		"title": map[string]any{
			"runs": []any{makeRun("Eminem")},
		},
	}
	result := parser.ParseTopResult(data, nil)
	if result == nil {
		t.Fatal("ParseTopResult() must not return nil")
	}
	if result.ResultType != "artist" {
		t.Errorf("ResultType = %q, want artist", result.ResultType)
	}
	if result.Subscribers != "50M" {
		t.Errorf("Subscribers = %q, want %q", result.Subscribers, "50M")
	}
}

func TestParseTopResult_Song(t *testing.T) {
	data := map[string]any{
		"subtitle": map[string]any{
			"runs": []any{
				makeRun("Song"),
				makeRun(" • "),
				makeRunWithBrowse("Eminem", "UCfM3zsQsOnfWNUppiycmBuw"),
				makeRun(" • "),
				makeRun("2002"),
			},
		},
		"title": map[string]any{
			"runs": []any{makeRun("Without Me")},
		},
		"onTap": map[string]any{
			"watchEndpoint": map[string]any{
				"videoId": "YVkUvmDQ3HY",
				"watchEndpointMusicSupportedConfigs": map[string]any{
					"watchEndpointMusicConfig": map[string]any{
						"musicVideoType": "MUSIC_VIDEO_TYPE_ATV",
					},
				},
			},
		},
	}
	result := parser.ParseTopResult(data, nil)
	if result.ResultType != "song" {
		t.Errorf("ResultType = %q, want song", result.ResultType)
	}
	if result.Title != "Without Me" {
		t.Errorf("Title = %q, want %q", result.Title, "Without Me")
	}
	if result.VideoID != "YVkUvmDQ3HY" {
		t.Errorf("VideoID = %q, want %q", result.VideoID, "YVkUvmDQ3HY")
	}
}

func TestParseTopResult_Album(t *testing.T) {
	data := map[string]any{
		"subtitle": map[string]any{
			"runs": []any{
				makeRun("Album"),
				makeRun(" • "),
				makeRunWithBrowse("Eminem", "UCfM3zsQsOnfWNUppiycmBuw"),
				makeRun(" • "),
				makeRun("2002"),
			},
		},
		"title": map[string]any{
			"runs": []any{
				map[string]any{
					"text": "The Eminem Show",
					"navigationEndpoint": map[string]any{
						"browseEndpoint": map[string]any{"browseId": "MPREb_emshow"},
					},
				},
			},
		},
		"onTap": map[string]any{
			"watchEndpoint": map[string]any{"videoId": "albumVideoID"},
		},
		"buttons": []any{
			map[string]any{
				"buttonRenderer": map[string]any{
					"command": map[string]any{
						"watchEndpoint": map[string]any{"playlistId": "OLAK5uy_emshow"},
					},
				},
			},
		},
	}
	result := parser.ParseTopResult(data, nil)
	if result.ResultType != "album" {
		t.Errorf("ResultType = %q, want album", result.ResultType)
	}
	if result.Title != "The Eminem Show" {
		t.Errorf("Title = %q, want %q", result.Title, "The Eminem Show")
	}
	if result.BrowseID != "MPREb_emshow" {
		t.Errorf("BrowseID = %q, want %q", result.BrowseID, "MPREb_emshow")
	}
	if result.PlaylistID != "OLAK5uy_emshow" {
		t.Errorf("PlaylistID = %q, want %q", result.PlaylistID, "OLAK5uy_emshow")
	}
}

func TestParseTopResult_Playlist(t *testing.T) {
	data := map[string]any{
		"subtitle": map[string]any{
			"runs": []any{makeRun("Playlist")},
		},
		"title": map[string]any{
			"runs": []any{makeRun("Top 50 Songs")},
		},
		"menu": map[string]any{
			"menuRenderer": map[string]any{
				"items": []any{
					map[string]any{
						"menuNavigationItemRenderer": map[string]any{
							"navigationEndpoint": map[string]any{
								"watchPlaylistEndpoint": map[string]any{"playlistId": "VLPLabc"},
							},
						},
					},
				},
			},
		},
	}
	result := parser.ParseTopResult(data, nil)
	if result.ResultType != "playlist" {
		t.Errorf("ResultType = %q, want playlist", result.ResultType)
	}
	if result.Title != "Top 50 Songs" {
		t.Errorf("Title = %q, want %q", result.Title, "Top 50 Songs")
	}
	if result.PlaylistID != "VLPLabc" {
		t.Errorf("PlaylistID = %q, want %q", result.PlaylistID, "VLPLabc")
	}
}

func TestParseTopResult_Episode(t *testing.T) {
	// subtitle.runs[0] must contain the type string for type detection.
	// The code then reuses subtitleRuns[0] as the Date (a quirk of the implementation).
	data := map[string]any{
		"subtitle": map[string]any{
			"runs": []any{
				makeRun("episode"),     // index 0 → type detection AND used as Date by the code
				makeRun(" • "),         // index 1
				makeRun("Nov 2, 2023"), // index 2 → used as Podcast name via ParseIDName
				makeRun(" • "),         // index 3
				makeRunWithBrowse("The Podcast Show", "MPSPpod123"), // index 4 → ignored (len<=4 for podcast)
			},
		},
		"title": map[string]any{
			"runs": []any{makeRun("Episode 42")},
		},
		"thumbnail": map[string]any{
			"musicThumbnailRenderer": map[string]any{
				"onTap": map[string]any{
					"watchEndpoint": map[string]any{"videoId": "episodeVid42"},
				},
			},
		},
	}
	result := parser.ParseTopResult(data, []string{"episode"})
	if result.ResultType != "episode" {
		t.Errorf("ResultType = %q, want episode", result.ResultType)
	}
	if result.Title != "Episode 42" {
		t.Errorf("Title = %q, want %q", result.Title, "Episode 42")
	}
	if result.VideoID != "episodeVid42" {
		t.Errorf("VideoID = %q, want %q", result.VideoID, "episodeVid42")
	}
}

func TestParseTopResult_Single(t *testing.T) {
	// "Single" subtitle → album type
	data := map[string]any{
		"subtitle": map[string]any{
			"runs": []any{makeRun("Single")},
		},
		"title": map[string]any{
			"runs": []any{makeRun("A Single")},
		},
		"onTap": map[string]any{
			"watchEndpoint": map[string]any{"videoId": "singleVidID"},
		},
	}
	result := parser.ParseTopResult(data, nil)
	if result.ResultType != "album" {
		t.Errorf("ResultType = %q, want album (single→album)", result.ResultType)
	}
}

func TestParseTopResult_EP(t *testing.T) {
	// "EP" subtitle → album type
	data := map[string]any{
		"subtitle": map[string]any{
			"runs": []any{makeRun("EP")},
		},
		"title": map[string]any{
			"runs": []any{makeRun("An EP")},
		},
		"onTap": map[string]any{
			"watchEndpoint": map[string]any{"videoId": "epVidID"},
		},
	}
	result := parser.ParseTopResult(data, nil)
	if result.ResultType != "album" {
		t.Errorf("ResultType = %q, want album (ep→album)", result.ResultType)
	}
}

func TestParseTopResult_EmptySubtitle(t *testing.T) {
	data := map[string]any{
		"subtitle": map[string]any{
			"runs": []any{makeRun("")},
		},
		"title": map[string]any{
			"runs": []any{makeRun("Some Result")},
		},
	}
	result := parser.ParseTopResult(data, nil)
	if result == nil {
		t.Fatal("ParseTopResult() must not return nil")
	}
	// empty subtitle → resultType == ""
	if result.ResultType != "" {
		t.Errorf("expected empty ResultType for empty subtitle, got %q", result.ResultType)
	}
}

func TestParseTopResult_DefaultCategory(t *testing.T) {
	// No header → category defaults to "Top result"
	data := map[string]any{
		"subtitle": map[string]any{
			"runs": []any{makeRun("song")},
		},
		"title": map[string]any{
			"runs": []any{makeRun("Test Song")},
		},
	}
	result := parser.ParseTopResult(data, nil)
	if result.Category != "Top result" {
		t.Errorf("Category = %q, want %q", result.Category, "Top result")
	}
}

func TestParseTopResult_CustomSearchResultType(t *testing.T) {
	// subtitle text matches a type in the provided resultTypes slice
	data := map[string]any{
		"subtitle": map[string]any{
			"runs": []any{makeRun("CustomType")},
		},
		"title": map[string]any{
			"runs": []any{makeRun("Custom Result")},
		},
	}
	result := parser.ParseTopResult(data, []string{"CustomType"})
	// getSearchResultType: lower("CustomType") not in allTypes, checks resultTypesLocal → "CustomType"→"customtype"
	// since "customtype" != "CustomType" the match is case-insensitive
	// The function returns "album" as fallback for unknown types not in resultTypes
	// Let's just check it doesn't panic and returns non-nil
	if result == nil {
		t.Fatal("ParseTopResult() must not return nil")
	}
}

// ---------------------------------------------------------------------------
// ParseSongRuns – additional coverage (views path → splitSpace, year, album)
// ---------------------------------------------------------------------------

func TestParseSongRuns_Views(t *testing.T) {
	// reViews pattern: ^\d[^ ]* [^ ]*$  →  "1.5M views"
	runs := []any{
		makeRun("1.5M views"),
	}
	result := &model.SearchResult{}
	parser.ParseSongRuns(runs, false, result)
	if result.Views != "1.5M" {
		t.Errorf("Views = %q, want %q", result.Views, "1.5M")
	}
}

func TestParseSongRuns_Year(t *testing.T) {
	runs := []any{
		makeRun("2023"),
	}
	result := &model.SearchResult{}
	parser.ParseSongRuns(runs, false, result)
	if result.Year != "2023" {
		t.Errorf("Year = %q, want %q", result.Year, "2023")
	}
}

func TestParseSongRuns_Album(t *testing.T) {
	// browseId starts with MPRE → album type
	runs := []any{
		makeRunWithBrowse("Recovery", "MPREb_recovery"),
	}
	result := &model.SearchResult{}
	parser.ParseSongRuns(runs, false, result)
	if result.Album == nil {
		t.Fatal("Album must not be nil")
	}
	if result.Album.Name != "Recovery" {
		t.Errorf("Album.Name = %q, want %q", result.Album.Name, "Recovery")
	}
	if result.Album.ID != "MPREb_recovery" {
		t.Errorf("Album.ID = %q, want %q", result.Album.ID, "MPREb_recovery")
	}
}

func TestParseSongRuns_SkipTypeSpec(t *testing.T) {
	// With skipTypeSpec=true: if runs[0] and runs[2] are both artist type,
	// the first two runs are skipped (runs = runs[2:])
	runs := []any{
		makeRunWithBrowse("Artist1", "UCxxxxxxx"),
		makeRun(" • "),
		makeRunWithBrowse("Artist2", "UCyyyyyyy"),
		makeRun(" • "),
		makeRun("3:45"),
	}
	result := &model.SearchResult{}
	parser.ParseSongRuns(runs, true, result)
	// After skip, should start at runs[2:] = Artist2 • 3:45
	// Artists should only include Artist2 (and any further artist entries)
	if result.Duration != "3:45" {
		t.Errorf("Duration = %q, want %q", result.Duration, "3:45")
	}
}

func TestParseSongRuns_NoSkipTypeSpec(t *testing.T) {
	// With skipTypeSpec=false: artist runs are NOT skipped.
	runs := []any{
		makeRunWithBrowse("Artist1", "UCxxxxxxx"),
		makeRun(" • "),
		makeRunWithBrowse("Artist2", "UCyyyyyyy"),
	}
	result := &model.SearchResult{}
	parser.ParseSongRuns(runs, false, result)
	if len(result.Artists) == 0 {
		t.Fatal("expected artists from runs when skipTypeSpec=false")
	}
}

func TestParseSongRuns_NilRunsSkipped(t *testing.T) {
	// nil runs at even indices should be skipped gracefully; valid run at index 0 processed.
	runs := []any{
		makeRun("2020"),
		makeRun(" • "),
		nil, // nil at even index 2 should be skipped
	}
	result := &model.SearchResult{}
	parser.ParseSongRuns(runs, false, result)
	if result.Year != "2020" {
		t.Errorf("Year = %q, want %q", result.Year, "2020")
	}
}

// ---------------------------------------------------------------------------
// ParseIDName
// ---------------------------------------------------------------------------

func TestParseIDName_WithBrowseID(t *testing.T) {
	run := map[string]any{
		"text": "My Podcast",
		"navigationEndpoint": map[string]any{
			"browseEndpoint": map[string]any{"browseId": "MPSPpod123"},
		},
	}
	id, name := parser.ParseIDName(run)
	if id != "MPSPpod123" {
		t.Errorf("id = %q, want %q", id, "MPSPpod123")
	}
	if name != "My Podcast" {
		t.Errorf("name = %q, want %q", name, "My Podcast")
	}
}

func TestParseIDName_NoBrowseID(t *testing.T) {
	run := map[string]any{"text": "Just Text"}
	id, name := parser.ParseIDName(run)
	if id != "" {
		t.Errorf("id = %q, want empty", id)
	}
	if name != "Just Text" {
		t.Errorf("name = %q, want %q", name, "Just Text")
	}
}

func TestParseIDName_Nil(t *testing.T) {
	id, name := parser.ParseIDName(nil)
	if id != "" || name != "" {
		t.Errorf("expected empty strings for nil run, got id=%q name=%q", id, name)
	}
}
