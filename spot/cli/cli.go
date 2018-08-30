package main

import (
	"fmt"

	"github.com/kristofferostlund/spot/spot/auth"
	"github.com/kristofferostlund/spot/spot/config"
	"github.com/kristofferostlund/spot/spot/playlist"
	"github.com/kristofferostlund/spot/spot/spotifytrack/fulltrack"
	"github.com/kristofferostlund/spot/spot/spotifyuser"
	"github.com/kristofferostlund/spot/spot/suggestion"
	"github.com/sirupsen/logrus"
)

func main() {
	client, err := auth.SpotifyClient(config.ClientID, config.ClientSecret)
	if err != nil {
		logrus.Error(err)

		return
	}

	user, err := spotifyuser.GetPublicProfile(client, config.UserName)
	if err != nil {
		logrus.Error(err)

		return
	}

	logrus.Infof("Fetching public playlists of user %s", config.UserName)

	currentPlaylists, err := playlist.GetMetalPlaylists(client, user)
	if err != nil {
		logrus.Error(err)

		return
	}

	uniqueTracks := fulltrack.GetUnique(playlist.FlattenTracks(currentPlaylists))

	logrus.Infof(
		"Playlist count: %3d, total track count: %3d",
		len(currentPlaylists),
		len(uniqueTracks),
	)

	discoveryPlaylists, err := playlist.GetDiscoveryPlaylists(client, user)
	if err != nil {
		logrus.Error(err)

		return
	}

	suggestions, err := suggestion.GetSuggestions(client, discoveryPlaylists, uniqueTracks)
	if err != nil {
		logrus.Error(err)

		return
	}

	fmt.Printf("\n%s\n", suggestion.CreatePrintableTable(suggestions))
}
