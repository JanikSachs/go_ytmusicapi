package parser

import (
	"strings"

	"github.com/JanikSachs/go_ytmusicapi/internal/model"
)

// ParsePlaylist maps a raw browse response for a playlist page to a typed Playlist struct.
// Both owned (musicEditablePlaylistDetailHeaderRenderer) and public
// (musicResponsiveHeaderRenderer) formats are handled.
func ParsePlaylist(playlistID string, response map[string]any) *model.Playlist {
	playlist := &model.Playlist{ID: playlistID, Tracks: []model.PlaylistTrack{}}

	tab0Items := NavList(response, []any{
		"contents", "twoColumnBrowseResultsRenderer",
		"tabs", 0, "tabRenderer", "content",
		"sectionListRenderer", "contents",
	})
	if len(tab0Items) == 0 {
		return playlist
	}

	firstItem, _ := tab0Items[0].(map[string]any)
	if firstItem == nil {
		return playlist
	}

	var header map[string]any

	// owned playlist
	if editable := NavMap(firstItem, []any{"musicEditablePlaylistDetailHeaderRenderer"}); editable != nil {
		header = NavMap(editable, []any{"header", "musicResponsiveHeaderRenderer"})
		playlist.ID = NavStr(editable, []any{"playlistId"})
	}

	// public playlist
	if header == nil {
		header = NavMap(firstItem, []any{"musicResponsiveHeaderRenderer"})
	}

	if header == nil {
		playlist.Tracks = []model.PlaylistTrack{}
		return playlist
	}

	playlist.Title = NavStr(header, []any{"title", "runs", 0, "text"})
	playlist.Thumbnails = ParseThumbnails(header)

	// subtitle: "Playlist • Privacy • Year" or "Playlist • Year"
	subtitleRuns := NavList(header, []any{"subtitle", "runs"})
	for i, runAny := range subtitleRuns {
		run, _ := runAny.(map[string]any)
		if run == nil {
			continue
		}
		text, _ := run["text"].(string)
		switch {
		case i == 2 && isPrivacy(text):
			playlist.Privacy = text
		case i == 4 && playlist.Privacy != "" && isYear(text):
			playlist.Year = text
		case i == 2 && !isPrivacy(text) && isYear(text):
			playlist.Year = text
		}
	}

	// secondSubtitle: "245 tracks • 5+ hours"
	secondSubRuns := NavList(header, []any{"secondSubtitle", "runs"})
	if len(secondSubRuns) > 2 {
		countRun, _ := secondSubRuns[0].(map[string]any)
		if countRun != nil {
			playlist.TrackCount = parseTrackCount(NavStr(countRun, []any{"text"}))
		}
		durRun, _ := secondSubRuns[2].(map[string]any)
		if durRun != nil {
			playlist.Duration, _ = durRun["text"].(string)
		}
	} else if len(secondSubRuns) == 1 {
		durRun, _ := secondSubRuns[0].(map[string]any)
		if durRun != nil {
			playlist.Duration, _ = durRun["text"].(string)
		}
	}

	// straplineTextOne = author
	strapRuns := NavList(header, []any{"straplineTextOne", "runs"})
	if len(strapRuns) > 0 {
		run, _ := strapRuns[0].(map[string]any)
		if run != nil {
			name, _ := run["text"].(string)
			id := NavStr(run, []any{"navigationEndpoint", "browseEndpoint", "browseId"})
			playlist.Author = &model.ArtistRef{Name: name, ID: id}
		}
	}

	// description
	descShelf := NavMap(header, []any{"description", "musicDescriptionShelfRenderer"})
	if descShelf != nil {
		playlist.Description = NavStr(descShelf, []any{"description", "runs", 0, "text"})
	}

	// tracks from secondary contents - handles both musicShelfRenderer and musicPlaylistShelfRenderer
	if tracks := parsePlaylistTracks(response); tracks != nil {
		playlist.Tracks = tracks
	}

	return playlist
}

func parsePlaylistTracks(response map[string]any) []model.PlaylistTrack {
	secondaryContents := NavList(response, []any{
		"contents", "twoColumnBrowseResultsRenderer",
		"secondaryContents", "sectionListRenderer", "contents",
	})

	for _, sectionAny := range secondaryContents {
		section, _ := sectionAny.(map[string]any)
		if section == nil {
			continue
		}

		var items []any
		if shelf := NavMap(section, []any{"musicShelfRenderer"}); shelf != nil {
			items = NavList(shelf, []any{"contents"})
		} else if pShelf := NavMap(section, []any{"musicPlaylistShelfRenderer"}); pShelf != nil {
			items = NavList(pShelf, []any{"contents"})
		}

		if len(items) == 0 {
			continue
		}

		var tracks []model.PlaylistTrack
		for _, itemAny := range items {
			item, _ := itemAny.(map[string]any)
			if item == nil {
				continue
			}
			mrlir := NavMap(item, []any{"musicResponsiveListItemRenderer"})
			if mrlir == nil {
				continue
			}
			track := parsePlaylistTrack(mrlir)
			if track != nil {
				tracks = append(tracks, *track)
			}
		}
		return tracks
	}
	return nil
}

func parsePlaylistTrack(item map[string]any) *model.PlaylistTrack {
	track := &model.PlaylistTrack{}

	track.Title = NavStr(item, []any{"flexColumns", 0, "musicResponsiveListItemFlexColumnRenderer", "text", "runs", 0, "text"})
	if track.Title == "" {
		return nil
	}

	// videoId from overlay play button
	track.VideoID = NavStr(item, []any{
		"overlay", "musicItemThumbnailOverlayRenderer",
		"content", "musicPlayButtonRenderer",
		"playNavigationEndpoint", "watchEndpoint", "videoId",
	})
	// fallback: playlistItemData
	if track.VideoID == "" {
		track.VideoID = NavStr(item, []any{"playlistItemData", "videoId"})
	}

	track.SetVideoID = NavStr(item, []any{"playlistItemData", "playlistSetVideoId"})

	// col1 = artists, col3 = album
	col1 := GetFlexColumnItem(item, 1)
	if col1 != nil {
		runs := NavList(col1, []any{"text", "runs"})
		for i, runAny := range runs {
			if i%2 != 0 {
				continue
			}
			run, _ := runAny.(map[string]any)
			if run == nil {
				continue
			}
			name, _ := run["text"].(string)
			id := NavStr(run, []any{"navigationEndpoint", "browseEndpoint", "browseId"})
			track.Artists = append(track.Artists, model.ArtistRef{Name: name, ID: id})
		}
	}

	col3 := GetFlexColumnItem(item, 3)
	if col3 != nil {
		albumName := NavStr(col3, []any{"text", "runs", 0, "text"})
		albumID := NavStr(col3, []any{"text", "runs", 0, "navigationEndpoint", "browseEndpoint", "browseId"})
		if albumName != "" {
			track.Album = &model.AlbumRef{Name: albumName, ID: albumID}
		}
	}

	// duration from fixedColumns
	durText := NavStr(item, []any{"fixedColumns", 0, "musicResponsiveListItemFixedColumnRenderer", "text", "runs", 0, "text"})
	track.Duration = durText
	if s := ParseDuration(durText); s != nil {
		track.DurationSec = s
	}

	// explicit badge
	label := NavStr(item, pathBadgeLabel)
	track.IsExplicit = strings.EqualFold(label, "Explicit")

	// thumbnails
	track.Thumbnails = ParseThumbnails(item)

	return track
}

func isPrivacy(s string) bool {
	switch strings.ToUpper(s) {
	case "PUBLIC", "PRIVATE", "UNLISTED":
		return true
	}
	return false
}

func isYear(s string) bool {
	if len(s) != 4 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
