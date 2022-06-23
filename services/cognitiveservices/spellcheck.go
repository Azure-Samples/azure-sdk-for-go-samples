// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package cognitiveservices

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/internal/config"
	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v1.0/spellcheck"
	"github.com/Azure/go-autorest/autorest"
)

func getSpellCheckClient(accountName string) spellcheck.BaseClient {
	apiKey := getFirstKey(accountName)
	spellCheckClient := spellcheck.New()
	csAuthorizer := autorest.NewCognitiveServicesAuthorizer(apiKey)
	spellCheckClient.Authorizer = csAuthorizer
	_ = spellCheckClient.AddToUserAgent(config.UserAgent())
	return spellCheckClient
}

//SpellCheck spell checks the given input
func SpellCheck(accountName string) (spellcheck.SpellCheck, error) {
	spellCheckClient := getSpellCheckClient(accountName)
	input := "Bill Gatas"

	spellCheckResult, err := spellCheckClient.SpellCheckerMethod(
		context.Background(),      // context
		input,                     // text to check
		"",                        // Accept-Language header
		"",                        // Pragma header
		"",                        // User-Agent header
		"",                        // X-MSEdge-ClientID header
		"",                        // X-MSEdge-ClientIP header
		"",                        // X-Search-Location header
		spellcheck.ActionType(""), // action type
		"",                        // app name
		"",                        // country code
		"",                        // client machine name
		"",                        // doc ID
		"",                        // market
		"",                        // session ID
		"",                        // set lang
		"proof",                   // user ID
		"",                        // mode
		"",                        // pre context text
		"",                        // post context text
	)

	return spellCheckResult, err
}
