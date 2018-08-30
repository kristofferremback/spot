package playlist

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"

	"github.com/kristofferostlund/spot/spot/cache"
	"github.com/kristofferostlund/spot/spot/config"
	"github.com/sirupsen/logrus"

	"github.com/zmb3/spotify"
)

type Playlist struct {
	SimplePlaylist  spotify.SimplePlaylist
	Tracks          []spotify.FullTrack
	ID              spotify.ID
	Name            string
	SnapshotID      string
	TracksPopulated bool
}

func CreatePlaylist(simplePlaylist spotify.SimplePlaylist) Playlist {
	return Playlist{
		SimplePlaylist:  simplePlaylist,
		Tracks:          []spotify.FullTrack{},
		ID:              simplePlaylist.ID,
		Name:            simplePlaylist.Name,
		SnapshotID:      simplePlaylist.SnapshotID,
		TracksPopulated: false,
	}
}

func GetMetalPlaylists(client spotify.Client, user *spotify.User) ([]Playlist, error) {
	playlists := []Playlist{}
	cachedPlaylists := []Playlist{}

	if err := cache.ReadCache(config.CacheFilename, &cachedPlaylists); err != nil {
		return playlists, err
	}

	simplePlaylists, err := listSimplePlaylists(client, user)
	if err != nil {
		return playlists, err
	}

	for _, playlist := range filterByPattern(
		simplePlaylists,
		config.MetalPlaylistPattern,
	) {
		if cachedPlaylist, isCached := findPlaylist(cachedPlaylists, func(p Playlist) bool {
			return p.SnapshotID == playlist.SnapshotID
		}); isCached {
			playlist = cachedPlaylist
		}

		if !playlist.TracksPopulated {
			tracks, err := listTracks(client, user, playlist.SimplePlaylist)
			if err != nil {
				return playlists, err
			}

			playlist.Tracks = tracks
			playlist.TracksPopulated = true
		}

		playlists = append(playlists, playlist)
	}

	if err := cache.WriteCache(config.CacheFilename, playlists); err != nil {
		return playlists, err
	}

	return playlists, nil
}

func GetDiscoveryPlaylists(client spotify.Client, user *spotify.User) ([]Playlist, error) {
	discoveryPlaylists := []Playlist{}

	simplePlaylists, err := listSimplePlaylists(client, user)
	if err != nil {
		return discoveryPlaylists, err
	}

	for _, playlist := range simplePlaylists {
		if _, ok := config.DiscoveryPlaylistNameMap[playlist.Name]; ok {
			tracks, err := listTracks(client, user, playlist)
			if err != nil {
				return discoveryPlaylists, err
			}

			discoveryPlaylist := CreatePlaylist(playlist)

			discoveryPlaylist.Tracks = tracks
			discoveryPlaylist.TracksPopulated = true

			discoveryPlaylists = append(discoveryPlaylists, discoveryPlaylist)
		}
	}

	return discoveryPlaylists, nil
}

func FlattenTracks(playlists []Playlist) []spotify.FullTrack {
	tracks := []spotify.FullTrack{}

	for _, playlist := range playlists {
		for _, track := range playlist.Tracks {
			tracks = append(tracks, track)
		}
	}

	return tracks
}

func listSimplePlaylists(client spotify.Client, user *spotify.User) ([]spotify.SimplePlaylist, error) {
	pageLimit := 50
	totalCount := -1
	playlists := []spotify.SimplePlaylist{}

	for totalCount != len(playlists) {
		offset := len(playlists)
		options := &spotify.Options{Limit: &pageLimit, Offset: &offset}

		page, err := client.GetPlaylistsForUserOpt(user.ID, options)
		if err != nil {
			errorMessage := "Error listing simple playlists for user %s: %v"
			return playlists, fmt.Errorf(errorMessage, user.DisplayName, err)
		}

		totalCount = page.Total
		playlists = append(playlists, page.Playlists...)
	}

	return playlists, nil
}

func listTracks(
	client spotify.Client,
	user *spotify.User,
	simplePlaylist spotify.SimplePlaylist,
) ([]spotify.FullTrack, error) {
	pageLimit := 100
	totalCount := -1
	tracks := []spotify.FullTrack{}

	logrus.Debugf("Listing tracks for playlist %s", simplePlaylist.Name)

	for totalCount != len(tracks) {
		offset := len(tracks)
		options := &spotify.Options{Limit: &pageLimit, Offset: &offset}

		page, err := client.GetPlaylistTracksOpt(
			user.ID,
			simplePlaylist.ID,
			options,
			"",
		)
		if err != nil {
			errorMessage := "Failed to get playlist track for simple playlist %s: %v"
			return tracks, fmt.Errorf(errorMessage, simplePlaylist.Name, err)
		}

		totalCount = page.Total

		for _, track := range page.Tracks {
			tracks = append(tracks, track.Track)
		}
	}

	logrus.Infof(
		"Listed %4v tracks for playlist %s",
		len(tracks),
		simplePlaylist.Name,
	)

	return tracks, nil
}

func getPlaylistNumber(re *regexp.Regexp, name string) int {
	value, err := strconv.Atoi(re.FindStringSubmatch(name)[1])

	if err != nil {
		return -1
	}

	return value
}

func filterByPattern(
	simplePlaylists []spotify.SimplePlaylist,
	pattern string,
) []Playlist {
	playlists := []Playlist{}
	re := regexp.MustCompile(pattern)

	for _, simplePlaylist := range simplePlaylists {
		if re.MatchString(simplePlaylist.Name) {
			playlists = append(playlists, CreatePlaylist(simplePlaylist))
		}
	}

	sort.Slice(playlists, func(i, j int) bool {
		current := getPlaylistNumber(re, playlists[i].Name)
		next := getPlaylistNumber(re, playlists[j].Name)

		return current > next
	})

	return playlists
}

func findPlaylist(playlists []Playlist, compare func(Playlist) bool) (Playlist, bool) {
	for _, playlist := range playlists {
		if compare(playlist) {
			return playlist, true
		}
	}

	return Playlist{}, false
}

func findSimplePlaylist(
	simplePlaylists []spotify.SimplePlaylist,
	compare func(spotify.SimplePlaylist) bool,
) (spotify.SimplePlaylist, bool) {
	for _, playlist := range simplePlaylists {
		if compare(playlist) {
			return playlist, true
		}
	}

	return spotify.SimplePlaylist{}, false
}
