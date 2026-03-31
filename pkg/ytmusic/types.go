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
