package cognitiveservices

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v1.0/customsearch"
	"github.com/Azure/go-autorest/autorest"
)

func getCustomSearchClient(accountName string) customsearch.CustomInstanceClient {
	apiKey := getFirstKey(accountName)
	customSearchClient := customsearch.NewCustomInstanceClient()
	csAuthorizer := autorest.NewCognitiveServicesAuthorizer(apiKey)
	customSearchClient.Authorizer = csAuthorizer
	customSearchClient.AddToUserAgent(helpers.UserAgent())
	return customSearchClient
}

//CustomSearch returns answers based on a custom search instance
func CustomSearch(accountName string) (customsearch.WebWebAnswer, error) {

	customSearchClient := getCustomSearchClient(accountName)
	query := "Xbox"
	customConfig := int32(00000) // subsitute with custom config id configured at https://www.customsearch.ai

	searchResponse, err := customSearchClient.Search(
		context.Background(), // context
		"",                   // X-BingApis-SDK header
		query,                // query keyword
		"",                   // Accept-Language header
		"",                   // User-Agent header
		"",                   // X-MSEdge-ClientID header
		"",                   // X-MSEdge-ClientIP header
		"",                   // X-Search-Location header
		&customConfig,        // custom config (see comment above)
		"",                   // country code
		nil,                  // count
		"",                   // market
		nil,                  // offset
		"",                   // safe search
		"",                   // set lang
		nil,                  // text decorations
		"",                   // text format
	)

	return *searchResponse.WebPages, err
}
