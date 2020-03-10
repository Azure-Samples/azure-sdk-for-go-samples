// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cognitiveservices

import (
	"context"
	"log"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/mgmt/2017-04-18/cognitiveservices"
	"github.com/Azure/go-autorest/autorest/to"
)

func getCognitiveSevicesManagementClient() cognitiveservices.AccountsClient {
	accountClient := cognitiveservices.NewAccountsClient(config.SubscriptionID())
	auth, _ := iam.GetResourceManagementAuthorizer()
	accountClient.Authorizer = auth
	accountClient.AddToUserAgent(config.UserAgent())
	return accountClient
}

func getFirstKey(accountName string) string {
	managementClient := getCognitiveSevicesManagementClient()
	keys, err := managementClient.ListKeys(context.Background(), config.GroupName(), accountName)
	if err != nil {
		log.Fatalf("failed to list keys: %v", err)
	}
	return *keys.Key1
}

//CreateCSAccount creates a Cognitive Services account of the specified type
func CreateCSAccount(accountName string, accountKind string) (*cognitiveservices.Account, error) {
	managementClient := getCognitiveSevicesManagementClient()
	location := "global"

	csAccount, err := managementClient.Create(
		context.Background(),
		config.GroupName(),
		accountName,
		cognitiveservices.Account{
			Kind: &accountKind,
			Sku: &cognitiveservices.Sku{
				Name: to.StringPtr("S1"),
				Tier: cognitiveservices.Standard,
			},
			Location:   &location,
			Properties: nil,
		})
	if err != nil {
		return nil, err
	}

	// although service returns that the management plane is ready to use,
	// the dataplane needs more time to be ready
	time.Sleep(time.Second * 10)
	return &csAccount, nil
}
