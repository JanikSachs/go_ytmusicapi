package auth_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JanikSachs/go_ytmusicapi/internal/auth"
)

const testCookieStr = `__Secure-3PAPISID=abcdef1234567890/XYZ; __Secure-3PSID=somevalue; HSID=hsidvalue`

func TestNewUnauthenticated(t *testing.T) {
	s := auth.NewUnauthenticated()
	if s == nil {
		t.Fatal("NewUnauthenticated() returned nil")
	}
	if s.AuthType() != auth.Unauthorized {
		t.Errorf("AuthType() = %v, want Unauthorized", s.AuthType())
	}
	headers := s.Headers()
	if len(headers) != 0 {
		t.Errorf("Unauthenticated session should produce no headers, got %v", headers)
	}
}

func TestNewBrowserAuth_Valid(t *testing.T) {
	origin := "https://music.youtube.com"
	s, err := auth.NewBrowserAuth(testCookieStr, origin)
	if err != nil {
		t.Fatalf("NewBrowserAuth() unexpected error: %v", err)
	}
	if s.AuthType() != auth.Browser {
		t.Errorf("AuthType() = %v, want Browser", s.AuthType())
	}

	headers := s.Headers()
	if headers["Cookie"] != testCookieStr {
		t.Errorf("Cookie header = %q, want %q", headers["Cookie"], testCookieStr)
	}
	if !strings.HasPrefix(headers["Authorization"], "SAPISIDHASH ") {
		t.Errorf("Authorization header = %q, want SAPISIDHASH prefix", headers["Authorization"])
	}
	if headers["X-Origin"] != origin {
		t.Errorf("X-Origin header = %q, want %q", headers["X-Origin"], origin)
	}
}

func TestNewBrowserAuth_MissingCookie(t *testing.T) {
	_, err := auth.NewBrowserAuth("HSID=somevalue; SSID=anothervalue", "https://music.youtube.com")
	if err == nil {
		t.Fatal("NewBrowserAuth() expected error for missing __Secure-3PAPISID, got nil")
	}
}

func TestNewOAuthAuth(t *testing.T) {
	token := "ya29.test-token"
	s := auth.NewOAuthAuth(token)
	if s.AuthType() != auth.OAuth {
		t.Errorf("AuthType() = %v, want OAuth", s.AuthType())
	}

	headers := s.Headers()
	wantAuth := "Bearer " + token
	if headers["Authorization"] != wantAuth {
		t.Errorf("Authorization header = %q, want %q", headers["Authorization"], wantAuth)
	}
	if headers["X-Goog-Request-Time"] == "" {
		t.Error("X-Goog-Request-Time header missing for OAuth session")
	}
}

func TestLoadBrowserAuthFromFile_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cookies.txt")
	if err := os.WriteFile(path, []byte(testCookieStr+"\n"), 0600); err != nil {
		t.Fatalf("write cookie file: %v", err)
	}

	origin := "https://music.youtube.com"
	s, err := auth.LoadBrowserAuthFromFile(path, origin)
	if err != nil {
		t.Fatalf("LoadBrowserAuthFromFile() unexpected error: %v", err)
	}
	if s.AuthType() != auth.Browser {
		t.Errorf("AuthType() = %v, want Browser", s.AuthType())
	}
	headers := s.Headers()
	if headers["Cookie"] != testCookieStr {
		t.Errorf("Cookie header = %q, want %q", headers["Cookie"], testCookieStr)
	}
}

func TestLoadBrowserAuthFromFile_MissingFile(t *testing.T) {
	_, err := auth.LoadBrowserAuthFromFile("/nonexistent/path/cookies.txt", "https://music.youtube.com")
	if err == nil {
		t.Fatal("LoadBrowserAuthFromFile() expected error for missing file, got nil")
	}
}

func TestLoadBrowserAuthFromFile_InvalidCookie(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.txt")
	if err := os.WriteFile(path, []byte("HSID=somevalue"), 0600); err != nil {
		t.Fatalf("write cookie file: %v", err)
	}
	_, err := auth.LoadBrowserAuthFromFile(path, "https://music.youtube.com")
	if err == nil {
		t.Fatal("LoadBrowserAuthFromFile() expected error for invalid cookie, got nil")
	}
}
