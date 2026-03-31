package endpoints_test

import (
	"testing"

	"github.com/JanikSachs/go_ytmusicapi/internal/endpoints"
)

func TestGetSearchParams(t *testing.T) {
	tests := []struct {
		name           string
		filter         string
		scope          string
		ignoreSpelling bool
		wantErr        bool
		wantEmpty      bool
		wantContains   string
	}{
		{
			name:      "no args returns empty",
			wantEmpty: true,
		},
		{
			name:         "songs filter",
			filter:       "songs",
			wantContains: "EgWKAQ",
		},
		{
			name:         "videos filter",
			filter:       "videos",
			wantContains: "EgWKAQ",
		},
		{
			name:         "albums filter",
			filter:       "albums",
			wantContains: "EgWKAQ",
		},
		{
			name:         "artists filter",
			filter:       "artists",
			wantContains: "EgWKAQ",
		},
		{
			name:         "playlists filter",
			filter:       "playlists",
			wantContains: "Eg-KAQwIABAAGAAgACgB",
		},
		{
			name:         "featured_playlists filter",
			filter:       "featured_playlists",
			wantContains: "EgeKAQQoA",
		},
		{
			name:         "community_playlists filter",
			filter:       "community_playlists",
			wantContains: "EgeKAQQoA",
		},
		{
			name:         "uploads scope",
			scope:        "uploads",
			wantContains: "agIYAw",
		},
		{
			name:         "library scope",
			scope:        "library",
			wantContains: "agIYBA",
		},
		{
			name:         "library scope with songs filter",
			scope:        "library",
			filter:       "songs",
			wantContains: "EgWKAQ",
		},
		{
			name:           "ignore spelling",
			ignoreSpelling: true,
			wantContains:   "EhGKAQ",
		},
		{
			name:    "invalid filter",
			filter:  "badfilter",
			wantErr: true,
		},
		{
			name:    "invalid scope",
			scope:   "badscope",
			wantErr: true,
		},
		{
			name:    "uploads with filter",
			scope:   "uploads",
			filter:  "songs",
			wantErr: true,
		},
		{
			name:    "library with community_playlists",
			scope:   "library",
			filter:  "community_playlists",
			wantErr: true,
		},
		{
			name:    "library with featured_playlists",
			scope:   "library",
			filter:  "featured_playlists",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := endpoints.GetSearchParams(tt.filter, tt.scope, tt.ignoreSpelling)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSearchParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if tt.wantEmpty && got != "" {
				t.Errorf("GetSearchParams() = %q, want empty", got)
			}
			if tt.wantContains != "" {
				found := false
				for i := 0; i <= len(got)-len(tt.wantContains); i++ {
					if got[i:i+len(tt.wantContains)] == tt.wantContains {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetSearchParams() = %q, want to contain %q", got, tt.wantContains)
				}
			}
		})
	}
}

func TestSearchBody(t *testing.T) {
	body := endpoints.SearchBody("hello", "params123")
	if body["query"] != "hello" {
		t.Errorf("expected query=hello, got %v", body["query"])
	}
	if body["params"] != "params123" {
		t.Errorf("expected params=params123, got %v", body["params"])
	}

	bodyNoParams := endpoints.SearchBody("hello", "")
	if _, ok := bodyNoParams["params"]; ok {
		t.Error("expected no params key when params is empty")
	}
}

func TestContext(t *testing.T) {
	ctx := endpoints.Context("en", "US", "")
	ctxMap, ok := ctx["context"].(map[string]any)
	if !ok {
		t.Fatal("expected context key")
	}
	client, ok := ctxMap["client"].(map[string]any)
	if !ok {
		t.Fatal("expected client key")
	}
	if client["clientName"] != "WEB_REMIX" {
		t.Errorf("expected clientName=WEB_REMIX, got %v", client["clientName"])
	}
	if client["hl"] != "en" {
		t.Errorf("expected hl=en, got %v", client["hl"])
	}
	if client["gl"] != "US" {
		t.Errorf("expected gl=US, got %v", client["gl"])
	}
}
