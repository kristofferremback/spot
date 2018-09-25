package spot

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"

	"github.com/kristofferostlund/spot/spot/config"
	"github.com/kristofferostlund/spot/spot/playlist"
	"github.com/kristofferostlund/spot/spot/spotifyrecommendation"
	"github.com/kristofferostlund/spot/spot/spotifytrack/fulltrack"
	"github.com/kristofferostlund/spot/spot/spotifyuser"
	"github.com/kristofferostlund/spot/spot/suggestion"
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
	case config.OperationTypeTrackRecommendations:
		recommend(client)

		break
	default:
		logrus.Errorf("Operation type %s is not a valid operation type", config.OperationType)

		break
	}
}

func recommend(client spotify.Client) {
	recommendations, err := getRecommendations(client)
	if err != nil {
		logrus.Error(err)

		return
	}

	defer fmt.Printf("\n%s\n", suggestion.CreatePrintableTable(recommendations.Suggestions))
}

func discover(client spotify.Client) {
	discovery, err := getDiscovery(client)
	if err != nil {
		logrus.Error(err)

		return
	}

	defer fmt.Printf("\n%s\n", suggestion.CreatePrintableTable(discovery.Suggestions))

	if config.OutputType == config.OutputTypePlaylist {
		remotePlaylist, err := playlist.SetRemotePlaylist(
			client,
			discovery.User,
			config.SpottedPlaylistName,
			suggestion.GetTracks(discovery.Suggestions),
		)
		if err != nil {
			logrus.Error(err)

			return
		}

		logrus.Infof("Set playlist %s with the suggested tracks.", remotePlaylist.Name)

		return
	}
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
