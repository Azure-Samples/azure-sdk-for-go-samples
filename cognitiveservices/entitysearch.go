package cognitiveservices

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v1.0/entitysearch"
	"github.com/Azure/go-autorest/autorest"
)

func getEntitySearchClient(accountName string) entitysearch.EntitiesClient {
	apiKey := getFirstKey(accountName)
	entitySearchClient := entitysearch.NewEntitiesClient()
	csAuthorizer := autorest.NewCognitiveServicesAuthorizer(apiKey)
	entitySearchClient.Authorizer = csAuthorizer
	return entitySearchClient
}

//SearchEntities retunrs a list of entities
func SearchEntities(accountName string) (entitysearch.Entities, error) {
	entitySearchClient := getEntitySearchClient(accountName)
	query := "tom cruise"
	market := "en-us"
	searchResponse, err := entitySearchClient.Search(
		context.Background(),
		"",
		query,
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		market,
		nil,
		nil,
		"",
		"")

	return *searchResponse.Entities, err
}
