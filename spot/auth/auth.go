package auth

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2/clientcredentials"
)

func SpotifyClient(clientId, clientSecret string) (spotify.Client, error) {
	var client = spotify.Client{}

	logrus.Debug("Creating Spotify client")

	config := &clientcredentials.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		TokenURL:     spotify.TokenURL,
	}

	token, err := config.Token(context.Background())
	if err != nil {
		return client, fmt.Errorf("Failed to create spotify client: %v", err)
	}

	logrus.Info("Spotify client successfully authenticated")

	client = spotify.Authenticator{}.NewClient(token)

	client.AutoRetry = true

	return client, nil
}
