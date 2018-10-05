package suggestion

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"

	"github.com/kristofferostlund/spot/spot/config"
	"github.com/kristofferostlund/spot/spot/fullalbum"
	"github.com/kristofferostlund/spot/spot/playlist"
	"github.com/kristofferostlund/spot/spot/spotifytrack"
	"github.com/kristofferostlund/spot/spot/spotifytrack/fulltrack"
	"github.com/kristofferostlund/spot/spot/utils"
)

type Suggestion struct {
	Playlist  playlist.Playlist
	Track     spotify.FullTrack
	Album     spotify.FullAlbum
	Relevance int
}

func (s *Suggestion) CalculateRelevance(tracksByArtist spotifytrack.ArtistFullTrackMap) {
	s.Relevance = 0

	modifiers := []int{
		s.Album.ReleaseDateTime().Year() - 2000,
		fulltrack.GetTrackCountByArtists(tracksByArtist, s.Track.Artists),
	}

	for word, penalty := range config.WordPenaltyMap {
		if strings.Contains(strings.ToLower(s.Track.Name), word) {
			modifiers = append(modifiers, penalty)
		}
	}

	if s.Playlist.Name == config.FavouredPlaylistName {
		modifiers = append(modifiers, config.FavouredPlaylistAddedScore)
	}

	for _, score := range modifiers {
		s.Relevance += score
	}
}

func CreateSuggestion(
	client spotify.Client,
	originPlaylist playlist.Playlist,
	track spotify.FullTrack,
) (Suggestion, error) {
	suggestion := Suggestion{Playlist: originPlaylist}

	album, err := fullalbum.GetAlbumByTrack(client, track)
	if err != nil {
		return suggestion, nil
	}

	if string(track.Album.ID) != string(album.ID) {
		for _, albumTrack := range album.Tracks.Tracks {
			if albumTrack.Name == track.Name {
				track, err = fulltrack.Get(client, albumTrack.ID)
				if err != nil {
					return suggestion, err
				}

				break
			}
		}
	}

	suggestion.Track = track
	suggestion.Album = album
	suggestion.CalculateRelevance(spotifytrack.ArtistFullTrackMap{})

	return suggestion, nil
}

func GetSuggestions(
	client spotify.Client,
	discoveryPlaylists []playlist.Playlist,
	existingTracks []spotify.FullTrack,
) ([]Suggestion, error) {
	trackMap := fulltrack.CreateMap(existingTracks)
	tracksByArtist := fulltrack.GroupByArtists(existingTracks)

	suggestions := []Suggestion{}

	logrus.Info("Generating suggestions")

	for _, discoveryPlaylist := range discoveryPlaylists {
		for _, track := range discoveryPlaylist.Tracks {
			if !fulltrack.InMap(trackMap, track) {
				suggestion, err := CreateSuggestion(client, discoveryPlaylist, track)
				if err != nil {
					return suggestions, err
				}

				if suggestion.Album.Tracks.Total > config.MinimumAlbumTotalCount {
					suggestion.CalculateRelevance(tracksByArtist)

					suggestions = append(suggestions, suggestion)
				}
			}
		}
	}

	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Relevance > suggestions[j].Relevance
	})

	logrus.Infof("Successfully generated %d suggestions", len(suggestions))

	return suggestions, nil
}

func GetSuggestionsFromTracks(
	client spotify.Client,
	baseTracks []spotify.FullTrack,
	existingTracks []spotify.FullTrack,
) ([]Suggestion, error) {
	trackMap := fulltrack.CreateMap(existingTracks)
	tracksByArtist := fulltrack.GroupByArtists(existingTracks)

	suggestions := []Suggestion{}

	logrus.Info("Generating suggestions")

	for _, track := range baseTracks {
		if !fulltrack.InMap(trackMap, track) {
			suggestion, err := CreateSuggestion(client, playlist.Playlist{}, track)
			if err != nil {
				return suggestions, err
			}

			if suggestion.Album.Tracks.Total > config.MinimumAlbumTotalCount {
				suggestion.CalculateRelevance(tracksByArtist)

				suggestions = append(suggestions, suggestion)
			}
		}
	}

	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Relevance > suggestions[j].Relevance
	})

	logrus.Infof("Successfully generated %d suggestions", len(suggestions))

	return suggestions, nil
}

func CreatePrintableTable(suggestions []Suggestion) string {
	output := fmt.Sprintf(
		"   %s %s %s %s %s %s %s\n",
		utils.FixedWidthString("Name", 30),
		utils.FixedWidthString("Playlist", 30),
		utils.FixedWidthString("Artist(s)", 30),
		utils.FixedWidthString("Album", 30),
		utils.FixedWidthString("Year", 6),
		utils.FixedWidthString("Score", 5),
		utils.FixedWidthString("Spotify URI", 36),
	)

	for index, s := range suggestions {
		output += fmt.Sprintf(
			"%-02d %s %s %s %s %-6d %-5d %36s\n",
			index+1,
			utils.FixedWidthString(s.Track.Name, 30),
			utils.FixedWidthString(s.Playlist.Name, 30),
			utils.FixedWidthString(utils.JoinArtists(s.Track.Artists, ", "), 30),
			utils.FixedWidthString(s.Album.Name, 30),
			s.Album.ReleaseDateTime().Year(),
			s.Relevance,
			s.Track.URI,
		)
	}

	return output
}

func GetTracks(suggestions []Suggestion) []spotify.FullTrack {
	tracks := []spotify.FullTrack{}

	for _, suggestion := range suggestions {
		tracks = append(tracks, suggestion.Track)
	}

	return tracks
}
