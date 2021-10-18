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

	sqlDatabase, err := createSqlDatabase(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("cosmos sql database:", *sqlDatabase.ID)

	sqlContainer, err := createSqlContainer(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("cosmos sql container:", *sqlContainer.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
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
						IsZoneRedundant:  to.BoolPtr(false),
						LocationName:     to.StringPtr(location),
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

func createSqlDatabase(ctx context.Context, conn *arm.Connection) (*armcosmos.SQLDatabaseGetResults, error) {
	sqlResourcesClient := armcosmos.NewSQLResourcesClient(conn, subscriptionID)

	pollerResp, err := sqlResourcesClient.BeginCreateUpdateSQLDatabase(
		ctx,
		resourceGroupName,
		accountName,
		databaseName,
		armcosmos.SQLDatabaseCreateUpdateParameters{
			ARMResourceProperties: armcosmos.ARMResourceProperties{
				Location: to.StringPtr(location),
			},
			Properties: &armcosmos.SQLDatabaseCreateUpdateProperties{
				Resource: &armcosmos.SQLDatabaseResource{
					ID: to.StringPtr(databaseName),
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
	return &resp.SQLDatabaseGetResults, nil
}

func createSqlContainer(ctx context.Context, conn *arm.Connection) (*armcosmos.SQLContainerGetResults, error) {
	sqlResourcesClient := armcosmos.NewSQLResourcesClient(conn, subscriptionID)

	pollerResp, err := sqlResourcesClient.BeginCreateUpdateSQLContainer(
		ctx,
		resourceGroupName,
		accountName,
		databaseName,
		containerName,
		armcosmos.SQLContainerCreateUpdateParameters{
			ARMResourceProperties: armcosmos.ARMResourceProperties{
				Location: to.StringPtr(location),
			},
			Properties: &armcosmos.SQLContainerCreateUpdateProperties{
				Resource: &armcosmos.SQLContainerResource{
					ID: to.StringPtr(containerName),
					IndexingPolicy: &armcosmos.IndexingPolicy{
						IndexingMode: armcosmos.IndexingModeConsistent.ToPtr(),
						Automatic:    to.BoolPtr(true),
						IncludedPaths: []*armcosmos.IncludedPath{
							{
								Path: to.StringPtr("/*"),
								Indexes: []*armcosmos.Indexes{
									{
										Kind:      armcosmos.IndexKindRange.ToPtr(),
										DataType:  armcosmos.DataTypeString.ToPtr(),
										Precision: to.Int32Ptr(-1),
									},
									{
										Kind:      armcosmos.IndexKindRange.ToPtr(),
										DataType:  armcosmos.DataTypeNumber.ToPtr(),
										Precision: to.Int32Ptr(-1),
									},
								},
							},
						},
						ExcludedPaths: []*armcosmos.ExcludedPath{},
					},
					PartitionKey: &armcosmos.ContainerPartitionKey{
						Paths: []*string{
							to.StringPtr("/AccountNumber"),
						},
						Kind: armcosmos.PartitionKindHash.ToPtr(),
					},
					DefaultTTL: to.Int32Ptr(100),
					UniqueKeyPolicy: &armcosmos.UniqueKeyPolicy{
						UniqueKeys: []*armcosmos.UniqueKey{
							{
								Paths: []*string{
									to.StringPtr("/testPath"),
								},
							},
						},
					},
					ConflictResolutionPolicy: &armcosmos.ConflictResolutionPolicy{
						Mode:                   armcosmos.ConflictResolutionModeLastWriterWins.ToPtr(),
						ConflictResolutionPath: to.StringPtr("/path"),
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
	return &resp.SQLContainerGetResults, nil
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
