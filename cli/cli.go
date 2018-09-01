package main

import (
	"github.com/kristofferostlund/spot/spot"
	"github.com/kristofferostlund/spot/spot/auth"
	"github.com/kristofferostlund/spot/spot/config"
	"github.com/kristofferostlund/spot/spot/server"

	"github.com/sirupsen/logrus"
)

func main() {
	if config.CredentialsFlow == config.CredentialsFlowRedirect {
		server.Serve(spot.Run)

		return
	}

	client, err := auth.SpotifyClient(config.ClientID, config.ClientSecret)
	if err != nil {
		logrus.Error(err)

		return
	}

	spot.Run(client)
}
