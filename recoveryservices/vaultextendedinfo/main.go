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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservices"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	vaultName         = "sample-recoveryservice-vault"
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

	vault, err := createRecoveryServiceVault(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("recovery service vault:", *vault.ID)

	vaultExtendedInfo, err := createVaultExtendedInfo(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("recovery service vault extended info:", *vaultExtendedInfo.ID)

	vaultExtendedInfo, err = getVaultExtendedInfo(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get recovery service vault extended info:", *vaultExtendedInfo.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createRecoveryServiceVault(ctx context.Context, cred azcore.TokenCredential) (*armrecoveryservices.Vault, error) {
	vaultClient := armrecoveryservices.NewVaultsClient(subscriptionID, cred, nil)
	pollerResp, err := vaultClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		vaultName,
		armrecoveryservices.Vault{
			Location: to.StringPtr(location),
			SKU: &armrecoveryservices.SKU{
				Name: armrecoveryservices.SKUNameStandard.ToPtr(),
			},
			Properties: &armrecoveryservices.VaultProperties{},
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
	return &resp.Vault, err
}

func createVaultExtendedInfo(ctx context.Context, cred azcore.TokenCredential) (*armrecoveryservices.VaultExtendedInfoResource, error) {
	vaultExtendedInfoClient := armrecoveryservices.NewVaultExtendedInfoClient(subscriptionID, cred, nil)
	resp, err := vaultExtendedInfoClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		vaultName,
		armrecoveryservices.VaultExtendedInfoResource{
			Properties: &armrecoveryservices.VaultExtendedInfo{
				Algorithm: to.StringPtr("None"),
				//IntegrityKey: to.StringPtr(),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.VaultExtendedInfoResource, err
}

func getVaultExtendedInfo(ctx context.Context, cred azcore.TokenCredential) (*armrecoveryservices.VaultExtendedInfoResource, error) {
	vaultExtendedInfoClient := armrecoveryservices.NewVaultExtendedInfoClient(subscriptionID, cred, nil)
	resp, err := vaultExtendedInfoClient.Get(ctx, resourceGroupName, vaultName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.VaultExtendedInfoResource, err
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
