package parser

import (
	"regexp"

	"github.com/JanikSachs/go_ytmusicapi/internal/model"
)

var (
	reViews    = regexp.MustCompile(`^\d[^ ]* [^ ]*$`)
	reDuration = regexp.MustCompile(`^(\d+:)*\d+:\d+$`)
	reYear     = regexp.MustCompile(`^\d{4}$`)
)

type runResult struct {
	runType  string // "artist", "album", "views", "duration", "year"
	artist   *model.ArtistRef
	album    *model.AlbumRef
	views    string
	duration string
	year     string
}

func parseSongRun(run map[string]any) runResult {
	text, _ := run["text"].(string)
	if navMap, ok := run["navigationEndpoint"]; ok && navMap != nil {
		browseID := NavStr(run, []any{"navigationEndpoint", "browseEndpoint", "browseId"})
		item := &model.ArtistRef{Name: text, ID: browseID}
		if browseID != "" && len(browseID) >= 4 && browseID[:4] == "MPRE" {
			return runResult{runType: "album", album: &model.AlbumRef{Name: text, ID: browseID}}
		}
		return runResult{runType: "artist", artist: item}
	}
	if reViews.MatchString(text) {
		parts := splitSpace(text)
		return runResult{runType: "views", views: parts[0]}
	}
	if reDuration.MatchString(text) {
		return runResult{runType: "duration", duration: text}
	}
	if reYear.MatchString(text) {
		return runResult{runType: "year", year: text}
	}
	return runResult{runType: "artist", artist: &model.ArtistRef{Name: text}}
}

// ParseSongRuns processes a list of run items and fills in a SearchResult.
func ParseSongRuns(runs []any, skipTypeSpec bool, result *model.SearchResult) {
	if skipTypeSpec && len(runs) > 2 {
		r0, _ := runs[0].(map[string]any)
		r1, _ := runs[1].(map[string]any)
		r2, _ := runs[2].(map[string]any)
		sep, _ := r1["text"].(string)
		if r0 != nil && r2 != nil && sep == DotSeparatorText {
			if parseSongRun(r0).runType == "artist" && parseSongRun(r2).runType == "artist" {
				runs = runs[2:]
			}
		}
	}
	for i, runAny := range runs {
		if i%2 != 0 {
			continue
		}
		run, _ := runAny.(map[string]any)
		if run == nil {
			continue
		}
		res := parseSongRun(run)
		switch res.runType {
		case "artist":
			result.Artists = append(result.Artists, *res.artist)
		case "album":
			result.Album = res.album
		case "views":
			result.Views = res.views
		case "duration":
			result.Duration = res.duration
			result.DurationSec = ParseDuration(res.duration)
		case "year":
			result.Year = res.year
		}
	}
}

func splitSpace(s string) []string {
	var parts []string
	cur := ""
	for _, ch := range s {
		if ch == ' ' || ch == '\u00a0' {
			if cur != "" {
				parts = append(parts, cur)
				cur = ""
			}
		} else {
			cur += string(ch)
		}
	}
	if cur != "" {
		parts = append(parts, cur)
	}
	return parts
}
