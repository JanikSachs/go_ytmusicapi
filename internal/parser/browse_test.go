package parser_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/JanikSachs/go_ytmusicapi/internal/parser"
)

func loadFixture(t *testing.T, name string) map[string]any {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", name, err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("failed to parse fixture %s: %v", name, err)
	}
	return m
}

// ---------------------------------------------------------------------------
// Album parser tests
// ---------------------------------------------------------------------------

func TestParseAlbum_Revival(t *testing.T) {
	fixture := loadFixture(t, "2024_03_get_album.json")
	album := parser.ParseAlbum(fixture)

	if album.Title != "Revival" {
		t.Errorf("Title = %q, want %q", album.Title, "Revival")
	}
	if album.Type != "Album" {
		t.Errorf("Type = %q, want %q", album.Type, "Album")
	}
	if album.Year != "2017" {
		t.Errorf("Year = %q, want %q", album.Year, "2017")
	}
	if album.TrackCount != 19 {
		t.Errorf("TrackCount = %d, want %d", album.TrackCount, 19)
	}
	if album.Duration == "" {
		t.Error("Duration must not be empty")
	}
	if len(album.Artists) == 0 {
		t.Fatal("Artists must not be empty")
	}
	if album.Artists[0].Name != "Eminem" {
		t.Errorf("Artists[0].Name = %q, want %q", album.Artists[0].Name, "Eminem")
	}
	if len(album.Tracks) == 0 {
		t.Fatal("Tracks must not be empty")
	}
	if album.Tracks[0].Title == "" {
		t.Error("Tracks[0].Title must not be empty")
	}
	if album.Tracks[0].VideoID == "" {
		t.Error("Tracks[0].VideoID must not be empty")
	}
	if album.Tracks[0].Duration == "" {
		t.Error("Tracks[0].Duration must not be empty")
	}
	if !album.Tracks[0].IsExplicit {
		t.Error("Tracks[0].IsExplicit expected true for first Revival track")
	}
}

func TestParseAlbum_EmptyResponse(t *testing.T) {
	album := parser.ParseAlbum(map[string]any{})
	if album == nil {
		t.Fatal("ParseAlbum must not return nil")
	}
	if album.Title != "" {
		t.Errorf("expected empty title, got %q", album.Title)
	}
	if album.Tracks == nil {
		t.Error("Tracks must be non-nil (empty slice) even on empty response")
	}
}

// ---------------------------------------------------------------------------
// Playlist parser tests
// ---------------------------------------------------------------------------

func TestParsePlaylist_Public(t *testing.T) {
	fixture := loadFixture(t, "2024_03_get_playlist_public.json")
	playlist := parser.ParsePlaylist("PLQ123", fixture)

	if playlist.Title == "" {
		t.Error("Title must not be empty")
	}
	if playlist.TrackCount == 0 {
		t.Error("TrackCount must be > 0")
	}
	if playlist.Duration == "" {
		t.Error("Duration must not be empty")
	}
	if len(playlist.Tracks) == 0 {
		t.Fatal("Tracks must not be empty")
	}
	t0 := playlist.Tracks[0]
	if t0.Title == "" {
		t.Error("Tracks[0].Title must not be empty")
	}
	if t0.VideoID == "" {
		t.Error("Tracks[0].VideoID must not be empty")
	}
}

func TestParsePlaylist_Owned(t *testing.T) {
	fixture := loadFixture(t, "2024_03_get_playlist.json")
	playlist := parser.ParsePlaylist("PLaZPMsuQNCsWn0iVMtGbaUXO6z-EdZaZm", fixture)

	if playlist.ID != "PLaZPMsuQNCsWn0iVMtGbaUXO6z-EdZaZm" {
		t.Errorf("ID = %q, want %q", playlist.ID, "PLaZPMsuQNCsWn0iVMtGbaUXO6z-EdZaZm")
	}
	if playlist.Title == "" {
		t.Error("Title must not be empty")
	}
	if playlist.Privacy == "" {
		t.Error("Privacy must not be empty for owned playlist")
	}
	if playlist.Author == nil {
		t.Error("Author must not be nil")
	}
}

func TestParsePlaylist_EmptyResponse(t *testing.T) {
	playlist := parser.ParsePlaylist("PL123", map[string]any{})
	if playlist == nil {
		t.Fatal("ParsePlaylist must not return nil")
	}
	if playlist.Tracks == nil {
		t.Error("Tracks must be non-nil (empty slice) even on empty response")
	}
}

// ---------------------------------------------------------------------------
// Artist parser: smoke-tests only (no real fixture available)
// ---------------------------------------------------------------------------

func TestParseArtist_EmptyResponse(t *testing.T) {
	artist := parser.ParseArtist(map[string]any{})
	if artist == nil {
		t.Fatal("ParseArtist must not return nil")
	}
	if artist.Name != "" {
		t.Errorf("expected empty name on empty response, got %q", artist.Name)
	}
}
