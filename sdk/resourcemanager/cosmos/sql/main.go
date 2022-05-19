// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cosmos/armcosmos"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"log"
	"os"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	accountName       = "sample-cosmos-account"
	databaseName      = "sample-sql-database"
	containerName     = "sample-sql-container"
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

	resourceGroup, err := createResourceGroup(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	databaseAccount, err := createDatabaseAccount(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("cosmos database account:", *databaseAccount.ID)

	sqlDatabase, err := createSqlDatabase(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("cosmos sql database:", *sqlDatabase.ID)

	sqlContainer, err := createSqlContainer(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("cosmos sql container:", *sqlContainer.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createDatabaseAccount(ctx context.Context, cred azcore.TokenCredential) (*armcosmos.DatabaseAccountGetResults, error) {
	databaseAccountsClient, err := armcosmos.NewDatabaseAccountsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

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
						IsZoneRedundant:  to.Ptr(false),
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

func createSqlDatabase(ctx context.Context, cred azcore.TokenCredential) (*armcosmos.SQLDatabaseGetResults, error) {
	sqlResourcesClient, err := armcosmos.NewSQLResourcesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := sqlResourcesClient.BeginCreateUpdateSQLDatabase(
		ctx,
		resourceGroupName,
		accountName,
		databaseName,
		armcosmos.SQLDatabaseCreateUpdateParameters{
			Location: to.Ptr(location),
			Properties: &armcosmos.SQLDatabaseCreateUpdateProperties{
				Resource: &armcosmos.SQLDatabaseResource{
					ID: to.Ptr(databaseName),
				},
				Options: &armcosmos.CreateUpdateOptions{
					Throughput: to.Ptr[int32](2000),
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
	return &resp.SQLDatabaseGetResults, nil
}

func createSqlContainer(ctx context.Context, cred azcore.TokenCredential) (*armcosmos.SQLContainerGetResults, error) {
	sqlResourcesClient, err := armcosmos.NewSQLResourcesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := sqlResourcesClient.BeginCreateUpdateSQLContainer(
		ctx,
		resourceGroupName,
		accountName,
		databaseName,
		containerName,
		armcosmos.SQLContainerCreateUpdateParameters{
			Location: to.Ptr(location),
			Properties: &armcosmos.SQLContainerCreateUpdateProperties{
				Resource: &armcosmos.SQLContainerResource{
					ID: to.Ptr(containerName),
					IndexingPolicy: &armcosmos.IndexingPolicy{
						IndexingMode: to.Ptr(armcosmos.IndexingModeConsistent),
						Automatic:    to.Ptr(true),
						IncludedPaths: []*armcosmos.IncludedPath{
							{
								Path: to.Ptr("/*"),
								Indexes: []*armcosmos.Indexes{
									{
										Kind:      to.Ptr(armcosmos.IndexKindRange),
										DataType:  to.Ptr(armcosmos.DataTypeString),
										Precision: to.Ptr[int32](-1),
									},
									{
										Kind:      to.Ptr(armcosmos.IndexKindRange),
										DataType:  to.Ptr(armcosmos.DataTypeNumber),
										Precision: to.Ptr[int32](-1),
									},
								},
							},
						},
						ExcludedPaths: []*armcosmos.ExcludedPath{},
					},
					PartitionKey: &armcosmos.ContainerPartitionKey{
						Paths: []*string{
							to.Ptr("/AccountNumber"),
						},
						Kind: to.Ptr(armcosmos.PartitionKindHash),
					},
					DefaultTTL: to.Ptr[int32](100),
					UniqueKeyPolicy: &armcosmos.UniqueKeyPolicy{
						UniqueKeys: []*armcosmos.UniqueKey{
							{
								Paths: []*string{
									to.Ptr("/testPath"),
								},
							},
						},
					},
					ConflictResolutionPolicy: &armcosmos.ConflictResolutionPolicy{
						Mode:                   to.Ptr(armcosmos.ConflictResolutionModeLastWriterWins),
						ConflictResolutionPath: to.Ptr("/path"),
					},
				},
				Options: &armcosmos.CreateUpdateOptions{
					Throughput: to.Ptr[int32](2000),
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
	return &resp.SQLContainerGetResults, nil
}

func createResourceGroup(ctx context.Context, cred azcore.TokenCredential) (*armresources.ResourceGroup, error) {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

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

func cleanup(ctx context.Context, cred azcore.TokenCredential) error {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

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
