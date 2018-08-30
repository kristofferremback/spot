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

func Get(client spotify.Client, id spotify.ID) (spotify.FullAlbum, error) {
	album, err := client.GetAlbum(id)
	if err != nil {
		return spotify.FullAlbum{}, fmt.Errorf("Failed to get full album %s: %v", id, err)
	}

	return *album, nil
}

func GetMany(client spotify.Client, albumIDs []spotify.ID) ([]spotify.FullAlbum, error) {
	albums := []spotify.FullAlbum{}
	chunks := utils.ChunkIDs(albumIDs, config.AlbumChunkSize)

	for _, chunk := range chunks {
		albumChunk, err := client.GetAlbums(chunk...)
		if err != nil {
			return albums, fmt.Errorf("Failed to get %d album(s): %v", len(albumIDs), err)
		}

		for _, album := range albumChunk {
			albums = append(albums, *album)
		}
	}

	return albums, nil
}

func GetManyBySimpleAlbum(client spotify.Client, simpleAlbums []spotify.SimpleAlbum) ([]spotify.FullAlbum, error) {
	albumIDs := []spotify.ID{}

	for _, album := range simpleAlbums {
		albumIDs = append(albumIDs, album.ID)
	}

	return GetMany(client, albumIDs)
}

func GetAlbumByTrack(client spotify.Client, track spotify.FullTrack) (spotify.FullAlbum, error) {
	album, err := Get(client, track.Album.ID)
	if err != nil {
		return spotify.FullAlbum{}, err
	}

	if album.Tracks.Total < config.MinimumAlbumTotalCount {
		for _, artist := range track.Artists {
			logrus.Infof("Listing playlists for artist %s", artist.Name)

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

	return GetManyBySimpleAlbum(client, albums)
}
