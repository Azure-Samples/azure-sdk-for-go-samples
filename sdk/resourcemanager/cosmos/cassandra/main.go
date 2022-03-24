package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cosmos/armcosmos"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
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
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createCassandraKeyspace(ctx context.Context, cred azcore.TokenCredential) (*armcosmos.CassandraKeyspaceGetResults, error) {
	cassandraResourcesClient := armcosmos.NewCassandraResourcesClient(subscriptionID, cred, nil)

	pollerResp, err := cassandraResourcesClient.BeginCreateUpdateCassandraKeyspace(
		ctx,
		resourceGroupName,
		accountName,
		keyspaceName,
		armcosmos.CassandraKeyspaceCreateUpdateParameters{
			Location: to.StringPtr(location),
			Properties: &armcosmos.CassandraKeyspaceCreateUpdateProperties{
				Resource: &armcosmos.CassandraKeyspaceResource{
					ID: to.StringPtr(keyspaceName),
				},
				Options: &armcosmos.CreateUpdateOptions{
					Throughput: to.Int32Ptr(2000),
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.CassandraKeyspaceGetResults, nil
}

func createCassandraTable(ctx context.Context, cred azcore.TokenCredential) (*armcosmos.CassandraTableGetResults, error) {
	cassandraResourcesClient := armcosmos.NewCassandraResourcesClient(subscriptionID, cred, nil)

	pollerResp, err := cassandraResourcesClient.BeginCreateUpdateCassandraTable(
		ctx,
		resourceGroupName,
		accountName,
		keyspaceName,
		tableName,
		armcosmos.CassandraTableCreateUpdateParameters{
			Location: to.StringPtr(location),
			Properties: &armcosmos.CassandraTableCreateUpdateProperties{
				Resource: &armcosmos.CassandraTableResource{
					ID:         to.StringPtr(tableName),
					DefaultTTL: to.Int32Ptr(100),
					Schema: &armcosmos.CassandraSchema{
						Columns: []*armcosmos.Column{
							{
								Name: to.StringPtr("columnA"),
								Type: to.StringPtr("Ascii"),
							},
						},
						PartitionKeys: []*armcosmos.CassandraPartitionKey{
							{
								Name: to.StringPtr("columnA"),
							},
						},
					},
				},
				Options: &armcosmos.CreateUpdateOptions{
					Throughput: to.Int32Ptr(2000),
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.CassandraTableGetResults, nil
}

func createDatabaseAccount(ctx context.Context, cred azcore.TokenCredential) (*armcosmos.DatabaseAccountGetResults, error) {
	databaseAccountsClient := armcosmos.NewDatabaseAccountsClient(subscriptionID, cred, nil)

	pollerResp, err := databaseAccountsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		accountName,
		armcosmos.DatabaseAccountCreateUpdateParameters{
			Location: to.StringPtr(location),
			Kind:     armcosmos.DatabaseAccountKindGlobalDocumentDB.ToPtr(),
			Properties: &armcosmos.DatabaseAccountCreateUpdateProperties{
				DatabaseAccountOfferType: to.StringPtr("Standard"),
				Locations: []*armcosmos.Location{
					{
						FailoverPriority: to.Int32Ptr(0),
						LocationName:     to.StringPtr(location),
					},
				},
				Capabilities: []*armcosmos.Capability{
					{
						Name: to.StringPtr("EnableCassandra"),
					},
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.DatabaseAccountGetResults, nil
}

func createResourceGroup(ctx context.Context, cred azcore.TokenCredential) (*armresources.ResourceGroup, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)

	resourceGroupResp, err := resourceGroupClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		armresources.ResourceGroup{
			Location: to.StringPtr(location),
		},
		nil)
	if err != nil {
		return nil, err
	}
	return &resourceGroupResp.ResourceGroup, nil
}

func cleanup(ctx context.Context, cred azcore.TokenCredential) (*http.Response, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)

	pollerResp, err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp.RawResponse, nil
}
