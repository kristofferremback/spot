package spotifyuser

import (
	"fmt"

	"github.com/zmb3/spotify"
)

func GetPublicProfile(client spotify.Client, username string) (*spotify.User, error) {
	var user *spotify.User

	user, err := client.GetUsersPublicProfile(spotify.ID(username))
	if err != nil {
		return user, fmt.Errorf("Failed to get Spotify user %s: %v", username, err)
	}

	return user, nil
}

func GetCurrentUser(client spotify.Client) (*spotify.User, error) {
	var user *spotify.User

	privateUser, err := client.CurrentUser()
	if err != nil {
		return user, fmt.Errorf("Failed to get current spotify user user: %v", err)
	}

	return &privateUser.User, nil
}
