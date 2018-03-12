// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cognitiveservices

import (
	"context"
	"log"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/mgmt/2017-04-18/cognitiveservices"
	"github.com/Azure/go-autorest/autorest"
)

func getCognitiveSevicesManagementClient() cognitiveservices.AccountsClient {
	token, _ := iam.GetResourceManagementToken(iam.AuthGrantType())
	accountClient := cognitiveservices.NewAccountsClient(internal.SubscriptionID())
	accountClient.Authorizer = autorest.NewBearerAuthorizer(token)
	accountClient.AddToUserAgent(internal.UserAgent())
	return accountClient
}

func getFirstKey(accountName string) string {
	managementClient := getCognitiveSevicesManagementClient()
	keys, err := managementClient.ListKeys(context.Background(), internal.ResourceGroupName(), accountName)
	if err != nil {
		log.Fatalf("failed to list keys: %v", err)
	}
	return *keys.Key1
}

//CreateCSAccount creates a Cognitive Services account of the specified type
func CreateCSAccount(accountName string, accountKind cognitiveservices.Kind) (*cognitiveservices.Account, error) {
	managementClient := getCognitiveSevicesManagementClient()
	location := "global"

	csAccount, err := managementClient.Create(
		context.Background(),
		internal.ResourceGroupName(),
		accountName,
		cognitiveservices.AccountCreateParameters{
			Kind: accountKind,
			Sku: &cognitiveservices.Sku{
				Name: "S1",
				Tier: cognitiveservices.Standard,
			},
			Location:   &location,
			Properties: &map[string]interface{}{},
		})
	if err != nil {
		return nil, err
	}

	// although service returns that the management plane is ready to use,
	// the dataplane needs more time to be ready
	time.Sleep(time.Second * 10)
	return &csAccount, nil
}
