package endpoints

import "net/url"

// ContinuationURLParams returns the URL query-parameter suffix required for a
// continuation (pagination) request against any YouTube Music endpoint, e.g.
//
//	baseURL + YTMParams + ContinuationURLParams(token)
func ContinuationURLParams(token string) string {
	escaped := url.QueryEscape(token)
	return "&ctoken=" + escaped + "&continuation=" + escaped + "&type=next"
}

// ContinuationBody builds the minimal JSON body for a continuation request.
// The context block must still be merged in by the caller.
func ContinuationBody(token string) map[string]any {
	return map[string]any{
		"continuation": token,
	}
}
