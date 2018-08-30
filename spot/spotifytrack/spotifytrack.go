package spotifytrack

import (
	"github.com/zmb3/spotify"
)

type FullTrackMap map[string]spotify.FullTrack
type ArtistFullTrackMap map[string][]spotify.FullTrack

type SimpleTrackMap map[string]spotify.SimpleTrack
type ArtistSimpleTrackMap map[string][]spotify.SimpleTrack
