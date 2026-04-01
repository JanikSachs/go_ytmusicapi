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
// Artist is the full detail of a YouTube Music artist page.
type Artist = model.Artist

// ArtistSongsSection holds the top songs listed on an artist page.
type ArtistSongsSection = model.ArtistSongsSection

// ArtistSong is a single song entry on an artist page.
type ArtistSong = model.ArtistSong

// ArtistSection holds a section of an artist's releases (albums, singles, videos).
type ArtistSection = model.ArtistSection

// ArtistAlbum is a single album/single/video entry in an artist section.
type ArtistAlbum = model.ArtistAlbum

// RelatedArtist is a related artist entry on an artist page.
type RelatedArtist = model.RelatedArtist

// Album is the full detail of a YouTube Music album browse page.
type Album = model.Album

// AlbumTrack is a single track within an album.
type AlbumTrack = model.AlbumTrack

// Playlist is the full detail of a YouTube Music playlist browse page.
type Playlist = model.Playlist

// PlaylistTrack is a single track within a playlist.
type PlaylistTrack = model.PlaylistTrack

// Song represents basic metadata for a YouTube Music song/video.
type Song = model.Song

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
