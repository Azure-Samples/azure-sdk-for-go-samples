package cognitiveservices

import (
	"context"
	"log"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/mgmt/2017-04-18/cognitiveservices"
	"github.com/Azure/go-autorest/autorest"
)

func getCognitiveSevicesManagementClient() cognitiveservices.AccountsClient {
	token, _ := iam.GetResourceManagementToken(iam.OAuthGrantTypeDeviceFlow)
	accountClient := cognitiveservices.NewAccountsClient(helpers.SubscriptionID())
	accountClient.Authorizer = autorest.NewBearerAuthorizer(token)

	return accountClient
}

func getFirstKey(accountName string) string {
	managementClient := getCognitiveSevicesManagementClient()
	keys, err := managementClient.ListKeys(context.Background(), helpers.ResourceGroupName(), accountName)
	if err != nil {
		log.Fatalf("failed to list keys: %v", err)
	}
	return *keys.Key1
}

//CreateCSAccount creates a Cognitive Services account of the specified type
func CreateCSAccount(accountName string, accountKind cognitiveservices.Kind) (*cognitiveservices.Account, error) {
	managementClient := getCognitiveSevicesManagementClient()
	location := "global"
	props := map[string]interface{}{}
	params := cognitiveservices.AccountCreateParameters{
		Kind: accountKind,
		Sku: &cognitiveservices.Sku{
			Name: "S1",
			Tier: cognitiveservices.Standard,
		},
		Location:   &location,
		Properties: &props,
	}

	csAccount, err := managementClient.Create(context.Background(), helpers.ResourceGroupName(), accountName, params)
	if err != nil {
		return nil, err
	}

	// Need to wait because although service returns that the account is ready, using the dataplane immediatley will fail
	time.Sleep(time.Second * 15)
	return &csAccount, nil
}
