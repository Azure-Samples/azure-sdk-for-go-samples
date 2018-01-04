package cognitiveservices

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v1.0/imagesearch"
	"github.com/Azure/go-autorest/autorest"
)

func getImageSearchClient(accountName string) imagesearch.ImagesClient {
	apiKey := getFirstKey(accountName)
	imageSearchClient := imagesearch.NewImagesClient()
	csAuthorizer := autorest.NewCognitiveServicesAuthorizer(apiKey)
	imageSearchClient.Authorizer = csAuthorizer
	return imageSearchClient
}

//SearchImages returns a list of images
func SearchImages(accountName string) (imagesearch.Images, error) {
	imageSearchClient := getImageSearchClient(accountName)
	query := "canadian rockies"

	images, err := imageSearchClient.Search(context.Background(),
		"",
		query,
		"",
		"",
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
		"",
		"",
		"",
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		"",
		"",
		"",
		nil)

	return images, err
}
