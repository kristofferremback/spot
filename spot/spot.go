package spot

import (
	"github.com/kristofferostlund/spot/spot/config"
	"github.com/kristofferostlund/spot/spot/playlist"
	"github.com/kristofferostlund/spot/spot/spotifytrack/fulltrack"
	"github.com/kristofferostlund/spot/spot/spotifyuser"
	"github.com/kristofferostlund/spot/spot/suggestion"
	"github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
)

type State struct {
	User               *spotify.User
	Playlists          []playlist.Playlist
	Tracks             []spotify.FullTrack
	DiscoveryPlaylists []playlist.Playlist
	Suggestions        []suggestion.Suggestion
}

func GetState(client spotify.Client) (State, error) {
	var err error
	state := State{}

	state.User, err = spotifyuser.GetPublicProfile(client, config.UserName)
	if err != nil {
		return state, err
	}

	logrus.Infof("Fetching public playlists of user %s", config.UserName)

	state.Playlists, err = playlist.GetMetalPlaylists(client, state.User)
	if err != nil {
		return state, err
	}

	state.Tracks = fulltrack.GetUnique(playlist.FlattenTracks(state.Playlists))

	logrus.Infof(
		"Playlist count: %3d, total track count: %3d",
		len(state.Playlists),
		len(state.Tracks),
	)

	state.DiscoveryPlaylists, err = playlist.GetDiscoveryPlaylists(client, state.User)
	if err != nil {
		return state, err
	}

	state.Suggestions, err = suggestion.GetSuggestions(client, state.DiscoveryPlaylists, state.Tracks)
	if err != nil {
		return state, err
	}

	return state, nil
}
