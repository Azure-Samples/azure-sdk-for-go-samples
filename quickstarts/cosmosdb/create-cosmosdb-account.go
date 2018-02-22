// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cosmosdb

import (
	"context"
	"fmt"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"

	"github.com/Azure/azure-sdk-for-go/services/cosmos-db/mgmt/2015-04-08/documentdb"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

func getAuthorizer() (*autorest.BearerAuthorizer, error) {
	token, err := iam.GetResourceManagementToken(iam.AuthGrantType())

	if err != nil {
		return nil, fmt.Errorf("Failure to get management token: %s", err.Error())
	}

	return autorest.NewBearerAuthorizer(token), nil
}

// CreateCosmosDbAccount creates a new CosmosDb account
func CreateCosmosDbAccount(context context.Context, cosmosDbAccountName string) (*documentdb.DatabaseAccount, error) {
	authorizer, authErr := getAuthorizer()

	if authErr != nil {
		return nil, fmt.Errorf("Failed to create Authorizer: %s", authErr.Error())
	}

	databaseAccountsClient := documentdb.NewDatabaseAccountsClient(helpers.SubscriptionID())
	databaseAccountsClient.Authorizer = authorizer

	dbAccountsClientFuture, createErr := databaseAccountsClient.CreateOrUpdate(
		context,
		helpers.ResourceGroupName(),
		cosmosDbAccountName,
		documentdb.DatabaseAccountCreateUpdateParameters{
			DatabaseAccountCreateUpdateProperties: &documentdb.DatabaseAccountCreateUpdateProperties{
				Locations: &[]documentdb.Location{
					documentdb.Location{LocationName: to.StringPtr(helpers.Location())},
				},
				DatabaseAccountOfferType: to.StringPtr("Standard"),
			},
			Name:     &cosmosDbAccountName,
			Location: to.StringPtr(helpers.Location()),
			Kind:     documentdb.GlobalDocumentDB,
			Tags:     *to.StringMapPtr(map[string]string{"sdk-sample": "golang"}),
		})

	if createErr != nil {
		return nil, fmt.Errorf("Error creating database account: %q", createErr.Error())
	}

	waitForCreateErr := dbAccountsClientFuture.WaitForCompletion(context, databaseAccountsClient.Client)

	if waitForCreateErr != nil {
		return nil, fmt.Errorf("Error while waiting for database to be created: %q", waitForCreateErr.Error())
	}

	databaseAccount, createResultErr := dbAccountsClientFuture.Result(databaseAccountsClient)

	if createResultErr != nil {
		return nil, fmt.Errorf("Error creating database account: %q", createResultErr.Error())
	}

	return &databaseAccount, nil
}
