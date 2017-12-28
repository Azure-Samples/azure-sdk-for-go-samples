package cognitiveservices

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v1.0/newssearch"
	"github.com/Azure/go-autorest/autorest"
)

func getNewsSearchClient(accountName string) newssearch.NewsClient {
	apiKey := getFirstKey(accountName)
	newsSearchClient := newssearch.NewNewsClient()
	csAuthorizer := autorest.NewCognitiveServicesAuthorizer(apiKey)
	newsSearchClient.Authorizer = csAuthorizer
	return newsSearchClient
}

//SearchNews returns a list of news
func SearchNews(accountName string) (newssearch.News, error) {
	newsSearchClient := getNewsSearchClient(accountName)
	query := "Quantum  Computing"

	news, err := newsSearchClient.Search(
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
		nil,
		nil,
		"",
		"",
		"",
		nil,
		"")

	return news, err
}
