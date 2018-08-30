package fulltrack

import (
	"fmt"

	"github.com/kristofferostlund/spot/spot/config"
	"github.com/kristofferostlund/spot/spot/spotifytrack"
	"github.com/kristofferostlund/spot/spot/utils"
	"github.com/zmb3/spotify"
)

func Get(client spotify.Client, id spotify.ID) (spotify.FullTrack, error) {
	track, err := client.GetTrack(id)
	if err != nil {
		return spotify.FullTrack{}, fmt.Errorf("Failed to get track %s: %v", id, err)
	}

	return *track, nil
}

func GetUnique(tracks []spotify.FullTrack) []spotify.FullTrack {
	uniqueTracks := []spotify.FullTrack{}

	for _, track := range CreateMap(tracks) {
		uniqueTracks = append(uniqueTracks, track)
	}

	return uniqueTracks
}

func GetCompliment(baseTracks, tracks []spotify.FullTrack) []spotify.FullTrack {
	existing := CreateMap(baseTracks)
	complementingTracks := []spotify.FullTrack{}

	for _, track := range tracks {
		if !InMap(existing, track) {
			complementingTracks = append(complementingTracks, track)
		}
	}

	return complementingTracks
}

func CreateMap(tracks []spotify.FullTrack) spotifytrack.FullTrackMap {
	trackCache := spotifytrack.FullTrackMap{}

	for _, track := range tracks {
		trackCache[getMapKey(track)] = track
	}

	return trackCache
}

func InMap(trackCache spotifytrack.FullTrackMap, track spotify.FullTrack) bool {
	_, exists := trackCache[getMapKey(track)]

	return exists
}

func InSlice(tracks []spotify.FullTrack, track spotify.FullTrack) bool {
	trackCacheKey := getMapKey(track)

	for _, sliceTrack := range tracks {
		if trackCacheKey == getMapKey(sliceTrack) {
			return true
		}
	}

	return false
}

func GroupByArtists(tracks []spotify.FullTrack) spotifytrack.ArtistFullTrackMap {
	grouped := spotifytrack.ArtistFullTrackMap{}

	for _, track := range tracks {
		artistsKey := getArtistGroupKey(track.Artists)

		if _, exists := grouped[artistsKey]; exists {
			grouped[artistsKey] = append(grouped[artistsKey], track)
		} else {
			grouped[artistsKey] = []spotify.FullTrack{track}
		}
	}

	return grouped
}

func GetTrackCountByArtists(tracksByArtist spotifytrack.ArtistFullTrackMap, artists []spotify.SimpleArtist) int {
	artistKey := getArtistGroupKey(artists)

	if tracks, exists := tracksByArtist[artistKey]; exists {
		return len(tracks)
	} else {
		return 0
	}
}

func getMapKey(track spotify.FullTrack) string {
	return fmt.Sprintf(
		"%s:%s",
		track.Name,
		utils.JoinArtists(track.Artists, config.ArtistJoinCharacter),
	)
}

func getArtistGroupKey(artists []spotify.SimpleArtist) string {
	return utils.JoinArtists(artists, config.ArtistJoinCharacter)
}
