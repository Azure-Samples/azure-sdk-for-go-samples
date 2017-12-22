package cognitiveservices

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v1.0/customsearch"
	"github.com/Azure/go-autorest/autorest"
)

func getCustomSearchClient(accountName string) customsearch.CustomInstanceClient {
	apiKey := getFirstKey(accountName)
	customSearchClient := customsearch.NewCustomInstanceClient()
	csAuthorizer := autorest.NewCognitiveServicesAuthorizer(apiKey)
	customSearchClient.Authorizer = csAuthorizer
	return customSearchClient
}

//CustomSearch returns answers based on a custom search instance
func CustomSearch(accountName string) (customsearch.WebWebAnswer, error) {

	customSearchClient := getCustomSearchClient(accountName)
	query := "Xbox"
	customConfig := int32(00000) // subsitute with custom config id configured at https://www.customsearch.ai

	searchResponse, err := customSearchClient.Search(context.Background(), "", query, "", "", "", "", "", &customConfig, "", nil, "", nil, "", "", nil, "")

	return *searchResponse.WebPages, err
}
