// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cognitiveservices

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v1.0/videosearch"
	"github.com/Azure/go-autorest/autorest"
)

func getVideoSearchClient(accountName string) videosearch.VideosClient {
	apiKey := getFirstKey(accountName)
	videoSearchClient := videosearch.NewVideosClient()
	csAuthorizer := autorest.NewCognitiveServicesAuthorizer(apiKey)
	videoSearchClient.Authorizer = csAuthorizer
	videoSearchClient.AddToUserAgent(helpers.UserAgent())
	return videoSearchClient
}

//SearchVideos returns a list of videos
func SearchVideos(accountName string) (videosearch.Videos, error) {
	videoSearchClient := getVideoSearchClient(accountName)
	query := "Nasa CubeSat"

	videos, err := videoSearchClient.Search(
		context.Background(), // context
		"",                   // X-BingApis-SDK header
		query,                // query keyword
		"",                   // Accept-Language header
		"",                   // User-Agent header
		"",                   // X-MSEdge-ClientID header
		"",                   // X-MSEdge-ClientIP header
		"",                   // X-Search-Location header
		"",                   // country code
		nil,                  // count
		"",                   // freshness
		"",                   // ID
		"",                   // video length
		"",                   // market
		nil,                  // offset
		"",                   // video pricing
		"",                   // video resolution
		"",                   // safe search
		"",                   // set lang
		nil,                  // text decorations
		"",                   // text format
	)

	return videos, err
}

//TrendingVideos returns the videos that are trending
func TrendingVideos(accountName string) (videosearch.TrendingVideos, error) {
	videoSearchClient := getVideoSearchClient(accountName)
	trendingVideos, err := videoSearchClient.Trending(context.Background(), "", "", "", "", "", "", "", "", "", "", nil, "")
	return trendingVideos, err
}
