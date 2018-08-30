package config

import (
	"flag"
	"os"
)

const (
	DefaultUserName = "drklump"

	MetalPlaylistPattern = "^Metal ([0-9]+)"
	CacheFilename        = ".ignored/.cache.json"

	DiscoverWeeklyName = "Discover Weekly"
	ReleaseRadarName   = "Release Radar"

	MinimumAlbumTotalCount = 3
	AlbumChunkSize         = 20

	ArtistJoinCharacter = ","
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
)

var usernameFlag = flag.String(
	"user",
	DefaultUserName,
	"Spotify user name",
)

func init() {
	flag.Parse()

	UserName = *usernameFlag
}
