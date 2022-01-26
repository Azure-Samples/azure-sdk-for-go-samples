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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datalake-store/armdatalakestore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	location          = "eastus2"
	resourceGroupName = "sample-resource-group"
	accountName       = "sample2datalake2account"
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

	account, err := createDataLakeStoreAccount(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("datalake store account:", *account.ID)

	account, err = getDataLakeStoreAccount(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get datalake store account:", *account.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createDataLakeStoreAccount(ctx context.Context, cred azcore.TokenCredential) (*armdatalakestore.Account, error) {
	accountsClient := armdatalakestore.NewAccountsClient(subscriptionID, cred, nil)
	pollerResp, err := accountsClient.BeginCreate(
		ctx,
		resourceGroupName,
		accountName,
		armdatalakestore.CreateDataLakeStoreAccountParameters{
			Location: to.StringPtr(location),
			Identity: &armdatalakestore.EncryptionIdentity{
				Type: to.StringPtr("SystemAssigned"),
			},
			Properties: &armdatalakestore.CreateDataLakeStoreAccountProperties{
				EncryptionConfig: &armdatalakestore.EncryptionConfig{
					Type: armdatalakestore.EncryptionConfigTypeServiceManaged.ToPtr(),
				},
				EncryptionState: armdatalakestore.EncryptionStateEnabled.ToPtr(),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, 30*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.Account, nil
}

func getDataLakeStoreAccount(ctx context.Context, cred azcore.TokenCredential) (*armdatalakestore.Account, error) {
	accountsClient := armdatalakestore.NewAccountsClient(subscriptionID, cred, nil)
	resp, err := accountsClient.Get(ctx, resourceGroupName, accountName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Account, nil
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
