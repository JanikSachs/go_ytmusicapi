// Package parser provides helpers for navigating raw YouTube Music API responses.
package parser

// Nav safely navigates a nested map[string]any / []any structure by a sequence of keys/indices.
// Returns nil if any step is missing or the path is invalid.
func Nav(obj any, path []any, optional ...bool) any {
	current := obj
	for _, key := range path {
		if current == nil {
			return nil
		}
		switch k := key.(type) {
		case string:
			m, ok := current.(map[string]any)
			if !ok {
				return nil
			}
			current = m[k]
		case int:
			s, ok := current.([]any)
			if !ok || k >= len(s) || k < 0 {
				return nil
			}
			current = s[k]
		default:
			return nil
		}
	}
	return current
}

// NavStr navigates and returns string value, or "" if not found.
func NavStr(obj any, path []any) string {
	v := Nav(obj, path, true)
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

// NavList navigates and returns []any, or nil if not found.
func NavList(obj any, path []any) []any {
	v := Nav(obj, path, true)
	if v == nil {
		return nil
	}
	l, _ := v.([]any)
	return l
}

// NavMap navigates and returns map[string]any, or nil if not found.
func NavMap(obj any, path []any) map[string]any {
	v := Nav(obj, path, true)
	if v == nil {
		return nil
	}
	m, _ := v.(map[string]any)
	return m
}
