package cognitiveservices

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v1.0/websearch"
	"github.com/Azure/go-autorest/autorest"
)

func getWebSearchClient(accountName string) websearch.WebClient {
	apiKey := getFirstKey(accountName)
	webSearchClient := websearch.NewWebClient()
	csAuthorizer := autorest.NewCognitiveServicesAuthorizer(apiKey)
	webSearchClient.Authorizer = csAuthorizer
	return webSearchClient
}

//SearchWeb returns a web answer contains a list of web pages
func SearchWeb(accountName string) (websearch.WebWebAnswer, error) {

	webSearchClient := getWebSearchClient(accountName)
	query := "tom cruise"
	searchResponse, err := webSearchClient.Search(
		context.Background(), // context
		"",                   // X-BingApis-SDK header
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
		"",                   // freshness
		"",                   // market
		nil,                  // offset
		nil,                  // promote
		nil,                  // response filter
		"",                   // safe search
		"",                   // set lang
		nil,                  // text decorations
		"",                   // text format
	)

	return *searchResponse.WebPages, err
}
