package parser_test

import (
	"testing"

	"github.com/JanikSachs/go_ytmusicapi/internal/parser"
)

func makeShelfWithContinuation(token string) map[string]any {
	return map[string]any{
		"continuations": []any{
			map[string]any{
				"nextContinuationData": map[string]any{
					"continuation":        token,
					"clickTrackingParams": "CABQ2q4E",
				},
			},
		},
	}
}

func TestExtractShelfContinuation_WithToken(t *testing.T) {
	shelf := makeShelfWithContinuation("CBQQAA==tokenABC")
	got := parser.ExtractShelfContinuation(shelf)
	if got != "CBQQAA==tokenABC" {
		t.Errorf("expected token %q, got %q", "CBQQAA==tokenABC", got)
	}
}

func TestExtractShelfContinuation_NoContinuations(t *testing.T) {
	shelf := map[string]any{}
	got := parser.ExtractShelfContinuation(shelf)
	if got != "" {
		t.Errorf("expected empty token, got %q", got)
	}
}

func TestExtractShelfContinuation_EmptyContinuations(t *testing.T) {
	shelf := map[string]any{
		"continuations": []any{},
	}
	got := parser.ExtractShelfContinuation(shelf)
	if got != "" {
		t.Errorf("expected empty token, got %q", got)
	}
}

func TestExtractShelfContinuation_MissingNextContinuationData(t *testing.T) {
	shelf := map[string]any{
		"continuations": []any{
			map[string]any{
				"someOtherKey": "value",
			},
		},
	}
	got := parser.ExtractShelfContinuation(shelf)
	if got != "" {
		t.Errorf("expected empty token, got %q", got)
	}
}
