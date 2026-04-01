package parser_test

import (
	"testing"

	"github.com/JanikSachs/go_ytmusicapi/internal/model"
	"github.com/JanikSachs/go_ytmusicapi/internal/parser"
)

// ---------------------------------------------------------------------------
// Nav – cover the default (unknown key type) branch
// ---------------------------------------------------------------------------

func TestNav_UnknownKeyType(t *testing.T) {
	data := map[string]any{"key": "value"}
	// Pass a float64 key (not string or int) → default case → returns nil
	result := parser.Nav(data, []any{3.14})
	if result != nil {
		t.Errorf("Nav() with float64 key should return nil, got %v", result)
	}
}

func TestNav_NilMid(t *testing.T) {
	// Path goes through a nil value mid-way
	data := map[string]any{"a": nil}
	result := parser.Nav(data, []any{"a", "b"})
	if result != nil {
		t.Errorf("Nav() through nil should return nil, got %v", result)
	}
}

func TestNav_NegativeIndex(t *testing.T) {
	data := []any{"x", "y"}
	result := parser.Nav(data, []any{-1})
	if result != nil {
		t.Errorf("Nav() with negative index should return nil, got %v", result)
	}
}

// ---------------------------------------------------------------------------
// GetFlexColumnItem – cover missing-runs branch
// ---------------------------------------------------------------------------

func TestGetFlexColumnItem_NoRunsKey(t *testing.T) {
	// renderer exists but has no "runs" in "text"
	item := map[string]any{
		"flexColumns": []any{
			map[string]any{
				"musicResponsiveListItemFlexColumnRenderer": map[string]any{
					"text": map[string]any{"notRuns": "something"},
				},
			},
		},
	}
	result := parser.GetFlexColumnItem(item, 0)
	if result != nil {
		t.Errorf("GetFlexColumnItem() with no 'runs' key should return nil, got %v", result)
	}
}

func TestGetFlexColumnItem_OutOfBounds(t *testing.T) {
	item := map[string]any{
		"flexColumns": []any{},
	}
	result := parser.GetFlexColumnItem(item, 0)
	if result != nil {
		t.Errorf("GetFlexColumnItem() out of bounds should return nil, got %v", result)
	}
}

func TestGetFlexColumnItem_NilRenderer(t *testing.T) {
	// Column map present but musicResponsiveListItemFlexColumnRenderer is nil
	item := map[string]any{
		"flexColumns": []any{
			map[string]any{
				"musicResponsiveListItemFlexColumnRenderer": nil,
			},
		},
	}
	result := parser.GetFlexColumnItem(item, 0)
	if result != nil {
		t.Errorf("GetFlexColumnItem() with nil renderer should return nil, got %v", result)
	}
}

// ---------------------------------------------------------------------------
// ParseDuration – cover the too-many-parts (>3 colon segments) path
// ---------------------------------------------------------------------------

func TestParseDuration_TooManyParts(t *testing.T) {
	// 4 segments → idx=0 has multiplier index 3 which is out of range
	got := parser.ParseDuration("1:2:3:4")
	if got != nil {
		t.Errorf("ParseDuration(%q) = %d, want nil", "1:2:3:4", *got)
	}
}

func TestParseDuration_NonNumericPart(t *testing.T) {
	got := parser.ParseDuration("1:xx")
	if got != nil {
		t.Errorf("ParseDuration(%q) should be nil for non-numeric part, got %d", "1:xx", *got)
	}
}

// ---------------------------------------------------------------------------
// isYear / parsePlaylist subtitle scenarios
// ---------------------------------------------------------------------------

func TestParsePlaylist_SubtitleWithPrivacyAndYear(t *testing.T) {
	// Build a playlist response with musicEditablePlaylistDetailHeaderRenderer
	// that has Privacy at subtitle[2] and Year at subtitle[4].
	response := map[string]any{
		"contents": map[string]any{
			"twoColumnBrowseResultsRenderer": map[string]any{
				"tabs": []any{
					map[string]any{
						"tabRenderer": map[string]any{
							"content": map[string]any{
								"sectionListRenderer": map[string]any{
									"contents": []any{
										map[string]any{
											"musicEditablePlaylistDetailHeaderRenderer": map[string]any{
												"playlistId": "PLeditable",
												"header": map[string]any{
													"musicResponsiveHeaderRenderer": map[string]any{
														"title": map[string]any{
															"runs": []any{map[string]any{"text": "My Playlist"}},
														},
														"subtitle": map[string]any{
															"runs": []any{
																map[string]any{"text": "Playlist"},
																map[string]any{"text": " • "},
																map[string]any{"text": "PUBLIC"},  // privacy
																map[string]any{"text": " • "},
																map[string]any{"text": "2022"},    // year
															},
														},
														"secondSubtitle": map[string]any{
															"runs": []any{
																map[string]any{"text": "10 songs"},
																map[string]any{"text": " • "},
																map[string]any{"text": "35 minutes"},
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
				"secondaryContents": map[string]any{
					"sectionListRenderer": map[string]any{
						"contents": []any{},
					},
				},
			},
		},
	}
	playlist := parser.ParsePlaylist("PLeditable", response)
	if playlist.Privacy != "PUBLIC" {
		t.Errorf("Privacy = %q, want PUBLIC", playlist.Privacy)
	}
	if playlist.Year != "2022" {
		t.Errorf("Year = %q, want 2022", playlist.Year)
	}
	if playlist.TrackCount != 10 {
		t.Errorf("TrackCount = %d, want 10", playlist.TrackCount)
	}
}

func TestParsePlaylist_SubtitleYearWithoutPrivacy(t *testing.T) {
	// subtitle[2] is a year (no privacy), year should go into playlist.Year
	response := map[string]any{
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
													"runs": []any{map[string]any{"text": "Public Playlist"}},
												},
												"subtitle": map[string]any{
													"runs": []any{
														map[string]any{"text": "Playlist"},
														map[string]any{"text": " • "},
														map[string]any{"text": "2021"}, // year, not privacy
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
						"contents": []any{},
					},
				},
			},
		},
	}
	playlist := parser.ParsePlaylist("PL_pub", response)
	if playlist.Year != "2021" {
		t.Errorf("Year = %q, want 2021", playlist.Year)
	}
	if playlist.Privacy != "" {
		t.Errorf("Privacy = %q, want empty", playlist.Privacy)
	}
}

func TestParsePlaylist_SubtitleInvalidYear(t *testing.T) {
	// subtitle[2] is a non-digit 4-char string – should not set year
	response := map[string]any{
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
													"runs": []any{map[string]any{"text": "Playlist"}},
												},
												"subtitle": map[string]any{
													"runs": []any{
														map[string]any{"text": "Playlist"},
														map[string]any{"text": " • "},
														map[string]any{"text": "abcd"}, // 4 chars but not digits
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
						"contents": []any{},
					},
				},
			},
		},
	}
	playlist := parser.ParsePlaylist("PL_no_year", response)
	if playlist.Year != "" {
		t.Errorf("Year = %q, want empty for non-digit 4-char string", playlist.Year)
	}
}

func TestParsePlaylist_SecondaryContents_PlaylistShelf(t *testing.T) {
	// Uses musicPlaylistShelfRenderer instead of musicShelfRenderer
	trackItem := map[string]any{
		"musicResponsiveListItemRenderer": map[string]any{
			"flexColumns": []any{
				map[string]any{
					"musicResponsiveListItemFlexColumnRenderer": map[string]any{
						"text": map[string]any{
							"runs": []any{map[string]any{"text": "Shelf Track"}},
						},
					},
				},
			},
			"overlay": map[string]any{
				"musicItemThumbnailOverlayRenderer": map[string]any{
					"content": map[string]any{
						"musicPlayButtonRenderer": map[string]any{
							"playNavigationEndpoint": map[string]any{
								"watchEndpoint": map[string]any{"videoId": "shelfVid"},
							},
						},
					},
				},
			},
		},
	}
	response := map[string]any{
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
													"runs": []any{map[string]any{"text": "PShelf Playlist"}},
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
									"contents": []any{trackItem},
								},
							},
						},
					},
				},
			},
		},
	}
	playlist := parser.ParsePlaylist("PL_pshelf", response)
	if len(playlist.Tracks) == 0 {
		t.Fatal("expected tracks from musicPlaylistShelfRenderer")
	}
	if playlist.Tracks[0].Title != "Shelf Track" {
		t.Errorf("Tracks[0].Title = %q, want %q", playlist.Tracks[0].Title, "Shelf Track")
	}
}

func TestParsePlaylist_SecondaryContents_NilSection(t *testing.T) {
	// nil section in secondaryContents should be skipped
	response := map[string]any{
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
													"runs": []any{map[string]any{"text": "Test"}},
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
						"contents": []any{nil},
					},
				},
			},
		},
	}
	playlist := parser.ParsePlaylist("PL_nil_sec", response)
	if len(playlist.Tracks) != 0 {
		t.Errorf("expected no tracks from nil section, got %d", len(playlist.Tracks))
	}
}

func TestParsePlaylist_SecondaryContents_OnlyDuration(t *testing.T) {
	// secondSubtitle has exactly 1 run (just the duration, no track count separator)
	response := map[string]any{
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
													"runs": []any{map[string]any{"text": "Short PL"}},
												},
												"secondSubtitle": map[string]any{
													"runs": []any{
														map[string]any{"text": "1 hour"},
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
						"contents": []any{},
					},
				},
			},
		},
	}
	playlist := parser.ParsePlaylist("PL_dur_only", response)
	if playlist.Duration != "1 hour" {
		t.Errorf("Duration = %q, want %q", playlist.Duration, "1 hour")
	}
}

func TestParsePlaylist_NilHeaderItem(t *testing.T) {
	// firstItem is nil after tab0Items has a nil element
	response := map[string]any{
		"contents": map[string]any{
			"twoColumnBrowseResultsRenderer": map[string]any{
				"tabs": []any{
					map[string]any{
						"tabRenderer": map[string]any{
							"content": map[string]any{
								"sectionListRenderer": map[string]any{
									"contents": []any{nil},
								},
							},
						},
					},
				},
			},
		},
	}
	playlist := parser.ParsePlaylist("PL_nil", response)
	if playlist == nil {
		t.Fatal("ParsePlaylist must not return nil")
	}
}

func TestParsePlaylist_NoHeader(t *testing.T) {
	// firstItem exists but has no recognized header renderer
	response := map[string]any{
		"contents": map[string]any{
			"twoColumnBrowseResultsRenderer": map[string]any{
				"tabs": []any{
					map[string]any{
						"tabRenderer": map[string]any{
							"content": map[string]any{
								"sectionListRenderer": map[string]any{
									"contents": []any{
										map[string]any{"unknownRenderer": map[string]any{}},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	playlist := parser.ParsePlaylist("PL_noheader", response)
	if playlist == nil {
		t.Fatal("ParsePlaylist must not return nil")
	}
	if playlist.Title != "" {
		t.Errorf("Title should be empty for no-header response, got %q", playlist.Title)
	}
}

// ---------------------------------------------------------------------------
// ParseAlbum – cover 1-run secondSubtitle path and playlist-endpoint button
// ---------------------------------------------------------------------------

func TestParseAlbum_SingleDurationRun(t *testing.T) {
	// secondSubtitle has exactly 1 run (just duration, no track count)
	response := map[string]any{
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
													"runs": []any{map[string]any{"text": "Single Track EP"}},
												},
												"subtitle": map[string]any{
													"runs": []any{map[string]any{"text": "EP"}, map[string]any{"text": " • "}, map[string]any{"text": "2020"}},
												},
												"secondSubtitle": map[string]any{
													"runs": []any{
														map[string]any{"text": "3 minutes 20 seconds"},
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
						"contents": []any{},
					},
				},
			},
		},
	}
	album := parser.ParseAlbum(response)
	if album.Title != "Single Track EP" {
		t.Errorf("Title = %q, want %q", album.Title, "Single Track EP")
	}
	if album.Duration != "3 minutes 20 seconds" {
		t.Errorf("Duration = %q, want %q", album.Duration, "3 minutes 20 seconds")
	}
	if album.TrackCount != 0 {
		t.Errorf("TrackCount should be 0 for single-run secondSubtitle, got %d", album.TrackCount)
	}
}

func TestParseAlbum_AudioPlaylistID_WatchPlaylistEndpoint(t *testing.T) {
	// Button uses watchPlaylistEndpoint (not watchEndpoint)
	response := map[string]any{
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
													"runs": []any{map[string]any{"text": "Watch Playlist Album"}},
												},
												"buttons": []any{
													nil, // nil button should be skipped
													map[string]any{
														"musicPlayButtonRenderer": map[string]any{
															"playNavigationEndpoint": map[string]any{
																"watchPlaylistEndpoint": map[string]any{"playlistId": "OLAK5uy_watchPL"},
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
				"secondaryContents": map[string]any{
					"sectionListRenderer": map[string]any{
						"contents": []any{},
					},
				},
			},
		},
	}
	album := parser.ParseAlbum(response)
	if album.AudioPlaylistID != "OLAK5uy_watchPL" {
		t.Errorf("AudioPlaylistID = %q, want %q", album.AudioPlaylistID, "OLAK5uy_watchPL")
	}
}

func TestParseAlbum_TrackCountEmpty(t *testing.T) {
	// secondSubtitle[0] text is empty string → parseTrackCount returns 0
	response := map[string]any{
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
													"runs": []any{map[string]any{"text": "Empty Count"}},
												},
												"secondSubtitle": map[string]any{
													"runs": []any{
														map[string]any{"text": ""},
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
						"contents": []any{},
					},
				},
			},
		},
	}
	album := parser.ParseAlbum(response)
	if album.TrackCount != 0 {
		t.Errorf("TrackCount = %d, want 0 for empty count text", album.TrackCount)
	}
}

// ---------------------------------------------------------------------------
// parseSongMenuData – cover the loop body (toggle item with LIBRARY_ADD/SAVED)
// ---------------------------------------------------------------------------

func TestParseSearchResult_Song_WithMenuToggle(t *testing.T) {
	// Song result with a toggleMenuServiceItemRenderer having LIBRARY_ADD icon
	data := mergeInto(makePlayButton("songLib", "MUSIC_VIDEO_TYPE_ATV"), map[string]any{
		"flexColumns": makeFlexCols([][]any{
			{makeRun("Library Song")},
			{makeRunWithBrowse("Artist", "UCabc"), makeRun(" • "), makeRun("3:20")},
		}),
		"menu": map[string]any{
			"menuRenderer": map[string]any{
				"items": []any{
					// nil item → should be skipped
					nil,
					// non-toggle item → should be skipped
					map[string]any{
						"menuNavigationItemRenderer": map[string]any{},
					},
					// toggle item with LIBRARY_ADD
					map[string]any{
						"toggleMenuServiceItemRenderer": map[string]any{
							"defaultIcon": map[string]any{"iconType": "LIBRARY_ADD"},
						},
					},
					// toggle item with LIBRARY_SAVED
					map[string]any{
						"toggleMenuServiceItemRenderer": map[string]any{
							"defaultIcon": map[string]any{"iconType": "LIBRARY_SAVED"},
						},
					},
					// toggle item with unrecognized icon
					map[string]any{
						"toggleMenuServiceItemRenderer": map[string]any{
							"defaultIcon": map[string]any{"iconType": "UNKNOWN"},
						},
					},
				},
			},
		},
	})

	result := parser.ParseSearchResult(data, "song", "Songs")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Title != "Library Song" {
		t.Errorf("Title = %q, want %q", result.Title, "Library Song")
	}
}

// ---------------------------------------------------------------------------
// ParseSearchResults – nil item in list should be skipped
// ---------------------------------------------------------------------------

func TestParseSearchResults_NilItemSkipped(t *testing.T) {
	items := []any{
		nil,
		map[string]any{
			"musicResponsiveListItemRenderer": map[string]any{
				"flexColumns": makeFlexCols([][]any{
					{makeRun("Valid Song")},
				}),
			},
		},
		// item without musicResponsiveListItemRenderer key → skipped
		map[string]any{"someOtherKey": "value"},
	}
	results := parser.ParseSearchResults(items, "song", "Songs")
	if len(results) != 1 {
		t.Fatalf("expected 1 result (nil and no-renderer items skipped), got %d", len(results))
	}
	if results[0].Title != "Valid Song" {
		t.Errorf("results[0].Title = %q, want %q", results[0].Title, "Valid Song")
	}
}

// ---------------------------------------------------------------------------
// ParsePlaylistTrack – cover video-id fallback and SetVideoID
// ---------------------------------------------------------------------------

func TestParsePlaylistTrack_VideoIDFallback(t *testing.T) {
	// No overlay play button, videoId comes from playlistItemData
	trackItem := map[string]any{
		"musicResponsiveListItemRenderer": map[string]any{
			"flexColumns": []any{
				map[string]any{
					"musicResponsiveListItemFlexColumnRenderer": map[string]any{
						"text": map[string]any{
							"runs": []any{map[string]any{"text": "Fallback Track"}},
						},
					},
				},
			},
			"playlistItemData": map[string]any{
				"videoId":          "fallbackVid123",
				"playlistSetVideoId": "setVid456",
			},
			"fixedColumns": []any{
				map[string]any{
					"musicResponsiveListItemFixedColumnRenderer": map[string]any{
						"text": map[string]any{
							"runs": []any{map[string]any{"text": "4:00"}},
						},
					},
				},
			},
		},
	}
	response := map[string]any{
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
													"runs": []any{map[string]any{"text": "Test PL"}},
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
									"contents": []any{trackItem},
								},
							},
						},
					},
				},
			},
		},
	}
	playlist := parser.ParsePlaylist("PL_fallback", response)
	if len(playlist.Tracks) == 0 {
		t.Fatal("expected tracks")
	}
	track := playlist.Tracks[0]
	if track.VideoID != "fallbackVid123" {
		t.Errorf("VideoID = %q, want fallbackVid123", track.VideoID)
	}
	if track.SetVideoID != "setVid456" {
		t.Errorf("SetVideoID = %q, want setVid456", track.SetVideoID)
	}
	if track.Duration != "4:00" {
		t.Errorf("Duration = %q, want 4:00", track.Duration)
	}
	if track.DurationSec == nil || *track.DurationSec != 240 {
		t.Errorf("DurationSec should be 240, got %v", track.DurationSec)
	}
}

// ---------------------------------------------------------------------------
// getSearchResultType – cover the resultTypesLocal match branch
// ---------------------------------------------------------------------------

func TestParseTopResult_MatchesResultTypesList(t *testing.T) {
	// subtitle "mytype" is in resultTypesLocal but not in allTypes
	data := map[string]any{
		"subtitle": map[string]any{
			"runs": []any{makeRun("mytype")},
		},
		"title": map[string]any{
			"runs": []any{makeRun("Custom Result")},
		},
	}
	// "mytype" is not in allTypes but IS in resultTypesLocal → matched (case-insensitive)
	result := parser.ParseTopResult(data, []string{"song", "mytype"})
	// getSearchResultType returns "mytype" (from resultTypesLocal)
	if result == nil {
		t.Fatal("ParseTopResult() must not return nil")
	}
	if string(result.ResultType) != "mytype" {
		t.Errorf("ResultType = %q, want %q", result.ResultType, "mytype")
	}
}

// ---------------------------------------------------------------------------
// ParseThumbnails – cover invalid (non-map) thumbnail item
// ---------------------------------------------------------------------------

func TestParseThumbnails_InvalidItem(t *testing.T) {
	// The thumbnails list contains a non-map item → should be skipped
	data := map[string]any{
		"thumbnail": map[string]any{
			"musicThumbnailRenderer": map[string]any{
				"thumbnail": map[string]any{
					"thumbnails": []any{
						"not-a-map",
						map[string]any{"url": "https://example.com/valid.jpg"},
					},
				},
			},
		},
	}
	thumbs := parser.ParseThumbnails(data)
	if len(thumbs) != 1 {
		t.Fatalf("expected 1 thumbnail (non-map skipped), got %d", len(thumbs))
	}
	if thumbs[0].URL != "https://example.com/valid.jpg" {
		t.Errorf("thumbs[0].URL = %q, want %q", thumbs[0].URL, "https://example.com/valid.jpg")
	}
}

// ---------------------------------------------------------------------------
// GetItemText – cover run-out-of-bounds branch
// ---------------------------------------------------------------------------

func TestGetItemText_RunOutOfBounds(t *testing.T) {
	item := map[string]any{
		"flexColumns": []any{
			map[string]any{
				"musicResponsiveListItemFlexColumnRenderer": map[string]any{
					"text": map[string]any{
						"runs": []any{map[string]any{"text": "Only one run"}},
					},
				},
			},
		},
	}
	// runIndex=5 is out of bounds → should return ""
	result := parser.GetItemText(item, 0, 5)
	if result != "" {
		t.Errorf("GetItemText() out-of-bounds runIndex = %q, want empty", result)
	}
}

// ---------------------------------------------------------------------------
// ParsePlaylistTrack – nil track title → track skipped
// ---------------------------------------------------------------------------

func TestParsePlaylistTrack_NilTitleSkipped(t *testing.T) {
	// musicResponsiveListItemRenderer with empty title → track is nil and skipped
	trackItem := map[string]any{
		"musicResponsiveListItemRenderer": map[string]any{
			"flexColumns": []any{
				map[string]any{
					"musicResponsiveListItemFlexColumnRenderer": map[string]any{
						"text": map[string]any{
							"runs": []any{map[string]any{"text": ""}},
						},
					},
				},
			},
		},
	}
	response := map[string]any{
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
													"runs": []any{map[string]any{"text": "Test"}},
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
									"contents": []any{
										trackItem,
										nil, // nil item in track list → skipped
										map[string]any{"noRenderer": "skipMe"}, // no musicResponsiveListItemRenderer
									},
								},
							},
						},
					},
				},
			},
		},
	}
	playlist := parser.ParsePlaylist("PL_nil_track", response)
	if len(playlist.Tracks) != 0 {
		t.Errorf("expected 0 tracks (empty title skipped), got %d", len(playlist.Tracks))
	}
}

// ---------------------------------------------------------------------------
// ExtractShelfContinuation – nil first continuation item
// ---------------------------------------------------------------------------

func TestExtractShelfContinuation_NilFirstItem(t *testing.T) {
	shelf := map[string]any{
		"continuations": []any{nil},
	}
	got := parser.ExtractShelfContinuation(shelf)
	if got != "" {
		t.Errorf("expected empty token for nil first item, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// model: SearchResult type constant coverage check
// ---------------------------------------------------------------------------

func TestSearchResultTypeConstants(t *testing.T) {
	types := []model.SearchResultType{
		model.ResultTypeSong, model.ResultTypeVideo, model.ResultTypeAlbum,
		model.ResultTypeArtist, model.ResultTypePlaylist, model.ResultTypeStation,
		model.ResultTypeProfile, model.ResultTypePodcast, model.ResultTypeEpisode,
	}
	for _, rt := range types {
		if string(rt) == "" {
			t.Errorf("SearchResultType constant should not be empty")
		}
	}
}
