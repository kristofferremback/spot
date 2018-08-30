package spotifyuser

import (
	"fmt"

	"github.com/zmb3/spotify"
)

func GetPublicProfile(client spotify.Client, username string) (*spotify.User, error) {
	var user *spotify.User

	user, err := client.GetUsersPublicProfile(spotify.ID(username))
	if err != nil {
		return user, fmt.Errorf("Failed to get Spotify user: %v", err)
	}

	return user, nil
}
