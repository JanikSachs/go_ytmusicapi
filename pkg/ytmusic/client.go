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
		"User-Agent":   transport.UserAgent,
		"Accept":       "*/*",
		"Content-Type": "application/json",
		"Origin":       transport.YTMDomain,
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

// Search performs a YouTube Music search and returns all results on the first page.
// opts is optional; only the first SearchOptions element is used.
func (c *Client) Search(ctx context.Context, query string, opts ...SearchOptions) ([]*model.SearchResult, error) {
	page, err := c.SearchPage(ctx, query, opts...)
	if err != nil {
		return nil, err
	}
	return page.Results, nil
}

// SearchPage performs a YouTube Music search and returns the first page of
// results together with a continuation token for retrieving further pages.
// opts is optional; only the first SearchOptions element is used.
func (c *Client) SearchPage(ctx context.Context, query string, opts ...SearchOptions) (*SearchPage, error) {
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
	apiContext := endpoints.Context(c.cfg.language, c.cfg.location, c.cfg.userID)
	endpoints.MergeBody(body, apiContext)

	response, err := c.sendRequest(ctx, "search", body)
	if err != nil {
		return nil, fmt.Errorf("ytmusic: search request: %w", err)
	}

	results, continuation := parseSearchResponse(response, o.Filter, o.Scope)
	return &SearchPage{
		Results:      results,
		Continuation: continuation,
		HasMore:      continuation != "",
	}, nil
}

// SearchNext fetches the next page of search results using a continuation token
// returned by a previous SearchPage or SearchNext call.
// Returns an empty page when continuation is empty.
func (c *Client) SearchNext(ctx context.Context, continuation string) (*SearchPage, error) {
	if continuation == "" {
		return &SearchPage{}, nil
	}

	urlSuffix := endpoints.ContinuationURLParams(continuation)
	body := endpoints.ContinuationBody(continuation)
	apiContext := endpoints.Context(c.cfg.language, c.cfg.location, c.cfg.userID)
	endpoints.MergeBody(body, apiContext)

	response, err := c.sendRequest(ctx, "search", body, urlSuffix)
	if err != nil {
		return nil, fmt.Errorf("ytmusic: search continuation request: %w", err)
	}

	results, nextContinuation := parseContinuationResponse(response)
	return &SearchPage{
		Results:      results,
		Continuation: nextContinuation,
		HasMore:      nextContinuation != "",
	}, nil
}

// parseSearchResponse extracts search results and an optional continuation token
// from a raw search API response. The continuation token is taken from the last
// musicShelfRenderer that carries one.
func parseSearchResponse(response map[string]any, filter, scope string) ([]*model.SearchResult, string) {
	var results []*model.SearchResult
	var continuation string

	contents, ok := response["contents"]
	if !ok {
		return results, continuation
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
		return results, continuation
	}

	sectionList := parser.NavMap(searchContents, []any{"sectionListRenderer"})
	if sectionList == nil {
		return results, continuation
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

		if tok := parser.ExtractShelfContinuation(shelf); tok != "" {
			continuation = tok
		}
	}

	return results, continuation
}

// parseContinuationResponse extracts search results and an optional next
// continuation token from a pagination follow-up response. The response uses a
// different top-level key (continuationContents) compared to the initial search.
func parseContinuationResponse(response map[string]any) ([]*model.SearchResult, string) {
	cc, _ := response["continuationContents"].(map[string]any)
	if cc == nil {
		return nil, ""
	}
	shelf, _ := cc["musicShelfContinuation"].(map[string]any)
	if shelf == nil {
		return nil, ""
	}
	items, _ := shelf["contents"].([]any)
	results := parser.ParseSearchResults(items, "", "")
	continuation := parser.ExtractShelfContinuation(shelf)
	return results, continuation
}

func getLocalizedResultTypes() []string {
	return []string{
		"album", "artist", "playlist", "song", "video", "station",
		"profile", "podcast", "episode",
	}
}

func (c *Client) sendRequest(ctx context.Context, endpoint string, body map[string]any, urlSuffix ...string) (map[string]any, error) {
	suffix := ""
	if len(urlSuffix) > 0 {
		suffix = urlSuffix[0]
	}
	ctx_ := endpoints.Context(c.cfg.language, c.cfg.location, c.cfg.userID)
	endpoints.MergeBody(body, ctx_)

	response, err := c.transport.Post(ctx, endpoint, suffix, body, c.session.Headers())
	if err != nil {
		return nil, fmt.Errorf("ytmusic: send request: %w", err)
	}
	return response, nil
}

// GetArtist returns the detail page for the artist identified by browseId.
// browseId is the channelId for the artist (e.g. "UCxxxxxxxx").
func (c *Client) GetArtist(ctx context.Context, browseID string) (*model.Artist, error) {
	if browseID == "" {
		return nil, &UserError{Message: "browseId must not be empty"}
	}
	// Strip "MPLA" prefix if present (matches Python behaviour)
	if len(browseID) > 4 && browseID[:4] == "MPLA" {
		browseID = browseID[4:]
	}

	body := endpoints.BrowseBody(browseID)
	ctx_ := endpoints.Context(c.cfg.language, c.cfg.location, c.cfg.userID)
	endpoints.MergeBody(body, ctx_)

	response, err := c.sendRequest(ctx, "browse", body)
	if err != nil {
		return nil, fmt.Errorf("ytmusic: get artist request: %w", err)
	}
	return parser.ParseArtist(response), nil
}

// GetAlbum returns the detail page for the album identified by browseId.
// browseId must start with "MPRE".
func (c *Client) GetAlbum(ctx context.Context, browseID string) (*model.Album, error) {
	if browseID == "" {
		return nil, &UserError{Message: "browseId must not be empty"}
	}
	if len(browseID) < 4 || browseID[:4] != "MPRE" {
		return nil, &UserError{Message: "invalid album browseId: must start with MPRE"}
	}

	body := endpoints.BrowseBody(browseID)
	ctx_ := endpoints.Context(c.cfg.language, c.cfg.location, c.cfg.userID)
	endpoints.MergeBody(body, ctx_)

	response, err := c.sendRequest(ctx, "browse", body)
	if err != nil {
		return nil, fmt.Errorf("ytmusic: get album request: %w", err)
	}
	return parser.ParseAlbum(response), nil
}

// GetPlaylist returns the detail page for the playlist identified by playlistId.
// playlistId should be the bare playlist ID (without a "VL" prefix).
func (c *Client) GetPlaylist(ctx context.Context, playlistID string) (*model.Playlist, error) {
	if playlistID == "" {
		return nil, &UserError{Message: "playlistId must not be empty"}
	}

	browseID := playlistID
	if len(browseID) < 2 || browseID[:2] != "VL" {
		browseID = "VL" + playlistID
	}

	body := endpoints.BrowseBody(browseID)
	ctx_ := endpoints.Context(c.cfg.language, c.cfg.location, c.cfg.userID)
	endpoints.MergeBody(body, ctx_)

	response, err := c.sendRequest(ctx, "browse", body)
	if err != nil {
		return nil, fmt.Errorf("ytmusic: get playlist request: %w", err)
	}
	return parser.ParsePlaylist(playlistID, response), nil
}

// GetSong returns basic metadata for the song/video identified by videoId.
// Note: streaming URLs require additional authenticated context and are not included.
// TODO: streaming URL resolution via the player endpoint with a valid signatureTimestamp.
func (c *Client) GetSong(ctx context.Context, videoID string) (*model.Song, error) {
	if videoID == "" {
		return nil, &UserError{Message: "videoId must not be empty"}
	}

	body := endpoints.PlayerBody(videoID)
	ctx_ := endpoints.Context(c.cfg.language, c.cfg.location, c.cfg.userID)
	endpoints.MergeBody(body, ctx_)

	response, err := c.sendRequest(ctx, "player", body)
	if err != nil {
		return nil, fmt.Errorf("ytmusic: get song request: %w", err)
	}

	song := &model.Song{VideoID: videoID}
	// Basic metadata from videoDetails
	vd := parser.NavMap(response, []any{"videoDetails"})
	if vd != nil {
		song.Title, _ = vd["title"].(string)
		song.Views, _ = vd["viewCount"].(string)
		// thumbnails
		song.Thumbnails = parser.ParseThumbnails(vd)
	}
	return song, nil
}
