package parser

import (
	"github.com/JanikSachs/go_ytmusicapi/internal/model"
)

// ParseArtist maps a raw browse response for an artist page to a typed Artist struct.
func ParseArtist(response map[string]any) *model.Artist {
	artist := &model.Artist{}

	header := NavMap(response, []any{"header", "musicImmersiveHeaderRenderer"})
	if header == nil {
		return artist
	}

	artist.Name = NavStr(header, []any{"title", "runs", 0, "text"})
	artist.Thumbnails = ParseThumbnails(header)

	subBtn := NavMap(header, []any{"subscriptionButton", "subscribeButtonRenderer"})
	if subBtn != nil {
		artist.ChannelID = NavStr(subBtn, []any{"channelId"})
		artist.Subscribers = NavStr(subBtn, []any{"subscriberCountText", "runs", 0, "text"})
	}

	artist.ShuffleID = NavStr(header, []any{"playButton", "buttonRenderer", "navigationEndpoint", "watchEndpoint", "playlistId"})
	artist.RadioID = NavStr(header, []any{"startRadioButton", "buttonRenderer", "navigationEndpoint", "watchEndpoint", "playlistId"})

	// results are in single column tab
	results := NavList(response, []any{
		"contents", "singleColumnBrowseResultsRenderer",
		"tabs", 0, "tabRenderer", "content",
		"sectionListRenderer", "contents",
	})

	for _, r := range results {
		section, _ := r.(map[string]any)
		if section == nil {
			continue
		}

		// description shelf
		if desc := NavMap(section, []any{"musicDescriptionShelfRenderer"}); desc != nil {
			artist.Description = NavStr(desc, []any{"description", "runs", 0, "text"})
			artist.Views = NavStr(desc, []any{"subheader", "runs", 0, "text"})
			continue
		}

		// songs shelf
		if shelf := NavMap(section, []any{"musicShelfRenderer"}); shelf != nil {
			// top-level music shelf on artist page = songs
			if artist.Songs == nil {
				artist.Songs = parseArtistSongsShelf(shelf)
			}
			continue
		}

		// carousel sections (albums, singles, videos, related)
		if carousel := NavMap(section, []any{"musicCarouselShelfRenderer"}); carousel != nil {
			parseArtistCarousel(carousel, artist)
		}
	}

	return artist
}

func parseArtistSongsShelf(shelf map[string]any) *model.ArtistSongsSection {
	section := &model.ArtistSongsSection{}

	// browseId comes from the title navigation endpoint
	titleRuns := NavList(shelf, []any{"title", "runs"})
	if len(titleRuns) > 0 {
		run, _ := titleRuns[0].(map[string]any)
		section.BrowseID = NavStr(run, []any{"navigationEndpoint", "browseEndpoint", "browseId"})
	}

	items := NavList(shelf, []any{"contents"})
	for _, itemAny := range items {
		item, _ := itemAny.(map[string]any)
		if item == nil {
			continue
		}
		mrlir := NavMap(item, []any{"musicResponsiveListItemRenderer"})
		if mrlir == nil {
			continue
		}
		song := parseArtistSong(mrlir)
		if song != nil {
			section.Results = append(section.Results, *song)
		}
	}
	return section
}

func parseArtistSong(item map[string]any) *model.ArtistSong {
	song := &model.ArtistSong{}

	song.Title = GetItemText(item, 0, 0)
	if song.Title == "" {
		return nil
	}

	song.VideoID = NavStr(item, []any{
		"overlay", "musicItemThumbnailOverlayRenderer",
		"content", "musicPlayButtonRenderer",
		"playNavigationEndpoint", "watchEndpoint", "videoId",
	})

	// col1 contains artist runs; col2 contains album
	col1 := GetFlexColumnItem(item, 1)
	if col1 != nil {
		runs := NavList(col1, []any{"text", "runs"})
		for i, runAny := range runs {
			if i%2 != 0 {
				continue
			}
			run, _ := runAny.(map[string]any)
			if run == nil {
				continue
			}
			text, _ := run["text"].(string)
			id := NavStr(run, []any{"navigationEndpoint", "browseEndpoint", "browseId"})
			song.Artists = append(song.Artists, model.ArtistRef{Name: text, ID: id})
		}
	}

	song.Album = GetItemText(item, 2, 0)

	song.Thumbnails = ParseThumbnails(item)

	return song
}

func parseArtistCarousel(carousel map[string]any, artist *model.Artist) {
	header := NavMap(carousel, []any{"header", "musicCarouselShelfBasicHeaderRenderer"})
	if header == nil {
		return
	}

	title := NavStr(header, []any{"title", "runs", 0, "text"})

	browseID := NavStr(header, []any{"title", "runs", 0, "navigationEndpoint", "browseEndpoint", "browseId"})
	params := NavStr(header, []any{"title", "runs", 0, "navigationEndpoint", "browseEndpoint", "params"})

	contents := NavList(carousel, []any{"contents"})

	switch title {
	case "Albums":
		artist.Albums = &model.ArtistSection{BrowseID: browseID, Params: params}
		artist.Albums.Results = parseArtistAlbums(contents)
	case "Singles":
		artist.Singles = &model.ArtistSection{BrowseID: browseID, Params: params}
		artist.Singles.Results = parseArtistAlbums(contents)
	case "Videos":
		artist.Videos = &model.ArtistSection{BrowseID: browseID, Params: params}
		artist.Videos.Results = parseArtistAlbums(contents)
	case "Related artists":
		artist.Related = parseRelatedArtists(contents)
	}
}

func parseArtistAlbums(contents []any) []model.ArtistAlbum {
	var albums []model.ArtistAlbum
	for _, itemAny := range contents {
		item, _ := itemAny.(map[string]any)
		if item == nil {
			continue
		}
		mtrir := NavMap(item, []any{"musicTwoRowItemRenderer"})
		if mtrir == nil {
			continue
		}
		album := model.ArtistAlbum{}
		album.Title = NavStr(mtrir, []any{"title", "runs", 0, "text"})
		album.BrowseID = NavStr(mtrir, []any{"title", "runs", 0, "navigationEndpoint", "browseEndpoint", "browseId"})
		album.Year = NavStr(mtrir, []any{"subtitle", "runs", 0, "text"})
		album.Thumbnails = ParseThumbnails(mtrir)
		albums = append(albums, album)
	}
	return albums
}

func parseRelatedArtists(contents []any) []model.RelatedArtist {
	var related []model.RelatedArtist
	for _, itemAny := range contents {
		item, _ := itemAny.(map[string]any)
		if item == nil {
			continue
		}
		mtrir := NavMap(item, []any{"musicTwoRowItemRenderer"})
		if mtrir == nil {
			continue
		}
		r := model.RelatedArtist{}
		r.Title = NavStr(mtrir, []any{"title", "runs", 0, "text"})
		r.BrowseID = NavStr(mtrir, []any{"navigationEndpoint", "browseEndpoint", "browseId"})
		r.Subscribers = NavStr(mtrir, []any{"subtitle", "runs", 0, "text"})
		r.Thumbnails = ParseThumbnails(mtrir)
		related = append(related, r)
	}
	return related
}
