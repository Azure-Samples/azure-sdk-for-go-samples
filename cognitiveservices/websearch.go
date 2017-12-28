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
		nil,
		"",
		"",
		nil,
		nil,
		nil,
		"",
		"",
		nil,
		"")

	return *searchResponse.WebPages, err
}
