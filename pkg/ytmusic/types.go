package ytmusic

import "github.com/JanikSachs/go_ytmusicapi/internal/model"

// Thumbnail is an image thumbnail.
type Thumbnail = model.Thumbnail

// ArtistRef is a reference to an artist.
type ArtistRef = model.ArtistRef

// AlbumRef is a reference to an album.
type AlbumRef = model.AlbumRef

// PodcastRef is a reference to a podcast.
type PodcastRef = model.PodcastRef

// SearchResultType is the kind of a search result.
type SearchResultType = model.SearchResultType

const (
	ResultTypeSong     = model.ResultTypeSong
	ResultTypeVideo    = model.ResultTypeVideo
	ResultTypeAlbum    = model.ResultTypeAlbum
	ResultTypeArtist   = model.ResultTypeArtist
	ResultTypePlaylist = model.ResultTypePlaylist
	ResultTypeStation  = model.ResultTypeStation
	ResultTypeProfile  = model.ResultTypeProfile
	ResultTypePodcast  = model.ResultTypePodcast
	ResultTypeEpisode  = model.ResultTypeEpisode
)

// SearchResult is a single item in a search response.
type SearchResult = model.SearchResult

// SearchPage holds one page of search results together with an optional
// continuation token that can be passed to Client.SearchNext to retrieve
// the following page.
type SearchPage struct {
	// Results contains the search results for this page.
	Results []*model.SearchResult

	// Continuation is the opaque token needed to fetch the next page.
	// It is empty when there are no further pages.
	Continuation string

	// HasMore reports whether a following page is available.
	HasMore bool
}

// SearchOptions configures a search request.
type SearchOptions struct {
	// Filter limits results to a specific type.
	// Allowed: songs, videos, albums, artists, playlists, community_playlists,
	//          featured_playlists, uploads, profiles, podcasts, episodes.
	Filter string

	// Scope limits the search to a library or uploads.
	// Allowed: library, uploads.
	Scope string

	// Limit is the maximum number of results to return. Default: 20.
	Limit int

	// IgnoreSpelling disables YTM's spell correction.
	IgnoreSpelling bool
}
