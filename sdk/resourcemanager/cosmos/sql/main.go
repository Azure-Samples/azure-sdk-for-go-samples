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
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
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

func createSqlDatabase(ctx context.Context, cred azcore.TokenCredential) (*armcosmos.SQLDatabaseGetResults, error) {
	sqlResourcesClient := armcosmos.NewSQLResourcesClient(subscriptionID, cred, nil)

	pollerResp, err := sqlResourcesClient.BeginCreateUpdateSQLDatabase(
		ctx,
		resourceGroupName,
		accountName,
		databaseName,
		armcosmos.SQLDatabaseCreateUpdateParameters{
			Location: to.StringPtr(location),
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

func createSqlContainer(ctx context.Context, cred azcore.TokenCredential) (*armcosmos.SQLContainerGetResults, error) {
	sqlResourcesClient := armcosmos.NewSQLResourcesClient(subscriptionID, cred, nil)

	pollerResp, err := sqlResourcesClient.BeginCreateUpdateSQLContainer(
		ctx,
		resourceGroupName,
		accountName,
		databaseName,
		containerName,
		armcosmos.SQLContainerCreateUpdateParameters{
			Location: to.StringPtr(location),
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
