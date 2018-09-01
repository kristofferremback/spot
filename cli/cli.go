package main

import (
	"fmt"

	"github.com/kristofferostlund/spot/spot"
	"github.com/kristofferostlund/spot/spot/auth"
	"github.com/kristofferostlund/spot/spot/config"
	"github.com/kristofferostlund/spot/spot/playlist"
	"github.com/kristofferostlund/spot/spot/server"
	"github.com/kristofferostlund/spot/spot/suggestion"

	"github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
)

func main() {
	if config.CredentialsFlow == config.CredentialsFlowRedirect {
		server.Serve(callback)

		return
	}

	client, err := auth.SpotifyClient(config.ClientID, config.ClientSecret)
	if err != nil {
		logrus.Error(err)

		return
	}

	callback(client)
}

func callback(client spotify.Client) {
	state, err := spot.GetState(client)
	if err != nil {
		logrus.Error(err)

		return
	}

	defer fmt.Printf("\n%s\n", suggestion.CreatePrintableTable(state.Suggestions))

	if config.OutputType == config.OutputTypePlaylist {
		remotePlaylist, err := playlist.SetRemotePlaylist(
			client,
			state.User,
			config.SpottedPlaylistName,
			suggestion.GetTracks(state.Suggestions),
		)
		if err != nil {
			logrus.Error(err)

			return
		}

		logrus.Infof("Set playlist %s with the suggested tracks.", remotePlaylist.Name)

		return
	}
}
