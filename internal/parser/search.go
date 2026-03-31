package parser

import (
	"strconv"
	"strings"

	"github.com/JanikSachs/go_ytmusicapi/internal/model"
)

var (
	pathThumbnails = []any{"thumbnail", "musicThumbnailRenderer", "thumbnail", "thumbnails"}
	pathBrowseID   = []any{"navigationEndpoint", "browseEndpoint", "browseId"}
	pathMenuItems  = []any{"menu", "menuRenderer", "items"}
	pathPlayButton = []any{"overlay", "musicItemThumbnailOverlayRenderer", "content", "musicPlayButtonRenderer"}
	pathBadgeLabel = []any{"badges", 0, "musicInlineBadgeRenderer", "accessibilityData", "accessibilityData", "label"}

	pathNavigationVideoType = []any{"watchEndpoint", "watchEndpointMusicSupportedConfigs", "watchEndpointMusicConfig", "musicVideoType"}

	pathMRLIR = "musicResponsiveListItemRenderer"
)

// concatPath creates a new []any by concatenating a and b, never mutating a or b.
func concatPath(a []any, b ...any) []any {
	result := make([]any, len(a)+len(b))
	copy(result, a)
	copy(result[len(a):], b)
	return result
}

// concatPaths creates a new []any by concatenating two slices, never mutating either.
func concatPaths(a, b []any) []any {
	result := make([]any, len(a)+len(b))
	copy(result, a)
	copy(result[len(a):], b)
	return result
}

// ParseThumbnails extracts thumbnails from a data item.
func ParseThumbnails(data map[string]any) []model.Thumbnail {
	list := NavList(data, pathThumbnails)
	if list == nil {
		return nil
	}
	var result []model.Thumbnail
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		t := model.Thumbnail{}
		t.URL, _ = m["url"].(string)
		if w, ok := m["width"].(float64); ok {
			t.Width = int(w)
		}
		if h, ok := m["height"].(float64); ok {
			t.Height = int(h)
		}
		result = append(result, t)
	}
	return result
}

// ParseMenuPlaylists fills shuffleId/radioId from the item's context menu.
func ParseMenuPlaylists(data map[string]any, result *model.SearchResult) {
	items := NavList(data, pathMenuItems)
	if items == nil {
		return
	}
	for _, itemAny := range items {
		item, _ := itemAny.(map[string]any)
		if item == nil {
			continue
		}
		mnir, _ := item["menuNavigationItemRenderer"].(map[string]any)
		if mnir == nil {
			continue
		}
		iconType := NavStr(mnir, []any{"icon", "iconType"})
		var watchKey string
		switch iconType {
		case "MUSIC_SHUFFLE":
			watchKey = "shuffle"
		case "MIX":
			watchKey = "radio"
		default:
			continue
		}
		watchID := NavStr(mnir, []any{"navigationEndpoint", "watchPlaylistEndpoint", "playlistId"})
		if watchID == "" {
			watchID = NavStr(mnir, []any{"navigationEndpoint", "watchEndpoint", "playlistId"})
		}
		if watchID != "" {
			if watchKey == "shuffle" {
				result.ShuffleID = watchID
			} else {
				result.RadioID = watchID
			}
		}
	}
}

// ParseSearchResult parses a single musicResponsiveListItemRenderer into a SearchResult.
func ParseSearchResult(data map[string]any, resultType string, category string) *model.SearchResult {
	result := &model.SearchResult{Category: category}

	videoType := NavStr(data, concatPaths(pathPlayButton, pathNavigationVideoType))

	if resultType == "" {
		if browseID := NavStr(data, pathBrowseID); browseID != "" {
			prefixMap := map[string]string{
				"VM": "playlist", "RD": "playlist", "VL": "playlist",
				"MPLA": "artist", "MPRE": "album", "MPSP": "podcast",
				"MPED": "episode", "UC": "artist",
			}
			for prefix, t := range prefixMap {
				if strings.HasPrefix(browseID, prefix) {
					resultType = t
					break
				}
			}
		}
		if resultType == "" {
			switch videoType {
			case "MUSIC_VIDEO_TYPE_ATV":
				resultType = "song"
			case "MUSIC_VIDEO_TYPE_PODCAST_EPISODE":
				resultType = "episode"
			default:
				resultType = "video"
			}
		}
	}

	result.ResultType = model.SearchResultType(resultType)

	defaultOffset := 0
	if resultType == "album" {
		defaultOffset = 2
	}

	if resultType != "artist" {
		result.Title = GetItemText(data, 0, 0)
	}

	switch resultType {
	case "artist":
		result.Artist = GetItemText(data, 0, 0)
		ParseMenuPlaylists(data, result)

	case "album":
		result.Artist = GetItemText(data, 1, 0)
		playNav := NavMap(data, concatPath(pathPlayButton, "playNavigationEndpoint"))
		if playNav != nil {
			result.PlaylistID = NavStr(playNav, []any{"watchEndpoint", "playlistId"})
		}

	case "playlist":
		flex := GetFlexColumnItem(data, 1)
		if flex != nil {
			runs := NavList(flex, []any{"text", "runs"})
			hasAuthor := len(runs) == defaultOffset+3
			itemCountStr := GetItemText(data, 1, defaultOffset+boolToInt(hasAuthor)*2)
			itemCountParts := strings.SplitN(itemCountStr, " ", 2)
			if len(itemCountParts) > 0 {
				if n, err := strconv.Atoi(itemCountParts[0]); err == nil {
					result.ItemCount = n
				}
			}
			if hasAuthor {
				result.Author = GetItemText(data, 1, defaultOffset)
			}
		}

	case "station":
		result.VideoID = NavStr(data, []any{"navigationEndpoint", "watchEndpoint", "videoId"})
		result.PlaylistID = NavStr(data, []any{"navigationEndpoint", "watchEndpoint", "playlistId"})

	case "profile":
		result.Name = GetItemText(data, 1, 2)

	case "song":
		result.Album = nil
		parseSongMenuData(data, result)
	}

	if resultType == "song" || resultType == "video" || resultType == "episode" {
		playNav := NavMap(data, pathPlayButton)
		if playNav != nil {
			result.VideoID = NavStr(playNav, []any{"playNavigationEndpoint", "watchEndpoint", "videoId"})
		}
		result.VideoType = videoType
	}

	if resultType == "song" || resultType == "video" || resultType == "album" {
		result.Duration = ""
		result.Year = ""
		flex := GetFlexColumnItem(data, 1)
		if flex != nil {
			runs := NavList(flex, []any{"text", "runs"})
			if flex2 := GetFlexColumnItem(data, 2); flex2 != nil {
				runs2 := NavList(flex2, []any{"text", "runs"})
				dummy := map[string]any{"text": ""}
				runs = append(runs, dummy)
				runs = append(runs, runs2...)
			}
			ParseSongRuns(runs, true, result)
		}
	}

	if resultType == "artist" || resultType == "album" || resultType == "playlist" || resultType == "profile" || resultType == "podcast" {
		result.BrowseID = NavStr(data, pathBrowseID)
	}

	if resultType == "song" || resultType == "album" {
		badgeLabel := NavStr(data, pathBadgeLabel)
		result.IsExplicit = badgeLabel != ""
	}

	if resultType == "episode" {
		flex := GetFlexColumnItem(data, 1)
		if flex != nil {
			runs := NavList(flex, []any{"text", "runs"})
			if len(runs) > defaultOffset {
				runs = runs[defaultOffset:]
				hasDate := len(runs) > 1
				liveBadge := Nav(data, []any{"badges", 0, "liveBadgeRenderer"}, true)
				result.Live = liveBadge != nil
				if hasDate {
					if r, ok := runs[0].(map[string]any); ok {
						result.Date, _ = r["text"].(string)
					}
				}
				idx := boolToInt(hasDate) * 2
				if len(runs) > idx {
					if r, ok := runs[idx].(map[string]any); ok {
						id, name := ParseIDName(r)
						result.Podcast = &model.PodcastRef{ID: id, Name: name}
					}
				}
			}
		}
	}

	result.Thumbnails = ParseThumbnails(data)
	return result
}

// ParseSearchResults parses a list of musicShelfRenderer content items.
func ParseSearchResults(results []any, resultType string, category string) []*model.SearchResult {
	var out []*model.SearchResult
	for _, r := range results {
		item, _ := r.(map[string]any)
		if item == nil {
			continue
		}
		mrlir, _ := item[pathMRLIR].(map[string]any)
		if mrlir == nil {
			continue
		}
		out = append(out, ParseSearchResult(mrlir, resultType, category))
	}
	return out
}

// ParseTopResult parses a musicCardShelfRenderer (top result card).
func ParseTopResult(data map[string]any, searchResultTypes []string) *model.SearchResult {
	subtitleText := NavStr(data, []any{"subtitle", "runs", 0, "text"})
	resultType := getSearchResultType(subtitleText, searchResultTypes)

	category := NavStr(data, []any{"header", "musicCardShelfHeaderBasicRenderer", "title", "runs", 0, "text"})
	if category == "" {
		category = "Top result"
	}

	result := &model.SearchResult{
		Category:   category,
		ResultType: model.SearchResultType(resultType),
	}

	if resultType == "artist" {
		sub2 := NavStr(data, []any{"subtitle", "runs", 2, "text"})
		if sub2 != "" {
			parts := strings.SplitN(sub2, " ", 2)
			result.Subscribers = parts[0]
		}
		titleRuns := NavList(data, []any{"title", "runs"})
		ParseSongRuns(titleRuns, false, result)
	}

	if resultType == "song" || resultType == "video" || resultType == "album" {
		result.VideoID = NavStr(data, []any{"onTap", "watchEndpoint", "videoId"})
		result.VideoType = NavStr(data, concatPath([]any{"onTap"}, pathNavigationVideoType...))
		result.Title = NavStr(data, []any{"title", "runs", 0, "text"})
		subtitleRuns := NavList(data, []any{"subtitle", "runs"})
		if len(subtitleRuns) > 2 {
			ParseSongRuns(subtitleRuns[2:], false, result)
		}
	}

	if resultType == "album" {
		result.BrowseID = NavStr(data, []any{"title", "runs", 0, "navigationEndpoint", "browseEndpoint", "browseId"})
		buttons := NavList(data, []any{"buttons"})
		if len(buttons) > 0 {
			btn, _ := buttons[0].(map[string]any)
			if btn != nil {
				cmd := NavMap(btn, []any{"buttonRenderer", "command"})
				if cmd != nil {
					result.PlaylistID = NavStr(cmd, []any{"watchEndpoint", "playlistId"})
				}
			}
		}
	}

	if resultType == "playlist" {
		result.PlaylistID = NavStr(data, []any{"menu", "menuRenderer", "items", 0, "menuNavigationItemRenderer", "navigationEndpoint", "watchPlaylistEndpoint", "playlistId"})
		result.Title = NavStr(data, []any{"title", "runs", 0, "text"})
	}

	if resultType == "episode" {
		result.Title = NavStr(data, []any{"title", "runs", 0, "text"})
		overlayPath := []any{"thumbnail", "musicThumbnailRenderer", "onTap", "watchEndpoint"}
		result.VideoID = NavStr(data, concatPath(overlayPath, "videoId"))
		subtitleRuns := NavList(data, []any{"subtitle", "runs"})
		if len(subtitleRuns) > 2 {
			r0, _ := subtitleRuns[0].(map[string]any)
			result.Date, _ = r0["text"].(string)
			if len(subtitleRuns) > 4 {
				r2, _ := subtitleRuns[2].(map[string]any)
				id, name := ParseIDName(r2)
				result.Podcast = &model.PodcastRef{ID: id, Name: name}
			}
		}
	}

	result.Thumbnails = ParseThumbnails(data)
	return result
}

func getSearchResultType(resultTypeLocal string, resultTypesLocal []string) string {
	if resultTypeLocal == "" {
		return ""
	}
	lower := strings.ToLower(resultTypeLocal)
	allTypes := []string{"album", "artist", "playlist", "song", "video", "station", "profile", "podcast", "episode"}

	switch lower {
	case "single", "ep":
		return "album"
	}

	for _, t := range allTypes {
		if t == lower {
			return t
		}
	}
	for _, t := range resultTypesLocal {
		if strings.ToLower(t) == lower {
			return t
		}
	}
	return "album"
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func parseSongMenuData(data map[string]any, result *model.SearchResult) {
	items := NavList(data, pathMenuItems)
	if items == nil {
		return
	}
	for _, itemAny := range items {
		item, _ := itemAny.(map[string]any)
		if item == nil {
			continue
		}
		toggleItem, _ := item["toggleMenuServiceItemRenderer"].(map[string]any)
		if toggleItem == nil {
			continue
		}
		iconType := NavStr(toggleItem, []any{"defaultIcon", "iconType"})
		if iconType != "LIBRARY_ADD" && iconType != "LIBRARY_SAVED" {
			continue
		}
	}
}
