// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cosmosdb

import (
	"context"
	"fmt"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
	"github.com/Azure/azure-sdk-for-go/services/cosmos-db/mgmt/2015-04-08/documentdb"
	"github.com/Azure/go-autorest/autorest/to"
)

func getDatabaseAccountClient() documentdb.DatabaseAccountsClient {
	dbAccountClient := documentdb.NewDatabaseAccountsClient(config.SubscriptionID())
	auth, _ := iam.GetResourceManagementAuthorizer()
	dbAccountClient.Authorizer = auth
	dbAccountClient.AddToUserAgent(config.UserAgent())
	return dbAccountClient
}

// CreateDatabaseAccount creates or updates an Azure Cosmos DB database account.
func CreateDatabaseAccount(ctx context.Context, accountName string) (dba documentdb.DatabaseAccount, err error) {
	dbAccountClient := getDatabaseAccountClient()
	future, err := dbAccountClient.CreateOrUpdate(
		ctx,
		config.GroupName(),
		accountName,
		documentdb.DatabaseAccountCreateUpdateParameters{
			Location: to.StringPtr(config.Location()),
			Kind:     documentdb.GlobalDocumentDB,
			DatabaseAccountCreateUpdateProperties: &documentdb.DatabaseAccountCreateUpdateProperties{
				DatabaseAccountOfferType: to.StringPtr("Standard"),
				Locations: &[]documentdb.Location{
					{
						FailoverPriority: to.Int32Ptr(0),
						LocationName:     to.StringPtr(config.Location()),
					},
				},
			},
		})
	if err != nil {
		return dba, fmt.Errorf("cannot create database account: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, dbAccountClient.Client)
	if err != nil {
		return dba, fmt.Errorf("cannot get the database account create or update future response: %v", err)
	}

	return future.Result(dbAccountClient)
}

// ListKeys gets the keys for a Azure Cosmos DB database account.
func ListKeys(ctx context.Context, accountName string) (documentdb.DatabaseAccountListKeysResult, error) {
	dbAccountClient := getDatabaseAccountClient()
	return dbAccountClient.ListKeys(ctx, config.GroupName(), accountName)
}
