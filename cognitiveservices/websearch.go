// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cognitiveservices

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v1.0/websearch"
	"github.com/Azure/go-autorest/autorest"
)

func getWebSearchClient(accountName string) websearch.WebClient {
	apiKey := getFirstKey(accountName)
	webSearchClient := websearch.NewWebClient()
	csAuthorizer := autorest.NewCognitiveServicesAuthorizer(apiKey)
	webSearchClient.Authorizer = csAuthorizer
	webSearchClient.AddToUserAgent(config.UserAgent())
	return webSearchClient
}

//SearchWeb returns a web answer contains a list of web pages
func SearchWeb(accountName string) (*websearch.WebWebAnswer, error) {
	webSearchClient := getWebSearchClient(accountName)
	query := "tom cruise"
	searchResponse, err := webSearchClient.Search(
		context.Background(), // context
		query,                // query keyword
		"",                   // Accept-Language header
		"",                   // Pragma header
		"",                   // User-Agent header
		"",                   // X-MSEdge-ClientID header
		"",                   // X-MSEdge-ClientIP header
		"",                   // X-Search-Location header
		nil,                  // answer count
		"",                   // country code
		nil,                  // count
		websearch.Week,       // freshness
		"",                   // market
		nil,                  // offset
		[]websearch.AnswerType{}, // promote
		[]websearch.AnswerType{}, // response filter
		websearch.Strict,         // safe search
		"",                       // set lang
		nil,                      // text decorations
		websearch.Raw,            // text format
	)
	if err != nil {
		return nil, err
	}

	return searchResponse.WebPages, nil
}
