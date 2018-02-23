// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cosmosdb

import (
	"context"
	"fmt"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"

	"github.com/Azure/azure-sdk-for-go/services/cosmos-db/mgmt/2015-04-08/documentdb"
)

// GetCosmosDbAccountKeys retrieves the access keys for the specified CosmosDB account
func GetCosmosDbAccountKeys(context context.Context, cosmosDbAccountName string) (*documentdb.DatabaseAccountListKeysResult, error) {
	authorizer, authErr := getAuthorizer()

	if authErr != nil {
		return nil, fmt.Errorf("Failed to create Authorizer: %s", authErr.Error())
	}

	databaseAccountsClient := documentdb.NewDatabaseAccountsClient(helpers.SubscriptionID())
	databaseAccountsClient.Authorizer = authorizer

	databaseAccountKeys, listKeysErr := databaseAccountsClient.ListKeys(
		context,
		helpers.ResourceGroupName(),
		cosmosDbAccountName)

	if listKeysErr != nil {
		return nil, fmt.Errorf("Error listing account keys: %q", listKeysErr.Error())
	}

	return &databaseAccountKeys, nil
}
