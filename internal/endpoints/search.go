// Package endpoints builds request bodies and params for YouTube Music API endpoints.
package endpoints

import (
	"fmt"
	"time"
)

// ValidSearchFilters lists all valid filter values.
var ValidSearchFilters = []string{
	"albums", "artists", "playlists", "community_playlists",
	"featured_playlists", "songs", "videos", "profiles", "podcasts", "episodes",
}

// ValidSearchScopes lists all valid scope values.
var ValidSearchScopes = []string{"library", "uploads"}

// SearchBody builds the JSON body for a search request.
func SearchBody(query string, params string) map[string]any {
	body := map[string]any{"query": query}
	if params != "" {
		body["params"] = params
	}
	return body
}

// GetSearchParams generates the encoded params string for a search request.
func GetSearchParams(filter, scope string, ignoreSpelling bool) (string, error) {
	if filter != "" {
		valid := false
		for _, f := range ValidSearchFilters {
			if f == filter {
				valid = true
				break
			}
		}
		if !valid {
			return "", fmt.Errorf("invalid filter %q; valid filters: %v", filter, ValidSearchFilters)
		}
	}
	if scope != "" {
		valid := false
		for _, s := range ValidSearchScopes {
			if s == scope {
				valid = true
				break
			}
		}
		if !valid {
			return "", fmt.Errorf("invalid scope %q; valid scopes: %v", scope, ValidSearchScopes)
		}
	}
	if scope == "uploads" && filter != "" {
		return "", fmt.Errorf("no filter can be set when scope is 'uploads'")
	}
	if scope == "library" && (filter == "community_playlists" || filter == "featured_playlists") {
		return "", fmt.Errorf("%q cannot be set when searching library", filter)
	}

	p1, p2, p3 := "", "", ""

	if filter == "" && scope == "" && !ignoreSpelling {
		return "", nil
	}

	if scope == "uploads" {
		return "agIYAw%3D%3D", nil
	}
	if scope == "library" {
		if filter != "" {
			p1 = "EgWKAQ"
			p2 = getParam2(filter)
			p3 = "AWoKEAUQCRADEAoYBA%3D%3D"
		} else {
			return "agIYBA%3D%3D", nil
		}
	}
	if scope == "" && filter != "" {
		if filter == "playlists" {
			params := "Eg-KAQwIABAAGAAgACgB"
			if !ignoreSpelling {
				return params + "MABqChAEEAMQCRAFEAo%3D", nil
			}
			return params + "MABCAggBagoQBBADEAkQBRAK", nil
		} else if filter == "featured_playlists" || filter == "community_playlists" {
			p1 = "EgeKAQQoA"
			if filter == "featured_playlists" {
				p2 = "Dg"
			} else {
				p2 = "EA"
			}
			if !ignoreSpelling {
				p3 = "BagwQDhAKEAMQBBAJEAU%3D"
			} else {
				p3 = "BQgIIAWoMEA4QChADEAQQCRAF"
			}
		} else {
			p1 = "EgWKAQ"
			p2 = getParam2(filter)
			if !ignoreSpelling {
				p3 = "AWoMEA4QChADEAQQCRAF"
			} else {
				p3 = "AUICCAFqDBAOEAoQAxAEEAkQBQ%3D%3D"
			}
		}
	}
	if scope == "" && filter == "" && ignoreSpelling {
		return "EhGKAQ4IARABGAEgASgAOAFAAUICCAE%3D", nil
	}

	return p1 + p2 + p3, nil
}

func getParam2(filter string) string {
	m := map[string]string{
		"songs": "II", "videos": "IQ", "albums": "IY",
		"artists": "Ig", "playlists": "Io", "profiles": "JY",
		"podcasts": "JQ", "episodes": "JI",
	}
	return m[filter]
}

// Context returns the standard YTM context body.
func Context(language, location, userID string) map[string]any {
	client := map[string]any{
		"clientName":    "WEB_REMIX",
		"clientVersion": "1." + time.Now().UTC().Format("20060102") + ".01.00",
	}
	if language != "" {
		client["hl"] = language
	}
	if location != "" {
		client["gl"] = location
	}
	user := map[string]any{}
	if userID != "" {
		user["onBehalfOfUser"] = userID
	}
	return map[string]any{
		"context": map[string]any{
			"client": client,
			"user":   user,
		},
	}
}

// MergeBody merges extra into body (in-place, returns body).
func MergeBody(body, extra map[string]any) map[string]any {
	for k, v := range extra {
		body[k] = v
	}
	return body
}
