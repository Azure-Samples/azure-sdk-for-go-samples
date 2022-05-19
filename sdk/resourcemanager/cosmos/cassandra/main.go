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
	keyspaceName      = "sample-cosmos-keyspace"
	tableName         = "sample-cosmos-table"
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

	cassandraKeyspace, err := createCassandraKeyspace(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("cosmos cassandra keyspace:", *cassandraKeyspace.ID)

	cassandraTable, err := createCassandraTable(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("cosmos cassandra table:", *cassandraTable.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createCassandraKeyspace(ctx context.Context, cred azcore.TokenCredential) (*armcosmos.CassandraKeyspaceGetResults, error) {
	cassandraResourcesClient, err := armcosmos.NewCassandraResourcesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := cassandraResourcesClient.BeginCreateUpdateCassandraKeyspace(
		ctx,
		resourceGroupName,
		accountName,
		keyspaceName,
		armcosmos.CassandraKeyspaceCreateUpdateParameters{
			Location: to.Ptr(location),
			Properties: &armcosmos.CassandraKeyspaceCreateUpdateProperties{
				Resource: &armcosmos.CassandraKeyspaceResource{
					ID: to.Ptr(keyspaceName),
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
	return &resp.CassandraKeyspaceGetResults, nil
}

func createCassandraTable(ctx context.Context, cred azcore.TokenCredential) (*armcosmos.CassandraTableGetResults, error) {
	cassandraResourcesClient, err := armcosmos.NewCassandraResourcesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := cassandraResourcesClient.BeginCreateUpdateCassandraTable(
		ctx,
		resourceGroupName,
		accountName,
		keyspaceName,
		tableName,
		armcosmos.CassandraTableCreateUpdateParameters{
			Location: to.Ptr(location),
			Properties: &armcosmos.CassandraTableCreateUpdateProperties{
				Resource: &armcosmos.CassandraTableResource{
					ID:         to.Ptr(tableName),
					DefaultTTL: to.Ptr[int32](100),
					Schema: &armcosmos.CassandraSchema{
						Columns: []*armcosmos.Column{
							{
								Name: to.Ptr("columnA"),
								Type: to.Ptr("Ascii"),
							},
						},
						PartitionKeys: []*armcosmos.CassandraPartitionKey{
							{
								Name: to.Ptr("columnA"),
							},
						},
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
	return &resp.CassandraTableGetResults, nil
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
						LocationName:     to.Ptr(location),
					},
				},
				Capabilities: []*armcosmos.Capability{
					{
						Name: to.Ptr("EnableCassandra"),
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
