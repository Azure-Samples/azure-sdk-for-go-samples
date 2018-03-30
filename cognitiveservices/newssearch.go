// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cognitiveservices

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v1.0/newssearch"
	"github.com/Azure/go-autorest/autorest"
)

func getNewsSearchClient(accountName string) newssearch.NewsClient {
	apiKey := getFirstKey(accountName)
	newsSearchClient := newssearch.NewNewsClient()
	csAuthorizer := autorest.NewCognitiveServicesAuthorizer(apiKey)
	newsSearchClient.Authorizer = csAuthorizer
	newsSearchClient.AddToUserAgent(helpers.UserAgent())
	return newsSearchClient
}

//SearchNews returns a list of news
func SearchNews(accountName string) (newssearch.News, error) {
	newsSearchClient := getNewsSearchClient(accountName)
	query := "Quantum Computing"

	news, err := newsSearchClient.Search(
		context.Background(), // context
		query,                // query keyword
		"",                   // Accept-Language header
		"",                   // User-Agent header
		"",                   // X-MSEdge-ClientID header
		"",                   // X-MSEdge-ClientIP header
		"",                   // X-Search-Location header
		"",                   // country code
		nil,                  // count
		newssearch.Month,     // freshness
		"",                   // market
		nil,                  // offset
		nil,                  // original image
		newssearch.Strict,    // safe search
		"",                   // set lang
		"",                   // sort by
		nil,                  // text decorations
		newssearch.Raw,       // text format
	)

	return news, err
}
