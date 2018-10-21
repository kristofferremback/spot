package main

import (
	"github.com/kristofferostlund/spot/spot"
	"github.com/kristofferostlund/spot/spot/auth"
	"github.com/kristofferostlund/spot/spot/authserver"
	"github.com/kristofferostlund/spot/spot/config"

	"github.com/sirupsen/logrus"
)

func main() {
	if config.CredentialsFlow == config.CredentialsFlowRedirect {
		client, exists, err := auth.CachedRedirect(
			config.ClientID,
			config.ClientSecret,
			config.RedirectURL,
		)

		if exists && err == nil {
			spot.Run(client)

			return
		} else if exists && err != nil {
			logrus.Warnf("Failed to read token cache: %v", err)
		}

		authserver.Serve(spot.Run)

		return
	}

	client, err := auth.SpotifyClient(config.ClientID, config.ClientSecret)
	if err != nil {
		logrus.Error(err)

		return
	}

	spot.Run(client)
}
