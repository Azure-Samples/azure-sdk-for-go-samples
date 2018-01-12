// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cognitiveservices

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v1.0/spellcheck"
	"github.com/Azure/go-autorest/autorest"
)

func getSpellCheckClient(accountName string) spellcheck.BaseClient {
	apiKey := getFirstKey(accountName)
	spellCheckClient := spellcheck.New()
	csAuthorizer := autorest.NewCognitiveServicesAuthorizer(apiKey)
	spellCheckClient.Authorizer = csAuthorizer
	spellCheckClient.AddToUserAgent(helpers.UserAgent())
	return spellCheckClient
}

//SpellCheck spell checks the given input
func SpellCheck(accountName string) (spellcheck.SpellCheck, error) {
	spellCheckClient := getSpellCheckClient(accountName)
	input := "Bill Gatas"

	spellCheckResult, err := spellCheckClient.SpellCheckerMethod(
		context.Background(), // context
		"",                   // X-BingApis-SDK header
		input,                // text to check
		"",                   // Accept-Language header
		"",                   // Pragma header
		"",                   // User-Agent header
		"",                   // X-MSEdge-ClientID header
		"",                   // X-MSEdge-ClientIP header
		"",                   // X-Search-Location header
		"",                   // action type
		"",                   // app name
		"",                   // country code
		"",                   // client machine name
		"",                   // doc ID
		"",                   // market
		"",                   // session ID
		"",                   // set lang
		"proof",              // user ID
		"",                   // mode
		"",                   // pre context text
		"",                   // post context text
	)

	return spellCheckResult, err
}
