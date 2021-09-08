package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/armkeyvault"
	"github.com/Azure/azure-sdk-for-go/sdk/resources/armresources"
)

var (
	subscriptionID    string
	TenantID          string
	ObjectID          string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	vaultName         = "sample2vault"
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

	ObjectID = os.Getenv("AZURE_OBJECT_ID")
	if len(ObjectID) == 0 {
		log.Fatal("AZURE_OBJECT_ID is not set.")
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

	vault, err := createVault(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("vault:", *vault.ID)

	vaultForDeployment, err := setVaultPermissionsForDeployment(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("vault for deployment:", *vaultForDeployment.ID)

	deletedVaults := deletedVaultList(ctx, conn)
	for i, v := range deletedVaults {
		log.Println("deleted vault:", i, *v.ID)
	}

	resp, err := deleteVault(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("deleted vault.", resp)

	resp, err = purgeDeleted(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("purge deleted vault.", resp)

	hsms, err := createManagedHsms(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("managed Hsms:", *hsms.ID)
	data, _ := json.Marshal(hsms)
	log.Println(string(data))

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createVault(ctx context.Context, conn *arm.Connection) (*armkeyvault.Vault, error) {
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
				TenantID: to.StringPtr(TenantID),
				AccessPolicies: []*armkeyvault.AccessPolicyEntry{
					{
						TenantID: to.StringPtr(TenantID),
						ObjectID: to.StringPtr(ObjectID),
						Permissions: &armkeyvault.Permissions{
							Keys: []*armkeyvault.KeyPermissions{
								armkeyvault.KeyPermissionsGet.ToPtr(),
								armkeyvault.KeyPermissionsList.ToPtr(),
								armkeyvault.KeyPermissionsCreate.ToPtr(),
							},
							Secrets: []*armkeyvault.SecretPermissions{
								armkeyvault.SecretPermissionsGet.ToPtr(),
								armkeyvault.SecretPermissionsList.ToPtr(),
							},
							Certificates: []*armkeyvault.CertificatePermissions{
								armkeyvault.CertificatePermissionsGet.ToPtr(),
								armkeyvault.CertificatePermissionsList.ToPtr(),
								armkeyvault.CertificatePermissionsCreate.ToPtr(),
							},
						},
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
	return &resp.Vault, nil
}

func setVaultPermissionsForDeployment(ctx context.Context, conn *arm.Connection) (*armkeyvault.Vault, error) {
	vaultsClient := armkeyvault.NewVaultsClient(conn, subscriptionID)

	pollerResp, err := vaultsClient.BeginCreateOrUpdate(ctx, resourceGroupName, vaultName, armkeyvault.VaultCreateOrUpdateParameters{
		Location: to.StringPtr(location),
		Properties: &armkeyvault.VaultProperties{
			SKU: &armkeyvault.SKU{
				Family: armkeyvault.SKUFamilyA.ToPtr(),
				Name:   armkeyvault.SKUNameStandard.ToPtr(),
			},
			TenantID:                     to.StringPtr(TenantID),
			EnabledForDeployment:         to.BoolPtr(true),
			EnabledForTemplateDeployment: to.BoolPtr(true),
			AccessPolicies: []*armkeyvault.AccessPolicyEntry{
				{
					TenantID: to.StringPtr(TenantID),
					ObjectID: to.StringPtr("00000000-0000-0000-0000-000000000000"),
					Permissions: &armkeyvault.Permissions{
						Keys: []*armkeyvault.KeyPermissions{
							armkeyvault.KeyPermissionsGet.ToPtr(),
							armkeyvault.KeyPermissionsList.ToPtr(),
							armkeyvault.KeyPermissionsCreate.ToPtr(),
						},
						Secrets: []*armkeyvault.SecretPermissions{
							armkeyvault.SecretPermissionsGet.ToPtr(),
							armkeyvault.SecretPermissionsGet.ToPtr(),
							armkeyvault.SecretPermissionsList.ToPtr(),
						},
					},
				},
			},
		},
	}, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.Vault, nil
}

func deletedVaultList(ctx context.Context, conn *arm.Connection) []*armkeyvault.DeletedVault {
	vaultsClient := armkeyvault.NewVaultsClient(conn, subscriptionID)

	deletedVaultResult := vaultsClient.ListDeleted(nil)

	deleteVaults := make([]*armkeyvault.DeletedVault, 0)
	for deletedVaultResult.NextPage(ctx) {
		resp := deletedVaultResult.PageResponse()
		deleteVaults = append(deleteVaults, resp.DeletedVaultListResult.Value...)
	}

	return deleteVaults
}

func deleteVault(ctx context.Context, conn *arm.Connection) (*http.Response, error) {
	vaultsClient := armkeyvault.NewVaultsClient(conn, subscriptionID)

	resp, err := vaultsClient.Delete(ctx, resourceGroupName, vaultName, nil)
	if err != nil {
		return nil, err
	}
	return resp.RawResponse, nil
}

func purgeDeleted(ctx context.Context, conn *arm.Connection) (*http.Response, error) {
	vaultsClient := armkeyvault.NewVaultsClient(conn, subscriptionID)

	pollerResp, err := vaultsClient.BeginPurgeDeleted(ctx, vaultName, location, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}

	return resp.RawResponse, nil
}

func createManagedHsms(ctx context.Context, conn *arm.Connection) (*armkeyvault.ManagedHsm, error) {
	client := armkeyvault.NewManagedHsmsClient(conn, subscriptionID)

	pollerResp, err := client.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		"sample-hsms",
		armkeyvault.ManagedHsm{
			ManagedHsmResource: armkeyvault.ManagedHsmResource{
				Location: to.StringPtr(location),
				SKU: &armkeyvault.ManagedHsmSKU{
					Family: armkeyvault.ManagedHsmSKUFamilyB.ToPtr(),
					Name:   armkeyvault.ManagedHsmSKUNameStandardB1.ToPtr(),
				},
			},
			Properties: &armkeyvault.ManagedHsmProperties{
				TenantID:   to.StringPtr(TenantID),
				CreateMode: armkeyvault.CreateModeDefault.ToPtr(),
				InitialAdminObjectIDs: []*string{
					to.StringPtr(ObjectID),
				},
			},
		},
		nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}

	return &resp.ManagedHsm, nil
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
