package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
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

	conn := arm.NewDefaultConnection(cred, &arm.ConnectionOptions{
		Logging: policy.LogOptions{
			IncludeBody: true,
		},
	})
	ctx := context.Background()

	resourceGroup, err := createResourceGroup(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	databaseAccount, err := createDatabaseAccount(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("cosmos database account:", *databaseAccount.ID)

	cassandraKeyspace, err := createCassandraKeyspace(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("cosmos cassandra keyspace:", *cassandraKeyspace.ID)

	cassandraTable, err := createCassandraTable(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("cosmos cassandra table:", *cassandraTable.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createCassandraKeyspace(ctx context.Context, conn *arm.Connection) (*armcosmos.CassandraKeyspaceGetResults, error) {
	cassandraResourcesClient := armcosmos.NewCassandraResourcesClient(conn, subscriptionID)

	pollerResp, err := cassandraResourcesClient.BeginCreateUpdateCassandraKeyspace(
		ctx,
		resourceGroupName,
		accountName,
		keyspaceName,
		armcosmos.CassandraKeyspaceCreateUpdateParameters{
			ARMResourceProperties: armcosmos.ARMResourceProperties{
				Location: to.StringPtr(location),
			},
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

func createCassandraTable(ctx context.Context, conn *arm.Connection) (*armcosmos.CassandraTableGetResults, error) {
	cassandraResourcesClient := armcosmos.NewCassandraResourcesClient(conn, subscriptionID)

	pollerResp, err := cassandraResourcesClient.BeginCreateUpdateCassandraTable(
		ctx,
		resourceGroupName,
		accountName,
		keyspaceName,
		tableName,
		armcosmos.CassandraTableCreateUpdateParameters{
			ARMResourceProperties: armcosmos.ARMResourceProperties{
				Location: to.StringPtr(location),
			},
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

func createDatabaseAccount(ctx context.Context, conn *arm.Connection) (*armcosmos.DatabaseAccountGetResults, error) {
	databaseAccountsClient := armcosmos.NewDatabaseAccountsClient(conn, subscriptionID)

	pollerResp, err := databaseAccountsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		accountName,
		armcosmos.DatabaseAccountCreateUpdateParameters{
			ARMResourceProperties: armcosmos.ARMResourceProperties{
				Location: to.StringPtr(location),
			},
			Kind: armcosmos.DatabaseAccountKindGlobalDocumentDB.ToPtr(),
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

func createResourceGroup(ctx context.Context, conn *arm.Connection) (*armresources.ResourceGroup, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn, subscriptionID)

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

func cleanup(ctx context.Context, conn *arm.Connection) (*http.Response, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn, subscriptionID)

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
