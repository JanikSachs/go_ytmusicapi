package parser

import (
	"strconv"
	"strings"
)

// GetFlexColumnItem returns the musicResponsiveListItemFlexColumnRenderer at index.
func GetFlexColumnItem(item map[string]any, index int) map[string]any {
	cols, _ := item["flexColumns"].([]any)
	if len(cols) <= index {
		return nil
	}
	col, _ := cols[index].(map[string]any)
	renderer, _ := col["musicResponsiveListItemFlexColumnRenderer"].(map[string]any)
	if renderer == nil {
		return nil
	}
	txt, ok := renderer["text"]
	if !ok {
		return nil
	}
	txtMap, _ := txt.(map[string]any)
	if _, hasRuns := txtMap["runs"]; !hasRuns {
		return nil
	}
	return renderer
}

// GetItemText returns the text from the run at runIndex inside flex column at colIndex.
func GetItemText(item map[string]any, colIndex int, runIndex int) string {
	col := GetFlexColumnItem(item, colIndex)
	if col == nil {
		return ""
	}
	runs := NavList(col, []any{"text", "runs"})
	if len(runs) <= runIndex {
		return ""
	}
	run, _ := runs[runIndex].(map[string]any)
	text, _ := run["text"].(string)
	return text
}

// ParseDuration parses a "m:ss" or "h:mm:ss" string to total seconds.
// Returns nil if the string is empty or not parseable.
func ParseDuration(s string) *int {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ":")
	multipliers := []int{1, 60, 3600}
	total := 0
	for i, p := range parts {
		idx := len(parts) - 1 - i
		if idx >= len(multipliers) {
			return nil
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil
		}
		total += n * multipliers[idx]
	}
	return &total
}

// ParseIDName extracts {id, name} from a sub-run that contains navigation.
func ParseIDName(run map[string]any) (id, name string) {
	if run == nil {
		return "", ""
	}
	id = NavStr(run, []any{"navigationEndpoint", "browseEndpoint", "browseId"})
	name, _ = run["text"].(string)
	return id, name
}

// DotSeparatorText is the bullet separator text used in YTMusic runs.
const DotSeparatorText = " • "
