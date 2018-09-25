package spot

import (
	"fmt"
	"math"
	"regexp"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"

	"github.com/kristofferostlund/spot/spot/config"
	"github.com/kristofferostlund/spot/spot/playlist"
	"github.com/kristofferostlund/spot/spot/spotifyrecommendation"
	"github.com/kristofferostlund/spot/spot/spotifytrack/fulltrack"
	"github.com/kristofferostlund/spot/spot/spotifyuser"
	"github.com/kristofferostlund/spot/spot/suggestion"
	"github.com/kristofferostlund/spot/spot/utils"
)

type State struct {
	User      *spotify.User
	Playlists []playlist.Playlist
	Tracks    []spotify.FullTrack
}

type Discovery struct {
	User               *spotify.User
	Playlists          []playlist.Playlist
	Tracks             []spotify.FullTrack
	DiscoveryPlaylists []playlist.Playlist
	Suggestions        []suggestion.Suggestion
}

type Recommendation struct {
	User        *spotify.User
	Playlists   []playlist.Playlist
	Tracks      []spotify.FullTrack
	Suggestions []suggestion.Suggestion
}

func Run(client spotify.Client) {
	switch config.OperationType {
	case config.OperationTypeDiscovery:
		discover(client)

		break
	case config.OperationTypeRecommendations:
		recommend(client)

		break
	case config.OperationTypeCheckTrackExists:
		checkTrackExists(client)

		break
	case config.OperationTypeCheckPlaylistHoles:
		checkPlaylistHoles(client)

		break
	default:
		logrus.Errorf("Operation type %s is not a valid operation type", config.OperationType)

		break
	}
}

func checkTrackExists(client spotify.Client) {
	status, err := client.PlayerCurrentlyPlaying()
	if err != nil {
		logrus.Error(err)

		return
	}

	if !status.Playing {
		logrus.Warn("User doesn't seem to listen to spotify currently")

		return
	}

	logrus.Infof(
		"User is listening to %s by %s. Checking if it's new",
		status.Item.Name,
		utils.JoinArtists(status.Item.Artists, ", "),
	)

	currentTrack, err := fulltrack.Get(client, status.Item.ID)
	if err != nil {
		logrus.Error(err)

		return
	}

	state, err := getState(client)
	if err != nil {
		logrus.Error(err)

		return
	}

	foundPlaylist, exists := playlist.FindPlaylistByTrack(state.Playlists, currentTrack)
	if exists {
		logrus.Infof("The track already on playlist %s", foundPlaylist.Name)

		return
	}

	logrus.Infof("The track is new, quite amazing I'd say!")
}

func checkPlaylistHoles(client spotify.Client) {
	numbers := []int{}
	holes := []int{}

	state, err := getState(client)
	if err != nil {
		logrus.Error(err)

		return
	}

	for index, list := range state.Playlists {
		pattern := regexp.MustCompile("Metal 0*(\\d+)")

		numberString := pattern.FindStringSubmatch(list.Name)[1]
		value, err := strconv.Atoi(numberString)

		if err != nil {
			logrus.Warnf("Failed to parse %s: %v", numberString, err)

			continue
		}

		numbers = append(numbers, value)

		if index > 0 && math.Abs(float64(numbers[index-1]-value)) != 1.0 {
			holes = append(holes, value+1)
		}
	}

	for _, hole := range holes {
		logrus.Infof("Found a potential hole at %d", hole)
	}
}

func recommend(client spotify.Client) {
	recommendations, err := getRecommendations(client)
	if err != nil {
		logrus.Error(err)

		return
	}

	defer fmt.Printf("\n%s\n", suggestion.CreatePrintableTable(recommendations.Suggestions))

	if config.OutputType == config.OutputTypePlaylist {
		createPlaylist(
			client,
			recommendations.User,
			config.SpottedRecommendationsPlaylistName,
			suggestion.GetTracks(recommendations.Suggestions),
		)
	}
}

func discover(client spotify.Client) {
	discovery, err := getDiscovery(client)
	if err != nil {
		logrus.Error(err)

		return
	}

	defer fmt.Printf("\n%s\n", suggestion.CreatePrintableTable(discovery.Suggestions))

	if config.OutputType == config.OutputTypePlaylist {
		createPlaylist(
			client,
			discovery.User,
			config.SpottedDiscoveryPlaylistName,
			suggestion.GetTracks(discovery.Suggestions),
		)
	}
}

func createPlaylist(
	client spotify.Client,
	user *spotify.User,
	name string,
	tracks []spotify.FullTrack,
) {
	remotePlaylist, err := playlist.SetRemotePlaylist(client, user, name, tracks)
	if err != nil {
		logrus.Error(err)

		return
	}

	logrus.Infof("Set playlist %s with the suggested tracks.", remotePlaylist.Name)

	return
}

func getState(client spotify.Client) (State, error) {
	state := State{}
	var err error

	if config.CredentialsFlow == config.CredentialsFlowRedirect {
		state.User, err = spotifyuser.GetCurrentUser(client)
	} else {
		state.User, err = spotifyuser.GetPublicProfile(client, config.UserName)
	}

	if err != nil {
		return state, err
	}

	logrus.Infof("Fetching playlists of user %s", state.User.ID)

	state.Playlists, err = playlist.GetPlaylistsMatchingPattern(
		client,
		state.User,
		config.PlaylistNamePattern,
	)
	if err != nil {
		return state, err
	}

	state.Tracks = fulltrack.GetUnique(playlist.FlattenTracks(state.Playlists))

	logrus.Infof(
		"Playlist count: %3d, total track count: %3d",
		len(state.Playlists),
		len(state.Tracks),
	)

	return state, nil
}

func getDiscovery(client spotify.Client) (Discovery, error) {
	state := State{}
	discovery := Discovery{}
	var err error

	state, err = getState(client)
	if err != nil {
		return discovery, err
	}

	discovery = Discovery{
		User:      state.User,
		Playlists: state.Playlists,
		Tracks:    state.Tracks,
	}

	discovery.DiscoveryPlaylists, err = playlist.GetDiscoveryPlaylists(client, discovery.User)
	if err != nil {
		return discovery, err
	}

	discovery.Suggestions, err = suggestion.GetSuggestions(client, discovery.DiscoveryPlaylists, discovery.Tracks)
	if err != nil {
		return discovery, err
	}

	return discovery, nil
}

func getRecommendations(client spotify.Client) (Recommendation, error) {
	state := State{}
	recommendations := Recommendation{}
	var err error

	state, err = getState(client)
	if err != nil {
		return recommendations, err
	}

	recommendations = Recommendation{
		User:      state.User,
		Playlists: state.Playlists,
		Tracks:    state.Tracks,
	}

	recommendedTracks, err := spotifyrecommendation.Recommend(client)
	if err != nil {
		return recommendations, err
	}

	recommendations.Suggestions, err = suggestion.GetSuggestionsFromTracks(
		client,
		recommendedTracks,
		recommendations.Tracks,
	)
	if err != nil {
		return recommendations, err
	}

	return recommendations, nil
}
