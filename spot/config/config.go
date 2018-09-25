package config

import (
	"flag"
	"fmt"
	"os"
	"time"
)

const (
	defaultUserName = "drklump"
	defaultPort     = 4000
	defaultAddress  = "localhost"

	OutputTypeConsole                = "console"
	OutputTypePlaylist               = "playlist"
	CredentialsFlowClientCredentials = "client-credentials"
	CredentialsFlowRedirect          = "redirect"

	OperationTypeDiscovery            = "discovery"
	OperationTypeTrackRecommendations = "recommendation"
	OperationTypeCheckTrackExists     = "check-track"
	OperationTypeCheckPlaylistHoles   = "check-playlist-holes"

	CountrySweden = "SE"

	defaultPlaylistPattern = "^Metal ([0-9]+)"
	CacheFilename          = ".ignored/.cache.json"
	TokenCacheFilename     = ".ignored/.token-cache.json"

	DiscoverWeeklyName = "Discover Weekly"
	ReleaseRadarName   = "Release Radar"

	MinimumAlbumTotalCount = 3
	AlbumChunkSize         = 20

	NumberPaddingSize = 20

	ArtistJoinCharacter = ","

	spottedPlaylistBase = "Spotted™ %s"
	redirectURLBase     = "http://%s:%d/authenticate"
)

var (
	ClientID     = os.Getenv("SPOTIFY_ID")
	ClientSecret = os.Getenv("SPOTIFY_SECRET")

	UserName = defaultUserName

	PlaylistNamePattern = defaultPlaylistPattern

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

	Port                = defaultPort
	Address             = defaultAddress
	RedirectURL         = ""
	SpottedPlaylistName = fmt.Sprintf(spottedPlaylistBase, time.Now().Format("2006-01-02"))

	CredentialsFlow = CredentialsFlowClientCredentials

	OutputType    = OutputTypeConsole
	OperationType = OperationTypeDiscovery

	Country = CountrySweden
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

var playlistNamePatternFlag = flag.String(
	"playlist-pattern",
	defaultPlaylistPattern,
	"The playlist name pattern to use as base",
)

var operationFlag = flag.String(
	"operation",
	OperationTypeDiscovery,
	"The operation to perform",
)

var countryFlag = flag.String(
	"country",
	CountrySweden,
	"The country to base recommendations on. Example: SE",
)

func init() {
	flag.Parse()

	UserName = *usernameFlag
	OutputType = *outputTypeFlag
	Port = *portFlag
	Address = *addressFlag
	CredentialsFlow = *credentialsFlowFlag
	RedirectURL = fmt.Sprintf(redirectURLBase, *addressFlag, *portFlag)
	PlaylistNamePattern = *playlistNamePatternFlag
	OperationType = *operationFlag
	Country = *countryFlag
}
