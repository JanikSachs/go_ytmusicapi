package ytmusic

import (
	"context"
	"fmt"

	"github.com/JanikSachs/go_ytmusicapi/internal/auth"
	"github.com/JanikSachs/go_ytmusicapi/internal/endpoints"
	"github.com/JanikSachs/go_ytmusicapi/internal/model"
	"github.com/JanikSachs/go_ytmusicapi/internal/parser"
	"github.com/JanikSachs/go_ytmusicapi/internal/transport"
)

// Client is the public YouTube Music API client.
type Client struct {
	cfg       *config
	session   *auth.Session
	transport *transport.Client
}

// NewClient creates a new Client with the provided options.
func NewClient(opts ...Option) (*Client, error) {
	cfg := defaultConfig()
	for _, o := range opts {
		o(cfg)
	}

	var session *auth.Session
	var err error
	switch {
	case cfg.session != nil:
		session = cfg.session
	case cfg.browserAuthFile != "":
		session, err = auth.LoadBrowserAuthFromFile(cfg.browserAuthFile, cfg.authOrigin)
		if err != nil {
			return nil, fmt.Errorf("ytmusic: load browser auth from file: %w", err)
		}
	case cfg.authCookie != "":
		session, err = auth.NewBrowserAuth(cfg.authCookie, cfg.authOrigin)
		if err != nil {
			return nil, fmt.Errorf("ytmusic: create browser auth session: %w", err)
		}
	case cfg.oauthToken != "":
		session = auth.NewOAuthAuth(cfg.oauthToken)
	default:
		session = auth.NewUnauthenticated()
	}

	baseHeaders := map[string]string{
		"User-Agent":      transport.UserAgent,
		"Accept":          "*/*",
		"Accept-Encoding": "gzip, deflate",
		"Content-Type":    "application/json",
		"Origin":          transport.YTMDomain,
	}

	transportOpts := []transport.ClientOption{
		transport.WithHTTPClient(cfg.httpClient),
	}
	if len(cfg.headers) > 0 {
		transportOpts = append(transportOpts, transport.WithExtraHeaders(cfg.headers))
	}
	if cfg.retryPolicy.MaxAttempts > 1 {
		transportOpts = append(transportOpts, transport.WithRetry(cfg.retryPolicy))
	}

	tc := transport.NewClient(baseHeaders, transportOpts...)

	return &Client{cfg: cfg, session: session, transport: tc}, nil
}

// Search performs a YouTube Music search.
// opts is optional; only the first SearchOptions element is used.
func (c *Client) Search(ctx context.Context, query string, opts ...SearchOptions) ([]*model.SearchResult, error) {
	var o SearchOptions
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.Limit == 0 {
		o.Limit = 20
	}

	params, err := endpoints.GetSearchParams(o.Filter, o.Scope, o.IgnoreSpelling)
	if err != nil {
		return nil, fmt.Errorf("ytmusic: search params: %w", err)
	}

	body := endpoints.SearchBody(query, params)
	ctx_ := endpoints.Context(c.cfg.language, c.cfg.location, c.cfg.userID)
	endpoints.MergeBody(body, ctx_)

	response, err := c.sendRequest(ctx, "search", body)
	if err != nil {
		return nil, fmt.Errorf("ytmusic: search request: %w", err)
	}

	return parseSearchResponse(response, o.Filter, o.Scope), nil
}

func parseSearchResponse(response map[string]any, filter, scope string) []*model.SearchResult {
	var results []*model.SearchResult

	contents, ok := response["contents"]
	if !ok {
		return results
	}

	var searchContents map[string]any
	contentsMap, _ := contents.(map[string]any)
	if tabbed, ok := contentsMap["tabbedSearchResultsRenderer"]; ok {
		tabbedMap, _ := tabbed.(map[string]any)
		tabs, _ := tabbedMap["tabs"].([]any)
		tabIdx := 0
		if scope == "library" {
			tabIdx = 1
		} else if scope == "uploads" {
			tabIdx = 2
		}
		if tabIdx < len(tabs) {
			tab, _ := tabs[tabIdx].(map[string]any)
			searchContents = parser.NavMap(tab, []any{"tabRenderer", "content"})
		}
	} else {
		searchContents = contentsMap
	}

	if searchContents == nil {
		return results
	}

	sectionList := parser.NavMap(searchContents, []any{"sectionListRenderer"})
	if sectionList == nil {
		return results
	}
	sections, _ := sectionList["contents"].([]any)

	for _, sectionAny := range sections {
		section, _ := sectionAny.(map[string]any)
		if section == nil {
			continue
		}

		if card := parser.NavMap(section, []any{"musicCardShelfRenderer"}); card != nil {
			res := parser.ParseTopResult(card, getLocalizedResultTypes())
			if res != nil {
				results = append(results, res)
			}
			if itemsAny, ok := card["contents"]; ok {
				items, _ := itemsAny.([]any)
				results = append(results, parser.ParseSearchResults(items, "", "Top result")...)
			}
			continue
		}

		shelf, _ := section["musicShelfRenderer"].(map[string]any)
		if shelf == nil {
			continue
		}

		category := parser.NavStr(shelf, []any{"title", "runs", 0, "text"})

		resultType := filter
		if resultType == "community_playlists" || resultType == "featured_playlists" {
			resultType = "playlist"
		}

		items, _ := shelf["contents"].([]any)
		shelfResults := parser.ParseSearchResults(items, resultType, category)
		results = append(results, shelfResults...)
	}

	return results
}

func getLocalizedResultTypes() []string {
	return []string{
		"album", "artist", "playlist", "song", "video", "station",
		"profile", "podcast", "episode",
	}
}

func (c *Client) sendRequest(ctx context.Context, endpoint string, body map[string]any) (map[string]any, error) {
	// URL query params beyond "?alt=json" are not needed for the current endpoints;
	// all search/filter params are encoded inside the request body.
	return c.transport.Post(ctx, endpoint, "", body, c.session.Headers())
}
