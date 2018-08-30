package simpletrack

import (
	"fmt"

	"github.com/kristofferostlund/spot/spot/config"
	"github.com/kristofferostlund/spot/spot/spotifytrack"
	"github.com/kristofferostlund/spot/spot/utils"
	"github.com/zmb3/spotify"
)

func GetUnique(tracks []spotify.SimpleTrack) []spotify.SimpleTrack {
	uniqueTracks := []spotify.SimpleTrack{}

	for _, track := range CreateMap(tracks) {
		uniqueTracks = append(uniqueTracks, track)
	}

	return uniqueTracks
}

func GetCompliment(baseTracks, tracks []spotify.SimpleTrack) []spotify.SimpleTrack {
	existing := CreateMap(baseTracks)
	complementingTracks := []spotify.SimpleTrack{}

	for _, track := range tracks {
		if !InMap(existing, track) {
			complementingTracks = append(complementingTracks, track)
		}
	}

	return complementingTracks
}

func CreateMap(tracks []spotify.SimpleTrack) spotifytrack.SimpleTrackMap {
	trackCache := spotifytrack.SimpleTrackMap{}

	for _, track := range tracks {
		trackCache[getMapKey(track)] = track
	}

	return trackCache
}

func InMap(trackCache spotifytrack.SimpleTrackMap, track spotify.SimpleTrack) bool {
	_, exists := trackCache[getMapKey(track)]

	return exists
}

func InSlice(tracks []spotify.SimpleTrack, track spotify.SimpleTrack) bool {
	trackCacheKey := getMapKey(track)

	for _, sliceTrack := range tracks {
		if trackCacheKey == getMapKey(sliceTrack) {
			return true
		}
	}

	return false
}

func GroupByArtists(tracks []spotify.SimpleTrack) spotifytrack.ArtistSimpleTrackMap {
	grouped := spotifytrack.ArtistSimpleTrackMap{}

	for _, track := range tracks {
		for _, artist := range track.Artists {
			if _, exists := grouped[artist.Name]; exists {
				grouped[artist.Name] = append(grouped[artist.Name], track)
			} else {
				grouped[artist.Name] = []spotify.SimpleTrack{track}
			}
		}
	}

	return grouped
}

func GetTrackCountByArtist(tracksByArtist spotifytrack.ArtistSimpleTrackMap, artists []spotify.SimpleArtist) int {
	artistKey := getArtistGroupKey(artists)

	if tracks, exists := tracksByArtist[artistKey]; exists {
		return len(tracks)
	} else {
		return 0
	}
}

func getMapKey(track spotify.SimpleTrack) string {
	return fmt.Sprintf(
		"%s:%s",
		track.Name,
		utils.JoinArtists(track.Artists, config.ArtistJoinCharacter),
	)
}

func getArtistGroupKey(artists []spotify.SimpleArtist) string {
	return utils.JoinArtists(artists, config.ArtistJoinCharacter)
}
