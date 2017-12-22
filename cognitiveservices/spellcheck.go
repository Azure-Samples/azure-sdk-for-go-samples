package cognitiveservices

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v1.0/spellcheck"
	"github.com/Azure/go-autorest/autorest"
)

func getSpellCheckClient(accountName string) spellcheck.BaseClient {
	apiKey := getFirstKey(accountName)
	spellCheckClient := spellcheck.New()
	csAuthorizer := autorest.NewCognitiveServicesAuthorizer(apiKey)
	spellCheckClient.Authorizer = csAuthorizer
	return spellCheckClient
}

//SpellCheck spell checks the given input
func SpellCheck(accountName string) (spellcheck.SpellCheck, error) {
	spellCheckClient := getSpellCheckClient(accountName)
	input := "Bill Gatas"

	spellCheckResult, err := spellCheckClient.SpellCheckerMethod(context.Background(), "", input, "", "", "", "", "", "", "", "", "", "", "", "", "", "", "proof", "", "", "")

	return spellCheckResult, err
}
