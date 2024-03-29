// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cosmos/armcosmos/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"log"
	"os"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	accountName       = "sample-cosmos-account"
)

var (
	resourcesClientFactory *armresources.ClientFactory
	cosmosClientFactory    *armcosmos.ClientFactory
)

var (
	resourceGroupClient    *armresources.ResourceGroupsClient
	databaseAccountsClient *armcosmos.DatabaseAccountsClient
)

func main() {
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set.")
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	resourcesClientFactory, err = armresources.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	resourceGroupClient = resourcesClientFactory.NewResourceGroupsClient()

	cosmosClientFactory, err = armcosmos.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	databaseAccountsClient = cosmosClientFactory.NewDatabaseAccountsClient()

	resourceGroup, err := createResourceGroup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	databaseAccount, err := createDatabaseAccount(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("cosmos database account:", *databaseAccount.ID)

	listKeysResp, err := listKeysDatabaseAccount(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(":", *listKeysResp.PrimaryReadonlyMasterKey)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createDatabaseAccount(ctx context.Context) (*armcosmos.DatabaseAccountGetResults, error) {

	pollerResp, err := databaseAccountsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		accountName,
		armcosmos.DatabaseAccountCreateUpdateParameters{
			Location: to.Ptr(location),
			Kind:     to.Ptr(armcosmos.DatabaseAccountKindGlobalDocumentDB),
			Properties: &armcosmos.DatabaseAccountCreateUpdateProperties{
				DatabaseAccountOfferType: to.Ptr("Standard"),
				Locations: []*armcosmos.Location{
					{
						FailoverPriority: to.Ptr[int32](0),
						LocationName:     to.Ptr(location),
					},
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.DatabaseAccountGetResults, nil
}

func listKeysDatabaseAccount(ctx context.Context) (*armcosmos.DatabaseAccountListKeysResult, error) {

	databaseAccountsListKeysResp, err := databaseAccountsClient.ListKeys(ctx, resourceGroupName, accountName, nil)
	if err != nil {
		return nil, err
	}

	return &databaseAccountsListKeysResp.DatabaseAccountListKeysResult, nil
}

func createResourceGroup(ctx context.Context) (*armresources.ResourceGroup, error) {

	resourceGroupResp, err := resourceGroupClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		armresources.ResourceGroup{
			Location: to.Ptr(location),
		},
		nil)
	if err != nil {
		return nil, err
	}
	return &resourceGroupResp.ResourceGroup, nil
}

func cleanup(ctx context.Context) error {

	pollerResp, err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return err
	}

	_, err = pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}
	return nil
}
