package parser

import (
	"strconv"
	"strings"

	"github.com/JanikSachs/go_ytmusicapi/internal/model"
)

// ParseAlbum maps a raw browse response for an album page to a typed Album struct.
// It supports the 2024 two-column response format.
func ParseAlbum(response map[string]any) *model.Album {
	album := &model.Album{}

	header := NavMap(response, []any{
		"contents", "twoColumnBrowseResultsRenderer",
		"tabs", 0, "tabRenderer", "content",
		"sectionListRenderer", "contents", 0,
		"musicResponsiveHeaderRenderer",
	})
	if header == nil {
		album.Tracks = []model.AlbumTrack{}
		return album
	}

	album.Title = NavStr(header, []any{"title", "runs", 0, "text"})
	album.Type = NavStr(header, []any{"subtitle", "runs", 0, "text"})
	album.Year = NavStr(header, []any{"subtitle", "runs", 2, "text"})
	album.Thumbnails = ParseThumbnails(header)

	desc := NavStr(header, []any{"description", "musicDescriptionShelfRenderer", "description", "runs", 0, "text"})
	album.Description = desc

	// straplineTextOne contains artist runs
	strapRuns := NavList(header, []any{"straplineTextOne", "runs"})
	for i, runAny := range strapRuns {
		if i%2 != 0 {
			continue
		}
		run, _ := runAny.(map[string]any)
		if run == nil {
			continue
		}
		name, _ := run["text"].(string)
		id := NavStr(run, []any{"navigationEndpoint", "browseEndpoint", "browseId"})
		album.Artists = append(album.Artists, model.ArtistRef{Name: name, ID: id})
	}

	// secondSubtitle: "19 songs • 1 hour, 17 minutes" or just duration
	secondSubRuns := NavList(header, []any{"secondSubtitle", "runs"})
	if len(secondSubRuns) > 2 {
		countText, _ := secondSubRuns[0].(map[string]any)
		if countText != nil {
			raw, _ := countText["text"].(string)
			album.TrackCount = parseTrackCount(raw)
		}
		durText, _ := secondSubRuns[2].(map[string]any)
		if durText != nil {
			album.Duration, _ = durText["text"].(string)
		}
	} else if len(secondSubRuns) == 1 {
		durText, _ := secondSubRuns[0].(map[string]any)
		if durText != nil {
			album.Duration, _ = durText["text"].(string)
		}
	}

	// audioPlaylistId from play button
	buttons := NavList(header, []any{"buttons"})
	for _, btnAny := range buttons {
		btn, _ := btnAny.(map[string]any)
		if btn == nil {
			continue
		}
		if pbr := NavMap(btn, []any{"musicPlayButtonRenderer"}); pbr != nil {
			id := NavStr(pbr, []any{"playNavigationEndpoint", "watchEndpoint", "playlistId"})
			if id == "" {
				id = NavStr(pbr, []any{"playNavigationEndpoint", "watchPlaylistEndpoint", "playlistId"})
			}
			if id != "" {
				album.AudioPlaylistID = id
				break
			}
		}
	}

	// tracks from secondary contents
	trackItems := NavList(response, []any{
		"contents", "twoColumnBrowseResultsRenderer",
		"secondaryContents", "sectionListRenderer", "contents", 0,
		"musicShelfRenderer", "contents",
	})
	for _, itemAny := range trackItems {
		item, _ := itemAny.(map[string]any)
		if item == nil {
			continue
		}
		mrlir := NavMap(item, []any{"musicResponsiveListItemRenderer"})
		if mrlir == nil {
			continue
		}
		track := parseAlbumTrack(mrlir)
		if track != nil {
			album.Tracks = append(album.Tracks, *track)
		}
	}
	if album.Tracks == nil {
		album.Tracks = []model.AlbumTrack{}
	}

	return album
}

func parseAlbumTrack(item map[string]any) *model.AlbumTrack {
	track := &model.AlbumTrack{}

	track.Title = NavStr(item, []any{"flexColumns", 0, "musicResponsiveListItemFlexColumnRenderer", "text", "runs", 0, "text"})
	if track.Title == "" {
		return nil
	}

	track.VideoID = NavStr(item, []any{
		"overlay", "musicItemThumbnailOverlayRenderer",
		"content", "musicPlayButtonRenderer",
		"playNavigationEndpoint", "watchEndpoint", "videoId",
	})

	// duration from fixedColumns
	durText := NavStr(item, []any{"fixedColumns", 0, "musicResponsiveListItemFixedColumnRenderer", "text", "runs", 0, "text"})
	track.Duration = durText
	if s := ParseDuration(durText); s != nil {
		track.DurationSec = s
	}

	// track number from index
	idxText := NavStr(item, []any{"index", "runs", 0, "text"})
	if n, err := strconv.Atoi(idxText); err == nil {
		track.TrackNumber = n
	}

	// explicit badge
	label := NavStr(item, pathBadgeLabel)
	track.IsExplicit = strings.EqualFold(label, "Explicit")

	return track
}

// parseTrackCount parses strings like "19 songs" → 19.
func parseTrackCount(s string) int {
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return 0
	}
	n, _ := strconv.Atoi(parts[0])
	return n
}
