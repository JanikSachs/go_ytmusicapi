package parser_test

import (
	"testing"

	"github.com/JanikSachs/go_ytmusicapi/internal/parser"
)

// buildArtistResponse creates a complete synthetic artist browse response
// covering all code-paths in ParseArtist and its helpers.
func buildArtistResponse() map[string]any {
	thumbnail := func(url string) map[string]any {
		return map[string]any{
			"thumbnail": map[string]any{
				"musicThumbnailRenderer": map[string]any{
					"thumbnail": map[string]any{
						"thumbnails": []any{
							map[string]any{"url": url, "width": float64(226), "height": float64(226)},
						},
					},
				},
			},
		}
	}

	mergeMap := func(base, extra map[string]any) map[string]any {
		for k, v := range extra {
			base[k] = v
		}
		return base
	}

	// song item inside the musicShelfRenderer
	songItem := map[string]any{
		"musicResponsiveListItemRenderer": mergeMap(
			map[string]any{
				"flexColumns": []any{
					map[string]any{
						"musicResponsiveListItemFlexColumnRenderer": map[string]any{
							"text": map[string]any{
								"runs": []any{map[string]any{"text": "Lose Yourself"}},
							},
						},
					},
					map[string]any{
						"musicResponsiveListItemFlexColumnRenderer": map[string]any{
							"text": map[string]any{
								"runs": []any{
									map[string]any{
										"text": "Eminem",
										"navigationEndpoint": map[string]any{
											"browseEndpoint": map[string]any{
												"browseId": "UCfM3zsQsOnfWNUppiycmBuw",
											},
										},
									},
								},
							},
						},
					},
					map[string]any{
						"musicResponsiveListItemFlexColumnRenderer": map[string]any{
							"text": map[string]any{
								"runs": []any{map[string]any{"text": "8 Mile"}},
							},
						},
					},
				},
				"overlay": map[string]any{
					"musicItemThumbnailOverlayRenderer": map[string]any{
						"content": map[string]any{
							"musicPlayButtonRenderer": map[string]any{
								"playNavigationEndpoint": map[string]any{
									"watchEndpoint": map[string]any{"videoId": "xFYQQPAOz7Y"},
								},
							},
						},
					},
				},
			},
			thumbnail("https://example.com/lose_yourself.jpg"),
		),
	}

	// song item missing title (should be skipped)
	emptyTitleSongItem := map[string]any{
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

	// songs shelf
	songsShelf := map[string]any{
		"musicShelfRenderer": map[string]any{
			"title": map[string]any{
				"runs": []any{
					map[string]any{
						"text": "Songs",
						"navigationEndpoint": map[string]any{
							"browseEndpoint": map[string]any{
								"browseId": "UCfM3zsQsOnfWNUppiycmBuw",
							},
						},
					},
				},
			},
			"contents": []any{songItem, emptyTitleSongItem, nil},
		},
	}

	// helper to make a musicTwoRowItemRenderer entry for albums/singles/videos
	makeTwoRow := func(title, browseID, year, thumbURL string) map[string]any {
		return map[string]any{
			"musicTwoRowItemRenderer": mergeMap(
				map[string]any{
					"title": map[string]any{
						"runs": []any{
							map[string]any{
								"text": title,
								"navigationEndpoint": map[string]any{
									"browseEndpoint": map[string]any{"browseId": browseID},
								},
							},
						},
					},
					"subtitle": map[string]any{
						"runs": []any{map[string]any{"text": year}},
					},
				},
				thumbnail(thumbURL),
			),
		}
	}

	makeRelated := func(title, browseID, subscribers, thumbURL string) map[string]any {
		return map[string]any{
			"musicTwoRowItemRenderer": mergeMap(
				map[string]any{
					"title": map[string]any{
						"runs": []any{map[string]any{"text": title}},
					},
					"navigationEndpoint": map[string]any{
						"browseEndpoint": map[string]any{"browseId": browseID},
					},
					"subtitle": map[string]any{
						"runs": []any{map[string]any{"text": subscribers}},
					},
				},
				thumbnail(thumbURL),
			),
		}
	}

	makeCarousel := func(title, browseID, params string, contents []any) map[string]any {
		return map[string]any{
			"musicCarouselShelfRenderer": map[string]any{
				"header": map[string]any{
					"musicCarouselShelfBasicHeaderRenderer": map[string]any{
						"title": map[string]any{
							"runs": []any{
								map[string]any{
									"text": title,
									"navigationEndpoint": map[string]any{
										"browseEndpoint": map[string]any{
											"browseId": browseID,
											"params":   params,
										},
									},
								},
							},
						},
					},
				},
				"contents": contents,
			},
		}
	}

	albumsCarousel := makeCarousel("Albums", "UCfM3zsQsOnfWNUppiycmBuw", "6gPTAUNwc0BH", []any{
		makeTwoRow("Recovery", "MPREb_recovery", "2010", "https://example.com/recovery.jpg"),
		nil, // nil should be skipped
		map[string]any{"unknownKey": "ignored"}, // no musicTwoRowItemRenderer, should be skipped
	})

	singlesCarousel := makeCarousel("Singles", "UCfM3zsQsOnfWNUppiycmBuw", "6gPTAUNwcSng", []any{
		makeTwoRow("Guts Over Fear", "MPREb_guts", "2014", "https://example.com/guts.jpg"),
	})

	videosCarousel := makeCarousel("Videos", "UCfM3zsQsOnfWNUppiycmBuw", "6gPTAUNwcVid", []any{
		makeTwoRow("Not Afraid (Video)", "MPREb_notafraid", "2010", "https://example.com/notafraid.jpg"),
	})

	relatedCarousel := makeCarousel("Related artists", "", "", []any{
		makeRelated("D12", "UCxxxxD12", "1.2M subscribers", "https://example.com/d12.jpg"),
		nil, // nil should be skipped
	})

	// Carousel with unknown title should be ignored
	unknownCarousel := map[string]any{
		"musicCarouselShelfRenderer": map[string]any{
			"header": map[string]any{
				"musicCarouselShelfBasicHeaderRenderer": map[string]any{
					"title": map[string]any{
						"runs": []any{map[string]any{"text": "Unknown Section"}},
					},
				},
			},
			"contents": []any{},
		},
	}

	// Carousel with missing header should be ignored
	noHeaderCarousel := map[string]any{
		"musicCarouselShelfRenderer": map[string]any{
			"contents": []any{},
		},
	}

	descShelf := map[string]any{
		"musicDescriptionShelfRenderer": map[string]any{
			"description": map[string]any{
				"runs": []any{map[string]any{"text": "Eminem is a rapper from Detroit."}},
			},
			"subheader": map[string]any{
				"runs": []any{map[string]any{"text": "5.2B views"}},
			},
		},
	}

	headerData := mergeMap(
		map[string]any{
			"title": map[string]any{
				"runs": []any{map[string]any{"text": "Eminem"}},
			},
			"subscriptionButton": map[string]any{
				"subscribeButtonRenderer": map[string]any{
					"channelId": "UCfM3zsQsOnfWNUppiycmBuw",
					"subscriberCountText": map[string]any{
						"runs": []any{map[string]any{"text": "50M subscribers"}},
					},
				},
			},
			"playButton": map[string]any{
				"buttonRenderer": map[string]any{
					"navigationEndpoint": map[string]any{
						"watchEndpoint": map[string]any{"playlistId": "RDAMEMkfjdjdkd"},
					},
				},
			},
			"startRadioButton": map[string]any{
				"buttonRenderer": map[string]any{
					"navigationEndpoint": map[string]any{
						"watchEndpoint": map[string]any{"playlistId": "RDAMPMkfjdjdkd"},
					},
				},
			},
		},
		thumbnail("https://example.com/eminem.jpg"),
	)

	return map[string]any{
		"header": map[string]any{
			"musicImmersiveHeaderRenderer": headerData,
		},
		"contents": map[string]any{
			"singleColumnBrowseResultsRenderer": map[string]any{
				"tabs": []any{
					map[string]any{
						"tabRenderer": map[string]any{
							"content": map[string]any{
								"sectionListRenderer": map[string]any{
									"contents": []any{
										descShelf,
										songsShelf,
										albumsCarousel,
										singlesCarousel,
										videosCarousel,
										relatedCarousel,
										unknownCarousel,
										noHeaderCarousel,
										nil, // nil sections should be skipped
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

func TestParseArtist_Full(t *testing.T) {
	data := buildArtistResponse()
	artist := parser.ParseArtist(data)

	if artist == nil {
		t.Fatal("ParseArtist() must not return nil")
	}
	if artist.Name != "Eminem" {
		t.Errorf("Name = %q, want %q", artist.Name, "Eminem")
	}
	if artist.ChannelID != "UCfM3zsQsOnfWNUppiycmBuw" {
		t.Errorf("ChannelID = %q, want %q", artist.ChannelID, "UCfM3zsQsOnfWNUppiycmBuw")
	}
	if artist.Subscribers != "50M subscribers" {
		t.Errorf("Subscribers = %q, want %q", artist.Subscribers, "50M subscribers")
	}
	if artist.ShuffleID != "RDAMEMkfjdjdkd" {
		t.Errorf("ShuffleID = %q, want %q", artist.ShuffleID, "RDAMEMkfjdjdkd")
	}
	if artist.RadioID != "RDAMPMkfjdjdkd" {
		t.Errorf("RadioID = %q, want %q", artist.RadioID, "RDAMPMkfjdjdkd")
	}
	if len(artist.Thumbnails) == 0 {
		t.Error("Thumbnails must not be empty")
	}
	if artist.Description != "Eminem is a rapper from Detroit." {
		t.Errorf("Description = %q, want %q", artist.Description, "Eminem is a rapper from Detroit.")
	}
	if artist.Views != "5.2B views" {
		t.Errorf("Views = %q, want %q", artist.Views, "5.2B views")
	}
}

func TestParseArtist_Songs(t *testing.T) {
	data := buildArtistResponse()
	artist := parser.ParseArtist(data)

	if artist.Songs == nil {
		t.Fatal("Songs section must not be nil")
	}
	if artist.Songs.BrowseID != "UCfM3zsQsOnfWNUppiycmBuw" {
		t.Errorf("Songs.BrowseID = %q, want %q", artist.Songs.BrowseID, "UCfM3zsQsOnfWNUppiycmBuw")
	}
	if len(artist.Songs.Results) == 0 {
		t.Fatal("Songs.Results must not be empty")
	}
	song := artist.Songs.Results[0]
	if song.Title != "Lose Yourself" {
		t.Errorf("Songs[0].Title = %q, want %q", song.Title, "Lose Yourself")
	}
	if song.VideoID != "xFYQQPAOz7Y" {
		t.Errorf("Songs[0].VideoID = %q, want %q", song.VideoID, "xFYQQPAOz7Y")
	}
	if len(song.Artists) == 0 {
		t.Fatal("Songs[0].Artists must not be empty")
	}
	if song.Artists[0].Name != "Eminem" {
		t.Errorf("Songs[0].Artists[0].Name = %q, want %q", song.Artists[0].Name, "Eminem")
	}
	if song.Album != "8 Mile" {
		t.Errorf("Songs[0].Album = %q, want %q", song.Album, "8 Mile")
	}
	if len(song.Thumbnails) == 0 {
		t.Error("Songs[0].Thumbnails must not be empty")
	}
}

func TestParseArtist_Albums(t *testing.T) {
	data := buildArtistResponse()
	artist := parser.ParseArtist(data)

	if artist.Albums == nil {
		t.Fatal("Albums section must not be nil")
	}
	if artist.Albums.BrowseID == "" {
		t.Error("Albums.BrowseID must not be empty")
	}
	if artist.Albums.Params == "" {
		t.Error("Albums.Params must not be empty")
	}
	if len(artist.Albums.Results) == 0 {
		t.Fatal("Albums.Results must not be empty")
	}
	album := artist.Albums.Results[0]
	if album.Title != "Recovery" {
		t.Errorf("Albums[0].Title = %q, want %q", album.Title, "Recovery")
	}
	if album.BrowseID != "MPREb_recovery" {
		t.Errorf("Albums[0].BrowseID = %q, want %q", album.BrowseID, "MPREb_recovery")
	}
	if album.Year != "2010" {
		t.Errorf("Albums[0].Year = %q, want %q", album.Year, "2010")
	}
	if len(album.Thumbnails) == 0 {
		t.Error("Albums[0].Thumbnails must not be empty")
	}
}

func TestParseArtist_Singles(t *testing.T) {
	data := buildArtistResponse()
	artist := parser.ParseArtist(data)

	if artist.Singles == nil {
		t.Fatal("Singles section must not be nil")
	}
	if len(artist.Singles.Results) == 0 {
		t.Fatal("Singles.Results must not be empty")
	}
	if artist.Singles.Results[0].Title != "Guts Over Fear" {
		t.Errorf("Singles[0].Title = %q, want %q", artist.Singles.Results[0].Title, "Guts Over Fear")
	}
}

func TestParseArtist_Videos(t *testing.T) {
	data := buildArtistResponse()
	artist := parser.ParseArtist(data)

	if artist.Videos == nil {
		t.Fatal("Videos section must not be nil")
	}
	if len(artist.Videos.Results) == 0 {
		t.Fatal("Videos.Results must not be empty")
	}
}

func TestParseArtist_Related(t *testing.T) {
	data := buildArtistResponse()
	artist := parser.ParseArtist(data)

	if len(artist.Related) == 0 {
		t.Fatal("Related must not be empty")
	}
	related := artist.Related[0]
	if related.Title != "D12" {
		t.Errorf("Related[0].Title = %q, want %q", related.Title, "D12")
	}
	if related.BrowseID != "UCxxxxD12" {
		t.Errorf("Related[0].BrowseID = %q, want %q", related.BrowseID, "UCxxxxD12")
	}
	if related.Subscribers != "1.2M subscribers" {
		t.Errorf("Related[0].Subscribers = %q, want %q", related.Subscribers, "1.2M subscribers")
	}
	if len(related.Thumbnails) == 0 {
		t.Error("Related[0].Thumbnails must not be empty")
	}
}

func TestParseArtist_NoHeader(t *testing.T) {
	artist := parser.ParseArtist(map[string]any{})
	if artist == nil {
		t.Fatal("ParseArtist must not return nil")
	}
	if artist.Name != "" {
		t.Errorf("expected empty name, got %q", artist.Name)
	}
}

func TestParseArtist_NoSubscriptionButton(t *testing.T) {
	data := map[string]any{
		"header": map[string]any{
			"musicImmersiveHeaderRenderer": map[string]any{
				"title": map[string]any{
					"runs": []any{map[string]any{"text": "Test Artist"}},
				},
			},
		},
	}
	artist := parser.ParseArtist(data)
	if artist.Name != "Test Artist" {
		t.Errorf("Name = %q, want %q", artist.Name, "Test Artist")
	}
	if artist.ChannelID != "" {
		t.Errorf("ChannelID should be empty when subscriptionButton is missing, got %q", artist.ChannelID)
	}
}
