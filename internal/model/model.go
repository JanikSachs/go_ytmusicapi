// Package model defines internal data representations.
package model

// Thumbnail represents a single image thumbnail.
type Thumbnail struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// ArtistRef is a reference to an artist (name + optional browse ID).
type ArtistRef struct {
	Name string `json:"name"`
	ID   string `json:"id,omitempty"`
}

// AlbumRef is a reference to an album.
type AlbumRef struct {
	Name string `json:"name"`
	ID   string `json:"id,omitempty"`
}

// PodcastRef is a reference to a podcast.
type PodcastRef struct {
	Name string `json:"name"`
	ID   string `json:"id,omitempty"`
}

// FeedbackTokens hold library add/remove feedback tokens.
type FeedbackTokens struct {
	Add    *string `json:"add"`
	Remove *string `json:"remove"`
}

// SearchResultType identifies the kind of search result.
type SearchResultType string

const (
	ResultTypeSong     SearchResultType = "song"
	ResultTypeVideo    SearchResultType = "video"
	ResultTypeAlbum    SearchResultType = "album"
	ResultTypeArtist   SearchResultType = "artist"
	ResultTypePlaylist SearchResultType = "playlist"
	ResultTypeStation  SearchResultType = "station"
	ResultTypeProfile  SearchResultType = "profile"
	ResultTypePodcast  SearchResultType = "podcast"
	ResultTypeEpisode  SearchResultType = "episode"
)

// Artist represents the full detail of a YouTube Music artist page.
type Artist struct {
	Name           string           `json:"name"`
	ChannelID      string           `json:"channelId"`
	Description    string           `json:"description,omitempty"`
	Views          string           `json:"views,omitempty"`
	Subscribers    string           `json:"subscribers,omitempty"`
	ShuffleID      string           `json:"shuffleId,omitempty"`
	RadioID        string           `json:"radioId,omitempty"`
	Thumbnails     []Thumbnail      `json:"thumbnails,omitempty"`
	Songs          *ArtistSongsSection `json:"songs,omitempty"`
	Albums         *ArtistSection   `json:"albums,omitempty"`
	Singles        *ArtistSection   `json:"singles,omitempty"`
	Videos         *ArtistSection   `json:"videos,omitempty"`
	Related        []RelatedArtist  `json:"related,omitempty"`
}

// ArtistSongsSection holds the top songs listed on an artist page.
type ArtistSongsSection struct {
	BrowseID string       `json:"browseId,omitempty"`
	Results  []ArtistSong `json:"results"`
}

// ArtistSong is a single song entry on an artist page.
type ArtistSong struct {
	VideoID     string      `json:"videoId,omitempty"`
	Title       string      `json:"title"`
	Artists     []ArtistRef `json:"artists,omitempty"`
	Album       string      `json:"album,omitempty"`
	Thumbnails  []Thumbnail `json:"thumbnails,omitempty"`
	Duration    string      `json:"duration,omitempty"`
	DurationSec *int        `json:"duration_seconds,omitempty"`
}

// ArtistSection holds a section of an artist's releases (albums, singles, videos).
type ArtistSection struct {
	BrowseID string        `json:"browseId,omitempty"`
	Params   string        `json:"params,omitempty"`
	Results  []ArtistAlbum `json:"results"`
}

// ArtistAlbum is a single album/single/video entry in an artist section.
type ArtistAlbum struct {
	Title      string      `json:"title"`
	BrowseID   string      `json:"browseId,omitempty"`
	Year       string      `json:"year,omitempty"`
	Thumbnails []Thumbnail `json:"thumbnails,omitempty"`
	IsExplicit bool        `json:"isExplicit,omitempty"`
}

// RelatedArtist is a related artist entry on an artist page.
type RelatedArtist struct {
	BrowseID    string      `json:"browseId,omitempty"`
	Title       string      `json:"title"`
	Subscribers string      `json:"subscribers,omitempty"`
	Thumbnails  []Thumbnail `json:"thumbnails,omitempty"`
}

// Album represents the full detail of a YouTube Music album browse page.
type Album struct {
	Title          string       `json:"title"`
	Type           string       `json:"type,omitempty"`
	Thumbnails     []Thumbnail  `json:"thumbnails,omitempty"`
	Description    string       `json:"description,omitempty"`
	Artists        []ArtistRef  `json:"artists,omitempty"`
	Year           string       `json:"year,omitempty"`
	TrackCount     int          `json:"trackCount,omitempty"`
	Duration       string       `json:"duration,omitempty"`
	AudioPlaylistID string      `json:"audioPlaylistId,omitempty"`
	Tracks         []AlbumTrack `json:"tracks"`
}

// AlbumTrack is a single track within an album.
type AlbumTrack struct {
	VideoID     string      `json:"videoId,omitempty"`
	Title       string      `json:"title"`
	Artists     []ArtistRef `json:"artists,omitempty"`
	Duration    string      `json:"duration,omitempty"`
	DurationSec *int        `json:"duration_seconds,omitempty"`
	IsExplicit  bool        `json:"isExplicit,omitempty"`
	TrackNumber int         `json:"trackNumber,omitempty"`
	InLibrary   bool        `json:"inLibrary,omitempty"`
}

// Playlist represents the full detail of a YouTube Music playlist browse page.
type Playlist struct {
	ID          string          `json:"id"`
	Privacy     string          `json:"privacy,omitempty"`
	Title       string          `json:"title"`
	Thumbnails  []Thumbnail     `json:"thumbnails,omitempty"`
	Description string          `json:"description,omitempty"`
	Author      *ArtistRef      `json:"author,omitempty"`
	Year        string          `json:"year,omitempty"`
	Duration    string          `json:"duration,omitempty"`
	TrackCount  int             `json:"trackCount,omitempty"`
	Tracks      []PlaylistTrack `json:"tracks"`
}

// PlaylistTrack is a single track within a playlist.
type PlaylistTrack struct {
	VideoID     string      `json:"videoId,omitempty"`
	SetVideoID  string      `json:"setVideoId,omitempty"`
	Title       string      `json:"title"`
	Artists     []ArtistRef `json:"artists,omitempty"`
	Album       *AlbumRef   `json:"album,omitempty"`
	Duration    string      `json:"duration,omitempty"`
	DurationSec *int        `json:"duration_seconds,omitempty"`
	IsExplicit  bool        `json:"isExplicit,omitempty"`
	IsAvailable bool        `json:"isAvailable,omitempty"`
	InLibrary   bool        `json:"inLibrary,omitempty"`
	Thumbnails  []Thumbnail `json:"thumbnails,omitempty"`
}

// Song represents basic metadata for a YouTube Music song/video.
// Streaming URL resolution requires additional authenticated context and is not included here.
type Song struct {
	VideoID     string      `json:"videoId"`
	Title       string      `json:"title,omitempty"`
	Artists     []ArtistRef `json:"artists,omitempty"`
	Album       *AlbumRef   `json:"album,omitempty"`
	Duration    string      `json:"duration,omitempty"`
	DurationSec *int        `json:"duration_seconds,omitempty"`
	Year        string      `json:"year,omitempty"`
	Views       string      `json:"views,omitempty"`
	Thumbnails  []Thumbnail `json:"thumbnails,omitempty"`
}

// SearchResult is a union-like type for all search result kinds.
type SearchResult struct {
	ResultType     SearchResultType `json:"resultType"`
	Category       string           `json:"category,omitempty"`
	Title          string           `json:"title,omitempty"`
	VideoID        string           `json:"videoId,omitempty"`
	VideoType      string           `json:"videoType,omitempty"`
	BrowseID       string           `json:"browseId,omitempty"`
	PlaylistID     string           `json:"playlistId,omitempty"`
	Artist         string           `json:"artist,omitempty"`
	Artists        []ArtistRef      `json:"artists,omitempty"`
	Album          *AlbumRef        `json:"album,omitempty"`
	Duration       string           `json:"duration,omitempty"`
	DurationSec    *int             `json:"duration_seconds,omitempty"`
	Year           string           `json:"year,omitempty"`
	Views          string           `json:"views,omitempty"`
	IsExplicit     bool             `json:"isExplicit,omitempty"`
	InLibrary      bool             `json:"inLibrary,omitempty"`
	// ItemCount is the number of tracks in a playlist result (0 when unknown).
	ItemCount int    `json:"itemCount,omitempty"`
	// Author is the playlist owner or contributing artist name.
	Author    string `json:"author,omitempty"`
	Name           string           `json:"name,omitempty"`        // profile handle
	Subscribers    string           `json:"subscribers,omitempty"` // artist subscriber count
	Date           string           `json:"date,omitempty"`        // episode date
	Live           bool             `json:"live,omitempty"`        // episode live status
	Podcast        *PodcastRef      `json:"podcast,omitempty"`
	ShuffleID      string           `json:"shuffleId,omitempty"`
	RadioID        string           `json:"radioId,omitempty"`
	Thumbnails     []Thumbnail      `json:"thumbnails,omitempty"`
	FeedbackTokens *FeedbackTokens  `json:"feedbackTokens,omitempty"`
}
