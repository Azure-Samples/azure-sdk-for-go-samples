package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/armcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/armkeyvault"
	"github.com/Azure/azure-sdk-for-go/sdk/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/to"
)

var (
	subscriptionID    string
	TenantID          string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	vaultName         = "sample2vault2"
	secretName        = "sample2secret"
)

func main() {
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set.")
	}

	TenantID = os.Getenv("AZURE_TENANT_ID")
	if len(TenantID) == 0 {
		log.Fatal("AZURE_TENANT_ID is not set.")
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}

	conn := armcore.NewDefaultConnection(cred, &armcore.ConnectionOptions{
		Logging: azcore.LogOptions{
			IncludeBody: true,
		},
	})
	ctx := context.Background()

	resourceGroup, err := createResourceGroup(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	vault, err := createVault(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("vault:", *vault.ID)

	secret, err := createSecret(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("secret:", *secret.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createVault(ctx context.Context, conn *armcore.Connection) (*armkeyvault.Vault, error) {
	vaultsClient := armkeyvault.NewVaultsClient(conn, subscriptionID)

	pollerResp, err := vaultsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		vaultName,
		armkeyvault.VaultCreateOrUpdateParameters{
			Location: to.StringPtr(location),
			Properties: &armkeyvault.VaultProperties{
				SKU: &armkeyvault.SKU{
					Family: armkeyvault.SKUFamilyA.ToPtr(),
					Name:   armkeyvault.SKUNameStandard.ToPtr(),
				},
				TenantID:       to.StringPtr(TenantID),
				AccessPolicies: []*armkeyvault.AccessPolicyEntry{},
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
	return resp.Vault, nil
}

func createSecret(ctx context.Context, conn *armcore.Connection) (*armkeyvault.Secret, error) {
	secretsClient := armkeyvault.NewSecretsClient(conn, subscriptionID)

	secretResp, err := secretsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		vaultName,
		secretName,
		armkeyvault.SecretCreateOrUpdateParameters{
			Properties: &armkeyvault.SecretProperties{
				Attributes: &armkeyvault.SecretAttributes{
					Attributes: armkeyvault.Attributes{
						Enabled: to.BoolPtr(true),
					},
				},
				Value: to.StringPtr("sample-secret-value"),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return secretResp.Secret, nil
}

func createResourceGroup(ctx context.Context, conn *armcore.Connection) (*armresources.ResourceGroup, error) {
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
	return resourceGroupResp.ResourceGroup, nil
}

func cleanup(ctx context.Context, conn *armcore.Connection) (*http.Response, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn, subscriptionID)

	pollerResp, err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
