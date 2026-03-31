package parser_test

import (
	"testing"

	"github.com/JanikSachs/go_ytmusicapi/internal/parser"
)

func TestParseThumbnails(t *testing.T) {
	data := map[string]any{
		"thumbnail": map[string]any{
			"musicThumbnailRenderer": map[string]any{
				"thumbnail": map[string]any{
					"thumbnails": []any{
						map[string]any{
							"url":    "https://example.com/img.jpg",
							"width":  float64(226),
							"height": float64(226),
						},
					},
				},
			},
		},
	}

	thumbs := parser.ParseThumbnails(data)
	if len(thumbs) != 1 {
		t.Fatalf("expected 1 thumbnail, got %d", len(thumbs))
	}
	if thumbs[0].URL != "https://example.com/img.jpg" {
		t.Errorf("expected URL https://example.com/img.jpg, got %s", thumbs[0].URL)
	}
	if thumbs[0].Width != 226 {
		t.Errorf("expected width 226, got %d", thumbs[0].Width)
	}
	if thumbs[0].Height != 226 {
		t.Errorf("expected height 226, got %d", thumbs[0].Height)
	}
}

func TestParseThumbnailsEmpty(t *testing.T) {
	thumbs := parser.ParseThumbnails(map[string]any{})
	if thumbs != nil {
		t.Errorf("expected nil thumbnails for empty data, got %v", thumbs)
	}
}

func TestParseSearchResult_Song(t *testing.T) {
	data := map[string]any{
		"flexColumns": []any{
			map[string]any{
				"musicResponsiveListItemFlexColumnRenderer": map[string]any{
					"text": map[string]any{
						"runs": []any{
							map[string]any{"text": "Bohemian Rhapsody"},
						},
					},
				},
			},
			map[string]any{
				"musicResponsiveListItemFlexColumnRenderer": map[string]any{
					"text": map[string]any{
						"runs": []any{
							map[string]any{
								"text": "Queen",
								"navigationEndpoint": map[string]any{
									"browseEndpoint": map[string]any{
										"browseId": "UCbnkAoWhS9i3YvcnysFnFQA",
									},
								},
							},
							map[string]any{"text": " • "},
							map[string]any{
								"text": "A Night at the Opera",
								"navigationEndpoint": map[string]any{
									"browseEndpoint": map[string]any{
										"browseId": "MPREb_1234567890",
									},
								},
							},
							map[string]any{"text": " • "},
							map[string]any{"text": "5:55"},
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
								"videoId": "fJ9rUzIMcZQ",
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
	}

	result := parser.ParseSearchResult(data, "song", "Songs")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Title != "Bohemian Rhapsody" {
		t.Errorf("expected title Bohemian Rhapsody, got %s", result.Title)
	}
	if result.ResultType != "song" {
		t.Errorf("expected resultType song, got %s", result.ResultType)
	}
	if result.VideoID != "fJ9rUzIMcZQ" {
		t.Errorf("expected videoId fJ9rUzIMcZQ, got %s", result.VideoID)
	}
	if result.Duration != "5:55" {
		t.Errorf("expected duration 5:55, got %s", result.Duration)
	}
	if result.DurationSec == nil || *result.DurationSec != 355 {
		secs := 0
		if result.DurationSec != nil {
			secs = *result.DurationSec
		}
		t.Errorf("expected durationSec 355, got %d", secs)
	}
}

func TestParseSearchResults(t *testing.T) {
	items := []any{
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
										"videoId": "abc123",
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
	}

	results := parser.ParseSearchResults(items, "song", "Songs")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Title != "Test Song" {
		t.Errorf("expected title Test Song, got %s", results[0].Title)
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected *int
	}{
		{"3:45", intPtr(225)},
		{"1:03:45", intPtr(3825)},
		{"0:30", intPtr(30)},
		{"", nil},
		{"   ", nil},
	}
	for _, tt := range tests {
		got := parser.ParseDuration(tt.input)
		if tt.expected == nil {
			if got != nil {
				t.Errorf("ParseDuration(%q) = %d, want nil", tt.input, *got)
			}
			continue
		}
		if got == nil {
			t.Errorf("ParseDuration(%q) = nil, want %d", tt.input, *tt.expected)
			continue
		}
		if *got != *tt.expected {
			t.Errorf("ParseDuration(%q) = %d, want %d", tt.input, *got, *tt.expected)
		}
	}
}

func intPtr(i int) *int { return &i }
