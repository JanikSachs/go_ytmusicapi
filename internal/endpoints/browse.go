package endpoints

// BrowseBody builds the JSON request body for a browse endpoint call.
func BrowseBody(browseID string) map[string]any {
	return map[string]any{"browseId": browseID}
}

// PlayerBody builds the JSON request body for a player endpoint call.
func PlayerBody(videoID string) map[string]any {
	return map[string]any{
		"videoId":  videoID,
		"playbackContext": map[string]any{
			"contentPlaybackContext": map[string]any{
				"signatureTimestamp": 0,
			},
		},
	}
}
