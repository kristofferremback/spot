package auth

import (
	"context"
	"fmt"

	"github.com/kristofferostlund/spot/spot/utils"
	"github.com/satori/go.uuid"

	"github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

func SpotifyClient(clientID, clientSecret string) (spotify.Client, error) {
	var client = spotify.Client{}

	logrus.Debug("Creating Spotify client")

	config := &clientcredentials.Config{
		ClientID:     clientID,
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

func RedirectAuthenticator(clientID, clientSecret, redirectURL string) (spotify.Authenticator, string) {
	authenticator := spotify.NewAuthenticator(
		redirectURL,
		spotify.ScopeUserReadPrivate,
		spotify.ScopePlaylistReadPrivate,
		spotify.ScopePlaylistModifyPrivate,
		spotify.ScopePlaylistModifyPublic,
	)

	authenticator.SetAuthInfo(clientID, clientSecret)

	return authenticator, uuid.Must(uuid.NewV4()).String()
}

func OpenAuthURL(authenticator spotify.Authenticator, state string) {
	utils.OpenBrowser(authenticator.AuthURL(state))
}

func RedirectClient(authenticator spotify.Authenticator, token *oauth2.Token) spotify.Client {
	client := authenticator.NewClient(token)

	client.AutoRetry = true

	return client
}
