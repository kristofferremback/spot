package spotifyrecommendation

import (
	"fmt"
	"math"

	"github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"

	"github.com/kristofferostlund/spot/spot/config"
	"github.com/kristofferostlund/spot/spot/fullalbum"
	"github.com/kristofferostlund/spot/spot/spotifytrack/fulltrack"
	"github.com/kristofferostlund/spot/spot/utils"
)

type RecommendationParameters struct {
	Seeds           spotify.Seeds
	TrackAttributes *spotify.TrackAttributes
	FromYear        int
	MinTrackCount   int
}

func Recommend(client spotify.Client) ([]spotify.FullTrack, error) {
	tracks := []spotify.FullTrack{}
	pageLimit := 5

	userTopArtists, err := client.CurrentUsersTopArtistsOpt(&spotify.Options{Limit: &pageLimit})
	if err != nil {
		return tracks, fmt.Errorf("Failed to get user's top artists: %v", err)
	}

	userTopTracks, err := client.CurrentUsersTopTracks()
	if err != nil {
		return tracks, fmt.Errorf("Failed to get user's top tracks: %v", err)
	}

	trackAttributes, err := getTrackAttributes(client, userTopTracks.Tracks)
	if err != nil {
		return tracks, err
	}

	for _, artist := range userTopArtists.Artists {
		logrus.Infof("Fetching recommendations seeded by artist %s", artist.Name)

		params := RecommendationParameters{
			FromYear:      2016,
			MinTrackCount: 100,
			Seeds: spotify.Seeds{
				Artists: []spotify.ID{artist.ID},
			},
			TrackAttributes: trackAttributes,
		}

		pageTracks, err := getRecommendedTracks(client, params)
		if err != nil {
			return tracks, err
		}

		logrus.Infof("Fetched %d recommendations seeded by artist %s", len(pageTracks), artist.Name)

		tracks = append(tracks, pageTracks...)
	}

	return tracks, nil
}

func getRecommendedTracks(client spotify.Client, params RecommendationParameters) ([]spotify.FullTrack, error) {
	pageLimit := 100
	trackCount := 0
	totalCount := 0
	tracks := []spotify.FullTrack{}

	options := spotify.Options{
		Limit:   &pageLimit,
		Offset:  &totalCount,
		Country: &config.Country,
	}

	page, err := client.GetRecommendations(params.Seeds, params.TrackAttributes, &options)
	if err != nil {
		return tracks, fmt.Errorf("Failed to get recommendations: %v", err)
	}

	totalCount += len(page.Tracks)

	fullTracks, err := fulltrack.GetMany(client, utils.GetSpotifyIDs(page.Tracks))
	if err != nil {
		return tracks, err
	}

	for _, track := range fullTracks {
		album, err := fullalbum.Get(client, track.Album.ID)
		if err != nil {
			return tracks, err
		}

		if album.ReleaseDateTime().Year() >= params.FromYear {
			trackCount++
			tracks = append(tracks, track)
		}
	}

	return tracks, nil
}

func getTrackAttributes(client spotify.Client, tracks []spotify.FullTrack) (*spotify.TrackAttributes, error) {
	var attributes *spotify.TrackAttributes

	features, err := client.GetAudioFeatures(utils.GetSpotifyIDs(tracks)...)
	if err != nil {
		return attributes, fmt.Errorf(
			"Failed to get audio features of %d track(s): %v",
			len(tracks),
			err,
		)
	}

	acousticness := []float64{}
	instrumentalness := []float64{}
	liveness := []float64{}
	energy := []float64{}
	valence := []float64{}

	for _, feature := range features {
		acousticness = append(acousticness, float64(feature.Acousticness))
		instrumentalness = append(instrumentalness, float64(feature.Instrumentalness))
		liveness = append(liveness, float64(feature.Liveness))
		energy = append(energy, float64(feature.Energy))
		valence = append(valence, float64(feature.Valence))
	}

	averageAcousticness := utils.AverageFloat(acousticness)
	averageInstrumentalness := utils.AverageFloat(instrumentalness)
	averageLiveness := utils.AverageFloat(liveness)
	averageEnergy := utils.AverageFloat(energy)
	averageValence := utils.AverageFloat(valence)

	attributes = spotify.NewTrackAttributes().
		MaxAcousticness(math.Min(averageAcousticness+.3, 0.8)).
		MaxInstrumentalness(math.Min(averageInstrumentalness+.3, 0.8)).
		MaxLiveness(math.Min(averageLiveness+.3, 0.8)).
		MinEnergy(math.Max(averageEnergy-.3, 0.3)).
		MaxValence(math.Min(averageValence+.3, 0.8))

	return attributes, nil
}
