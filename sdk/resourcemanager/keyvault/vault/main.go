// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"log"
	"os"
)

var (
	subscriptionID    string
	TenantID          string
	ObjectID          string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	vaultName         = "sample2vaultalan"
)

var (
	resourcesClientFactory *armresources.ClientFactory
	keyvaultClientFactory  *armkeyvault.ClientFactory
)

var (
	resourceGroupClient *armresources.ResourceGroupsClient
	vaultsClient        *armkeyvault.VaultsClient
	managedHsmsClient   *armkeyvault.ManagedHsmsClient
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
	ctx := context.Background()

	resourcesClientFactory, err = armresources.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	resourceGroupClient = resourcesClientFactory.NewResourceGroupsClient()

	keyvaultClientFactory, err = armkeyvault.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	vaultsClient = keyvaultClientFactory.NewVaultsClient()
	managedHsmsClient = keyvaultClientFactory.NewManagedHsmsClient()

	resourceGroup, err := createResourceGroup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	vault, err := createVault(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("vault:", *vault.ID)

	vaultForDeployment, err := setVaultPermissionsForDeployment(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("vault for deployment:", *vaultForDeployment.ID)

	deletedVaults, err := deletedVaultList(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for i, v := range deletedVaults {
		log.Println("deleted vault:", i, *v.ID)
	}

	err = deleteVault(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("deleted vault.")

	err = purgeDeleted(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("purge deleted vault.")

	hsms, err := createManagedHsms(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("managed Hsms:", *hsms.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createVault(ctx context.Context) (*armkeyvault.Vault, error) {

	pollerResp, err := vaultsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		vaultName,
		armkeyvault.VaultCreateOrUpdateParameters{
			Location: to.Ptr(location),
			Properties: &armkeyvault.VaultProperties{
				SKU: &armkeyvault.SKU{
					Family: to.Ptr(armkeyvault.SKUFamilyA),
					Name:   to.Ptr(armkeyvault.SKUNameStandard),
				},
				TenantID: to.Ptr(TenantID),
				AccessPolicies: []*armkeyvault.AccessPolicyEntry{
					{
						TenantID: to.Ptr(TenantID),
						ObjectID: to.Ptr(ObjectID),
						Permissions: &armkeyvault.Permissions{
							Keys: []*armkeyvault.KeyPermissions{
								to.Ptr(armkeyvault.KeyPermissionsGet),
								to.Ptr(armkeyvault.KeyPermissionsList),
								to.Ptr(armkeyvault.KeyPermissionsCreate),
							},
							Secrets: []*armkeyvault.SecretPermissions{
								to.Ptr(armkeyvault.SecretPermissionsGet),
								to.Ptr(armkeyvault.SecretPermissionsList),
							},
							Certificates: []*armkeyvault.CertificatePermissions{
								to.Ptr(armkeyvault.CertificatePermissionsGet),
								to.Ptr(armkeyvault.CertificatePermissionsList),
								to.Ptr(armkeyvault.CertificatePermissionsCreate),
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

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Vault, nil
}

func setVaultPermissionsForDeployment(ctx context.Context) (*armkeyvault.Vault, error) {

	pollerResp, err := vaultsClient.BeginCreateOrUpdate(ctx, resourceGroupName, vaultName, armkeyvault.VaultCreateOrUpdateParameters{
		Location: to.Ptr(location),
		Properties: &armkeyvault.VaultProperties{
			SKU: &armkeyvault.SKU{
				Family: to.Ptr(armkeyvault.SKUFamilyA),
				Name:   to.Ptr(armkeyvault.SKUNameStandard),
			},
			TenantID:                     to.Ptr(TenantID),
			EnabledForDeployment:         to.Ptr(true),
			EnabledForTemplateDeployment: to.Ptr(true),
			AccessPolicies: []*armkeyvault.AccessPolicyEntry{
				{
					TenantID: to.Ptr(TenantID),
					ObjectID: to.Ptr("00000000-0000-0000-0000-000000000000"),
					Permissions: &armkeyvault.Permissions{
						Keys: []*armkeyvault.KeyPermissions{
							to.Ptr(armkeyvault.KeyPermissionsGet),
							to.Ptr(armkeyvault.KeyPermissionsList),
							to.Ptr(armkeyvault.KeyPermissionsCreate),
						},
						Secrets: []*armkeyvault.SecretPermissions{
							to.Ptr(armkeyvault.SecretPermissionsGet),
							to.Ptr(armkeyvault.SecretPermissionsGet),
							to.Ptr(armkeyvault.SecretPermissionsList),
						},
					},
				},
			},
		},
	}, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Vault, nil
}

func deletedVaultList(ctx context.Context) ([]*armkeyvault.DeletedVault, error) {

	deletedVaultResult := vaultsClient.NewListDeletedPager(nil)

	deleteVaults := make([]*armkeyvault.DeletedVault, 0)
	for deletedVaultResult.More() {
		resp, err := deletedVaultResult.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		deleteVaults = append(deleteVaults, resp.DeletedVaultListResult.Value...)
	}

	return deleteVaults, nil
}

func deleteVault(ctx context.Context) error {

	_, err := vaultsClient.Delete(ctx, resourceGroupName, vaultName, nil)
	if err != nil {
		return err
	}
	return nil
}

func purgeDeleted(ctx context.Context) error {

	pollerResp, err := vaultsClient.BeginPurgeDeleted(ctx, vaultName, location, nil)
	if err != nil {
		return err
	}

	_, err = pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func createManagedHsms(ctx context.Context) (*armkeyvault.ManagedHsm, error) {

	pollerResp, err := managedHsmsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		"sample-hsmsxx",
		armkeyvault.ManagedHsm{
			Location: to.Ptr(location),
			SKU: &armkeyvault.ManagedHsmSKU{
				Family: to.Ptr(armkeyvault.ManagedHsmSKUFamilyB),
				Name:   to.Ptr(armkeyvault.ManagedHsmSKUNameStandardB1),
			},
			Properties: &armkeyvault.ManagedHsmProperties{
				TenantID:   to.Ptr(TenantID),
				CreateMode: to.Ptr(armkeyvault.CreateModeDefault),
				InitialAdminObjectIDs: []*string{
					to.Ptr(ObjectID),
				},
			},
		},
		nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &resp.ManagedHsm, nil
}

func createResourceGroup(ctx context.Context) (*armresources.ResourceGroup, error) {

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

func cleanup(ctx context.Context) error {

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
