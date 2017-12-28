package cognitiveservices

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v1.0/videosearch"
	"github.com/Azure/go-autorest/autorest"
)

func getVideoSearchClient(accountName string) videosearch.VideosClient {
	apiKey := getFirstKey(accountName)
	videoSearchClient := videosearch.NewVideosClient()
	csAuthorizer := autorest.NewCognitiveServicesAuthorizer(apiKey)
	videoSearchClient.Authorizer = csAuthorizer
	return videoSearchClient
}

//SearchVideos returns a list of videos
func SearchVideos(accountName string) (videosearch.Videos, error) {
	videoSearchClient := getVideoSearchClient(accountName)
	query := "Nasa CubeSat"

	videos, err := videoSearchClient.Search(
		context.Background(),
		"",
		query,
		"",
		"",
		"",
		"",
		"",
		"",
		nil,
		"",
		"",
		"",
		"",
		nil,
		"",
		"",
		"",
		"",
		nil,
		"")

	return videos, err
}

//TrendingVideos returns the videos that are trending
func TrendingVideos(accountName string) (videosearch.TrendingVideos, error) {
	videoSearchClient := getVideoSearchClient(accountName)
	trendingVideos, err := videoSearchClient.Trending(context.Background(), "", "", "", "", "", "", "", "", "", "", nil, "")
	return trendingVideos, err
}
