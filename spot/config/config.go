package config

import (
	"flag"
	"fmt"
	"os"
)

const (
	defaultUserName                  = "drklump"
	defaultPort                      = 4000
	defaultAddress                   = "localhost"
	OutputTypeConsole                = "console"
	OutputTypePlaylist               = "playlist"
	CredentialsFlowClientCredentials = "client-credentials"
	CredentialsFlowRedirect          = "redirect"

	MetalPlaylistPattern = "^Metal ([0-9]+)"
	CacheFilename        = ".ignored/.cache.json"

	DiscoverWeeklyName          = "Discover Weekly"
	ReleaseRadarName            = "Release Radar"
	OutputDiscoveryPlaylistName = "Kristoffer's Total Discovery"

	MinimumAlbumTotalCount = 3
	AlbumChunkSize         = 20

	ArtistJoinCharacter = ","

	RedirectURLBase = "http://%s:%d/authenticate"
)

var (
	ClientID     = os.Getenv("SPOTIFY_ID")
	ClientSecret = os.Getenv("SPOTIFY_SECRET")

	UserName = ""

	DiscoveryPlaylistNames   = []string{DiscoverWeeklyName, ReleaseRadarName}
	DiscoveryPlaylistNameMap = map[string]string{
		DiscoverWeeklyName: DiscoverWeeklyName,
		ReleaseRadarName:   ReleaseRadarName,
	}

	FavouredPlaylistName       = ReleaseRadarName
	FavouredPlaylistAddedScore = 20

	WordPenaltyMap = map[string]int{
		"instrumental": -50,
		"acoustic":     -30,
		"re-imagined":  -30,
		"remix":        -30,
	}

	Port        = defaultPort
	Address     = defaultAddress
	RedirectURL = ""

	CredentialsFlow = CredentialsFlowClientCredentials

	OutputType = OutputTypeConsole
)

var usernameFlag = flag.String(
	"user",
	defaultUserName,
	"Spotify user name",
)

var outputTypeFlag = flag.String(
	"output-type",
	OutputTypeConsole,
	"The method. \"console\" or \"playlist\"",
)

var addressFlag = flag.String(
	"address",
	defaultAddress,
	"The address the server to run on",
)

var portFlag = flag.Int(
	"port",
	defaultPort,
	"The port for the server to listen on",
)

var credentialsFlowFlag = flag.String(
	"credentials-flow",
	CredentialsFlowClientCredentials,
	"The credentials flow to use. \"client-credentials\" or \"redirect\"",
)

func init() {
	flag.Parse()

	UserName = *usernameFlag
	OutputType = *outputTypeFlag
	Port = *portFlag
	Address = *addressFlag
	CredentialsFlow = *credentialsFlowFlag
	RedirectURL = fmt.Sprintf(RedirectURLBase, *addressFlag, *portFlag)
}
