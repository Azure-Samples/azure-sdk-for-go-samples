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
	accountName       = "sample-cosmos-mongodb"
	mongodbName       = "sample-mongodb"
	collectionName    = "sample-mongodb-collection"
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

	mongodbDatabase, err := createMongoDBDatabase(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("cosmos mongodb:", *mongodbDatabase.ID)

	getMongodb, err := getMongoDB(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get cosmos mongodb:", *getMongodb.ID)

	mongodbCollection, err := createMongoDBCollection(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("cosmos mongodb collection:", *mongodbCollection.ID)

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
			Kind:     armcosmos.DatabaseAccountKindMongoDB.ToPtr(),
			Properties: &armcosmos.DatabaseAccountCreateUpdateProperties{
				DatabaseAccountOfferType: to.StringPtr("Standard"),
				Locations: []*armcosmos.Location{
					{
						FailoverPriority: to.Int32Ptr(0),
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

func createMongoDBDatabase(ctx context.Context, cred azcore.TokenCredential) (*armcosmos.MongoDBDatabaseGetResults, error) {
	mongodbResourcesClient := armcosmos.NewMongoDBResourcesClient(subscriptionID, cred, nil)

	pollerResp, err := mongodbResourcesClient.BeginCreateUpdateMongoDBDatabase(
		ctx,
		resourceGroupName,
		accountName,
		mongodbName,
		armcosmos.MongoDBDatabaseCreateUpdateParameters{
			Location: to.StringPtr(location),
			Properties: &armcosmos.MongoDBDatabaseCreateUpdateProperties{
				Resource: &armcosmos.MongoDBDatabaseResource{
					ID: to.StringPtr(mongodbName),
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
	return &resp.MongoDBDatabaseGetResults, nil
}

func createMongoDBCollection(ctx context.Context, cred azcore.TokenCredential) (*armcosmos.MongoDBCollectionGetResults, error) {
	mongodbResourcesClient := armcosmos.NewMongoDBResourcesClient(subscriptionID, cred, nil)

	pollerResp, err := mongodbResourcesClient.BeginCreateUpdateMongoDBCollection(
		ctx,
		resourceGroupName,
		accountName,
		mongodbName,
		collectionName,
		armcosmos.MongoDBCollectionCreateUpdateParameters{
			Location: to.StringPtr(location),
			Properties: &armcosmos.MongoDBCollectionCreateUpdateProperties{
				Resource: &armcosmos.MongoDBCollectionResource{
					ID: to.StringPtr(collectionName),
					ShardKey: map[string]*string{
						"sample-shard-key": to.StringPtr("Hash"),
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
	return &resp.MongoDBCollectionGetResults, nil
}

func getMongoDB(ctx context.Context, cred azcore.TokenCredential) (*armcosmos.MongoDBDatabaseGetResults, error) {
	mongodbResourcesClient := armcosmos.NewMongoDBResourcesClient(subscriptionID, cred, nil)

	resp, err := mongodbResourcesClient.GetMongoDBDatabase(ctx, resourceGroupName, accountName, mongodbName, nil)
	if err != nil {
		return nil, err
	}

	return &resp.MongoDBDatabaseGetResults, nil
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
