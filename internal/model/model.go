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
