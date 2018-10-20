package playlist

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"

	"github.com/kristofferostlund/spot/spot/spotifytrack/fulltrack"

	"github.com/kristofferostlund/spot/spot/cache"
	"github.com/kristofferostlund/spot/spot/config"
	"github.com/kristofferostlund/spot/spot/utils"
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

func GetPlaylistsMatchingPattern(client spotify.Client, user *spotify.User, pattern string) ([]Playlist, error) {
	playlists := []Playlist{}
	cachedPlaylists := []Playlist{}

	if err := cache.ReadCache(config.CacheFilename, &cachedPlaylists); err != nil {
		return playlists, err
	}

	simplePlaylists, err := listSimplePlaylists(client, user)
	if err != nil {
		return playlists, err
	}

	for _, playlist := range filterByPatternWithIgnored(simplePlaylists, pattern) {
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

func SetRemotePlaylist(
	client spotify.Client,
	user *spotify.User,
	name string,
	tracks []spotify.FullTrack,
) (Playlist, error) {
	remotePlaylist := Playlist{}
	var err error

	playlists, err := listSimplePlaylists(client, user)
	if err != nil {
		return remotePlaylist, err
	}

	if foundPlaylist, exists := findSimplePlaylist(
		playlists, func(p spotify.SimplePlaylist) bool { return p.Name == name },
	); exists {
		remotePlaylist = CreatePlaylist(foundPlaylist)

		remotePlaylist, err = truncatePlaylist(client, user, remotePlaylist)
		if err != nil {
			return remotePlaylist, err
		}
	} else {
		created, err := createPlaylist(client, user, name)
		if err != nil {
			return remotePlaylist, err
		}

		remotePlaylist = CreatePlaylist(created.SimplePlaylist)
	}

	return addTracks(client, remotePlaylist, tracks)
}

func FindPlaylistByTrack(playlists []Playlist, track spotify.FullTrack) (Playlist, bool) {
	return findPlaylist(playlists, func(playlist Playlist) bool {
		trackMap := fulltrack.CreateMap(playlist.Tracks)

		return fulltrack.InMap(trackMap, track)
	})
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

	sort.Slice(playlists, func(i, j int) bool {
		current := utils.MakeStringSortable(playlists[i].Name, config.NumberPaddingSize)
		next := utils.MakeStringSortable(playlists[j].Name, config.NumberPaddingSize)

		return current > next
	})

	return playlists, nil
}

func listTracks(
	client spotify.Client,
	user *spotify.User,
	simplePlaylist spotify.SimplePlaylist,
) ([]spotify.FullTrack, error) {
	pageLimit := 100
	totalCount := -1
	totalAttempts := 0
	tracks := []spotify.FullTrack{}
	var maxAttempts int

	logrus.Debugf("Listing tracks for playlist %s", simplePlaylist.Name)

	for totalCount != len(tracks) {
		totalAttempts++

		offset := len(tracks)
		options := &spotify.Options{Limit: &pageLimit, Offset: &offset}

		page, err := client.GetPlaylistTracksOpt(
			simplePlaylist.ID,
			options,
			"",
		)
		if err != nil {
			errorMessage := "Failed to get playlist track for simple playlist %s: %v"
			return tracks, fmt.Errorf(errorMessage, simplePlaylist.Name, err)
		}

		totalCount = page.Total
		maxAttempts = totalCount/pageLimit + 2

		if totalAttempts == maxAttempts {
			logrus.Warnf(
				"Reached the estimated request limit for %s. Fetched %d/%d tracks on %d/%d requests.",
				simplePlaylist.Name,
				len(tracks),
				totalCount,
				totalAttempts,
				maxAttempts,
			)

			break
		}

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

func filterByPatternWithIgnored(
	simplePlaylists []spotify.SimplePlaylist,
	pattern string,
) []Playlist {
	playlists := []Playlist{}
	re := regexp.MustCompile(pattern)

	for _, simplePlaylist := range simplePlaylists {
		if simplePlaylist.Name == config.DiscoverWeeklyName ||
			simplePlaylist.Name == config.ReleaseRadarName {
			continue
		}

		if re.MatchString(simplePlaylist.Name) {
			playlists = append(playlists, CreatePlaylist(simplePlaylist))
		}
	}

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

func createPlaylist(client spotify.Client, user *spotify.User, name string) (spotify.FullPlaylist, error) {
	fullPlaylist, err := client.CreatePlaylistForUser(user.ID, name, fmt.Sprintf("Autogenerated playlist by Spot"), true)
	if err != nil {
		return spotify.FullPlaylist{}, fmt.Errorf("Failed to create playlist %s: %v", name, err)
	}

	logrus.Infof("Successfully created playlist %s", fullPlaylist.Name)

	return *fullPlaylist, nil
}

func truncatePlaylist(client spotify.Client, user *spotify.User, playlist Playlist) (Playlist, error) {
	tracks := playlist.Tracks
	var err error

	if !playlist.TracksPopulated {
		tracks, err = listTracks(client, user, playlist.SimplePlaylist)
		if err != nil {
			return playlist, err
		}
	}

	if len(tracks) == 0 {
		return playlist, nil
	}

	playlist.SnapshotID, err = client.RemoveTracksFromPlaylist(
		playlist.ID,
		utils.GetSpotifyIDs(tracks)...,
	)
	if err != nil {
		return playlist, fmt.Errorf("Failed to truncate playlist %s: %v", playlist.Name, err)
	}

	logrus.Infof("Successfully truncated playlist %s", playlist.Name)

	return playlist, nil
}

func addTracks(client spotify.Client, playlist Playlist, tracks []spotify.FullTrack) (Playlist, error) {
	var err error

	chunks := utils.ChunkIDs(utils.GetSpotifyIDs(tracks), 100)

	for _, trackIDs := range chunks {
		playlist.SnapshotID, err = client.AddTracksToPlaylist(
			playlist.ID,
			trackIDs...,
		)

		if err != nil {
			return playlist, fmt.Errorf("Failed to add tracks to playlist %s: %v", playlist.Name, err)
		}
	}

	logrus.Infof("Successfully added %d tracks to playlist %s", len(tracks), playlist.Name)

	return playlist, nil
}
