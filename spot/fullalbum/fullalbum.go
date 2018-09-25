package fullalbum

import (
	"fmt"
	"sort"

	"github.com/sirupsen/logrus"

	"github.com/kristofferostlund/spot/spot/config"
	"github.com/kristofferostlund/spot/spot/spotifytrack/simpletrack"
	"github.com/kristofferostlund/spot/spot/utils"

	"github.com/zmb3/spotify"
)

var albumCache = map[spotify.ID]spotify.FullAlbum{}

func Get(client spotify.Client, id spotify.ID) (spotify.FullAlbum, error) {
	if album, exists := albumCache[id]; exists {
		return album, nil
	}

	album, err := client.GetAlbum(id)
	if err != nil {
		return spotify.FullAlbum{}, fmt.Errorf("Failed to get full album %s: %v", id, err)
	}

	albumCache[id] = *album

	return *album, nil
}

func GetMany(client spotify.Client, albumIDs []spotify.ID) ([]spotify.FullAlbum, error) {
	albums := []spotify.FullAlbum{}
	uncachedAlbumIDs := []spotify.ID{}
	albumMap := map[spotify.ID]spotify.FullAlbum{}

	for _, id := range albumIDs {
		if album, exists := albumCache[id]; exists {
			albumMap[id] = album

			continue
		}

		uncachedAlbumIDs = append(uncachedAlbumIDs, id)
	}

	chunks := utils.ChunkIDs(uncachedAlbumIDs, config.AlbumChunkSize)

	for _, chunk := range chunks {
		albumChunk, err := client.GetAlbums(chunk...)
		if err != nil {
			return albums, fmt.Errorf("Failed to get %d album(s): %v", len(albumIDs), err)
		}

		for _, album := range albumChunk {
			albumMap[album.ID] = *album
			albumCache[album.ID] = *album
		}
	}

	for _, id := range albumIDs {
		if album, exists := albumMap[id]; exists {
			albums = append(albums, album)

			continue
		}

		logrus.Warnf("Somehow missed the album for ID %v", id)
	}

	return albums, nil
}

func GetAlbumByTrack(client spotify.Client, track spotify.FullTrack) (spotify.FullAlbum, error) {
	album, err := Get(client, track.Album.ID)
	if err != nil {
		return spotify.FullAlbum{}, err
	}

	if album.Tracks.Total < config.MinimumAlbumTotalCount {
		for _, artist := range track.Artists {
			logrus.Infof("Listing albums for artist %s", artist.Name)

			artistAlbums, err := listArtistAlbums(client, artist.ID)
			if err != nil {
				return album, err
			}

			sort.Slice(artistAlbums, func(i, j int) bool {
				return artistAlbums[i].Tracks.Total > artistAlbums[j].Tracks.Total
			})

			for _, artistAlbum := range artistAlbums {
				isMatch := simpletrack.InSlice(artistAlbum.Tracks.Tracks, track.SimpleTrack) &&
					artistAlbum.Tracks.Total > album.Tracks.Total

				if isMatch {
					return artistAlbum, nil
				}
			}
		}
	}

	return album, nil
}

func listArtistAlbums(client spotify.Client, artistID spotify.ID) ([]spotify.FullAlbum, error) {
	pageLimit := 50
	totalCount := -1
	albumType := spotify.AlbumTypeAlbum | spotify.AlbumTypeSingle
	albums := []spotify.SimpleAlbum{}

	for totalCount != len(albums) {
		offset := len(albums)
		options := spotify.Options{Limit: &pageLimit, Offset: &offset}

		page, err := client.GetArtistAlbumsOpt(artistID, &options, &albumType)
		if err != nil {
			return []spotify.FullAlbum{}, fmt.Errorf("Failed to get albums for the artist %s: %v", artistID, err)
		}

		albums = append(albums, page.Albums...)
		totalCount = page.Total
	}

	return GetMany(client, utils.GetSpotifyIDs(albums))
}
