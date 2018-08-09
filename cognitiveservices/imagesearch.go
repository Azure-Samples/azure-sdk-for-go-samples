// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cognitiveservices

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v1.0/imagesearch"
	"github.com/Azure/go-autorest/autorest"
)

func getImageSearchClient(accountName string) imagesearch.ImagesClient {
	apiKey := getFirstKey(accountName)
	imageSearchClient := imagesearch.NewImagesClient()
	csAuthorizer := autorest.NewCognitiveServicesAuthorizer(apiKey)
	imageSearchClient.Authorizer = csAuthorizer
	imageSearchClient.AddToUserAgent(config.UserAgent())
	return imageSearchClient
}

//SearchImages returns a list of images
func SearchImages(accountName string) (imagesearch.Images, error) {
	imageSearchClient := getImageSearchClient(accountName)
	query := "canadian rockies"

	images, err := imageSearchClient.Search(
		context.Background(),  // context
		query,                 // query keyword
		"",                    // Accept-Language header
		"",                    // User-Agent header
		"",                    // X-MSEdge-ClientID header
		"",                    // X-MSEdge-ClientIP header
		"",                    // X-Search-Location header
		imagesearch.Square,    // image aspect
		imagesearch.ColorOnly, // image color
		"",                // country code
		nil,               // count
		imagesearch.Month, // freshness
		nil,               // height
		"",                // ID
		imagesearch.ImageContent(""), // image content
		imagesearch.Photo,            // image type
		imagesearch.ImageLicenseAll,  // image license
		"",                       // market
		nil,                      // max file size
		nil,                      // max height
		nil,                      // max width
		nil,                      // min file size
		nil,                      // min height
		nil,                      // min width
		nil,                      // offset
		imagesearch.Strict,       // safe search
		imagesearch.ImageSizeAll, // image size
		"",  // set lang
		nil, // width
	)

	return images, err
}
