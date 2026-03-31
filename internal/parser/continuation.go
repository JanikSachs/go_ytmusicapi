package parser

// ExtractShelfContinuation extracts the continuation token from a musicShelfRenderer
// or musicShelfContinuation map. The token lives at:
//
//	continuations[0].nextContinuationData.continuation
func ExtractShelfContinuation(shelf map[string]any) string {
	continuations, _ := shelf["continuations"].([]any)
	if len(continuations) == 0 {
		return ""
	}
	c, _ := continuations[0].(map[string]any)
	if c == nil {
		return ""
	}
	ncd, _ := c["nextContinuationData"].(map[string]any)
	if ncd == nil {
		return ""
	}
	token, _ := ncd["continuation"].(string)
	return token
}
