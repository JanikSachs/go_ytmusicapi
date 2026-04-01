package endpoints_test

import (
	"strings"
	"testing"

	"github.com/JanikSachs/go_ytmusicapi/internal/endpoints"
)

// ---------------------------------------------------------------------------
// BrowseBody / PlayerBody
// ---------------------------------------------------------------------------

func TestBrowseBody(t *testing.T) {
	body := endpoints.BrowseBody("MPREb_abc123")
	if body["browseId"] != "MPREb_abc123" {
		t.Errorf("BrowseBody()['browseId'] = %v, want %q", body["browseId"], "MPREb_abc123")
	}
	if len(body) != 1 {
		t.Errorf("BrowseBody() should have 1 key, got %d", len(body))
	}
}

func TestBrowseBody_Empty(t *testing.T) {
	body := endpoints.BrowseBody("")
	if body["browseId"] != "" {
		t.Errorf("BrowseBody('')['browseId'] = %v, want empty string", body["browseId"])
	}
}

func TestPlayerBody(t *testing.T) {
	body := endpoints.PlayerBody("xFYQQPAOz7Y")
	if body["videoId"] != "xFYQQPAOz7Y" {
		t.Errorf("PlayerBody()['videoId'] = %v, want %q", body["videoId"], "xFYQQPAOz7Y")
	}
	ctx, ok := body["playbackContext"].(map[string]any)
	if !ok {
		t.Fatal("PlayerBody()['playbackContext'] should be map[string]any")
	}
	cpc, ok := ctx["contentPlaybackContext"].(map[string]any)
	if !ok {
		t.Fatal("PlayerBody().playbackContext['contentPlaybackContext'] should be map[string]any")
	}
	if cpc["signatureTimestamp"] != 0 {
		t.Errorf("signatureTimestamp = %v, want 0", cpc["signatureTimestamp"])
	}
}

// ---------------------------------------------------------------------------
// ContinuationURLParams / ContinuationBody
// ---------------------------------------------------------------------------

func TestContinuationURLParams(t *testing.T) {
	token := "CBQQAA==tokenABC"
	got := endpoints.ContinuationURLParams(token)
	if !strings.HasPrefix(got, "&ctoken=") {
		t.Errorf("ContinuationURLParams() = %q, expected to start with &ctoken=", got)
	}
	if !strings.Contains(got, "&continuation=") {
		t.Errorf("ContinuationURLParams() = %q, expected to contain &continuation=", got)
	}
	if !strings.HasSuffix(got, "&type=next") {
		t.Errorf("ContinuationURLParams() = %q, expected to end with &type=next", got)
	}
}

func TestContinuationURLParams_SpecialChars(t *testing.T) {
	token := "CB==QQ+/"
	got := endpoints.ContinuationURLParams(token)
	// Special characters should be URL-encoded (no raw +, /, =)
	if strings.Contains(got, "+") || strings.Contains(got, "/") {
		t.Errorf("ContinuationURLParams() should URL-encode special chars, got %q", got)
	}
}

func TestContinuationBody(t *testing.T) {
	token := "CBQQAA==tokenABC"
	body := endpoints.ContinuationBody(token)
	if body["continuation"] != token {
		t.Errorf("ContinuationBody()['continuation'] = %v, want %q", body["continuation"], token)
	}
	if len(body) != 1 {
		t.Errorf("ContinuationBody() should have 1 key, got %d", len(body))
	}
}

// ---------------------------------------------------------------------------
// MergeBody
// ---------------------------------------------------------------------------

func TestMergeBody(t *testing.T) {
	base := map[string]any{"query": "hello", "existing": "value"}
	extra := map[string]any{"context": map[string]any{"client": "WEB_REMIX"}, "extra": 42}

	result := endpoints.MergeBody(base, extra)
	if result["query"] != "hello" {
		t.Errorf("MergeBody()['query'] = %v, want %q", result["query"], "hello")
	}
	if result["extra"] != 42 {
		t.Errorf("MergeBody()['extra'] = %v, want 42", result["extra"])
	}
	if result["context"] == nil {
		t.Error("MergeBody()['context'] should not be nil")
	}
	// MergeBody should return the original base map (in-place)
	if &result == nil {
		t.Error("MergeBody() should return non-nil")
	}
}

func TestMergeBody_OverwritesExisting(t *testing.T) {
	base := map[string]any{"key": "original"}
	extra := map[string]any{"key": "overwritten"}
	result := endpoints.MergeBody(base, extra)
	if result["key"] != "overwritten" {
		t.Errorf("MergeBody() should overwrite: got %v, want %q", result["key"], "overwritten")
	}
}

func TestMergeBody_EmptyExtra(t *testing.T) {
	base := map[string]any{"key": "value"}
	result := endpoints.MergeBody(base, map[string]any{})
	if result["key"] != "value" {
		t.Errorf("MergeBody() with empty extra: got %v, want %q", result["key"], "value")
	}
}

// ---------------------------------------------------------------------------
// Context – extra cases not covered in existing tests
// ---------------------------------------------------------------------------

func TestContext_NoUserID(t *testing.T) {
	ctx := endpoints.Context("", "", "")
	ctxMap, ok := ctx["context"].(map[string]any)
	if !ok {
		t.Fatal("expected context key")
	}
	client, ok := ctxMap["client"].(map[string]any)
	if !ok {
		t.Fatal("expected client key")
	}
	if _, hasHL := client["hl"]; hasHL {
		t.Error("should not have 'hl' when language is empty")
	}
	if _, hasGL := client["gl"]; hasGL {
		t.Error("should not have 'gl' when location is empty")
	}
	user, ok := ctxMap["user"].(map[string]any)
	if !ok {
		t.Fatal("expected user key")
	}
	if _, hasUser := user["onBehalfOfUser"]; hasUser {
		t.Error("should not have 'onBehalfOfUser' when userID is empty")
	}
}

func TestContext_WithUserID(t *testing.T) {
	ctx := endpoints.Context("de", "DE", "user123")
	ctxMap := ctx["context"].(map[string]any)
	user := ctxMap["user"].(map[string]any)
	if user["onBehalfOfUser"] != "user123" {
		t.Errorf("onBehalfOfUser = %v, want %q", user["onBehalfOfUser"], "user123")
	}
}

// ---------------------------------------------------------------------------
// GetSearchParams – additional cases for playlists with ignoreSpelling
// ---------------------------------------------------------------------------

func TestGetSearchParams_PlaylistsIgnoreSpelling(t *testing.T) {
	got, err := endpoints.GetSearchParams("playlists", "", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == "" {
		t.Error("expected non-empty params for playlists+ignoreSpelling")
	}
}

func TestGetSearchParams_FeaturedPlaylistsIgnoreSpelling(t *testing.T) {
	got, err := endpoints.GetSearchParams("featured_playlists", "", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == "" {
		t.Error("expected non-empty params for featured_playlists+ignoreSpelling")
	}
}

func TestGetSearchParams_CommunityPlaylistsIgnoreSpelling(t *testing.T) {
	got, err := endpoints.GetSearchParams("community_playlists", "", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == "" {
		t.Error("expected non-empty params for community_playlists+ignoreSpelling")
	}
}

func TestGetSearchParams_LibraryWithFilter(t *testing.T) {
	filters := []string{"albums", "artists", "videos", "profiles", "podcasts", "episodes"}
	for _, f := range filters {
		got, err := endpoints.GetSearchParams(f, "library", false)
		if err != nil {
			t.Errorf("GetSearchParams(%q, library) unexpected error: %v", f, err)
			continue
		}
		if got == "" {
			t.Errorf("GetSearchParams(%q, library) should return non-empty params", f)
		}
	}
}

func TestGetSearchParams_SongsIgnoreSpelling(t *testing.T) {
	got, err := endpoints.GetSearchParams("songs", "", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == "" {
		t.Error("expected non-empty params for songs+ignoreSpelling")
	}
}
