// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cognitiveservices

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v1.0/entitysearch"
	"github.com/Azure/go-autorest/autorest"
)

func getEntitySearchClient(accountName string) entitysearch.EntitiesClient {
	apiKey := getFirstKey(accountName)
	entitySearchClient := entitysearch.NewEntitiesClient()
	csAuthorizer := autorest.NewCognitiveServicesAuthorizer(apiKey)
	entitySearchClient.Authorizer = csAuthorizer
	entitySearchClient.AddToUserAgent(helpers.UserAgent())
	return entitySearchClient
}

//SearchEntities retunrs a list of entities
func SearchEntities(accountName string) (entitysearch.Entities, error) {
	entitySearchClient := getEntitySearchClient(accountName)
	query := "tom cruise"
	market := "en-us"
	searchResponse, err := entitySearchClient.Search(
		context.Background(), // context
		"",                   // X-BingApis-SDK header
		query,                // query keyword
		"",                   // Accept-Language header
		"",                   // pragma header
		"",                   // User-Agent header
		"",                   // X-MSEdge-ClientID header
		"",                   // X-MSEdge-ClientIP header
		"",                   // X-Search-Location header
		"",                   // country code
		market,               // market
		nil,                  // response filter
		nil,                  // response format
		"",                   // safe search
		"",                   // set lang
	)

	return *searchResponse.Entities, err
}
