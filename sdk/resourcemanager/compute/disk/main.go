// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"log"
	"os"
)

var (
	subscriptionID        string
	TenantID              string
	location              = "westus"
	resourceGroupName     = "sample-resource-group"
	diskName              = "sample-disk"
	vaultName             = "sample2vault"
	keyName               = "sample2key"
	diskEncryptionSetName = "sample-disk-encryption"
)

var (
	resourcesClientFactory *armresources.ClientFactory
	keyvaultClientFactory  *armkeyvault.ClientFactory
	computeClientFactory   *armcompute.ClientFactory
)

var (
	resourceGroupClient      *armresources.ResourceGroupsClient
	vaultsClient             *armkeyvault.VaultsClient
	keysClient               *armkeyvault.KeysClient
	disksClient              *armcompute.DisksClient
	diskEncryptionSetsClient *armcompute.DiskEncryptionSetsClient
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
	keysClient = keyvaultClientFactory.NewKeysClient()

	computeClientFactory, err = armcompute.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	disksClient = computeClientFactory.NewDisksClient()
	diskEncryptionSetsClient = computeClientFactory.NewDiskEncryptionSetsClient()

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

	key, err := createKey(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("key:", *key.ID)

	disk, err := createDisk(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("virtual disk:", *disk.ID)

	diskEncryptionSet, err := diskEncryptionSets(ctx, *vault.ID, *key.Properties.KeyURIWithVersion)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("disk encryption set:", *diskEncryptionSet.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err := cleanup(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createDisk(ctx context.Context) (*armcompute.Disk, error) {

	pollerResp, err := disksClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		diskName,
		armcompute.Disk{
			Location: to.Ptr(location),
			SKU: &armcompute.DiskSKU{
				Name: to.Ptr(armcompute.DiskStorageAccountTypesStandardLRS),
			},
			Properties: &armcompute.DiskProperties{
				CreationData: &armcompute.CreationData{
					CreateOption: to.Ptr(armcompute.DiskCreateOptionEmpty),
				},
				DiskSizeGB: to.Ptr[int32](64),
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

	return &resp.Disk, nil
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
						ObjectID: to.Ptr("00000000-0000-0000-0000-000000000000"),
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
func createKey(ctx context.Context) (*armkeyvault.Key, error) {

	secretResp, err := keysClient.CreateIfNotExist(
		ctx,
		resourceGroupName,
		vaultName,
		keyName,
		armkeyvault.KeyCreateParameters{
			Properties: &armkeyvault.KeyProperties{
				Attributes: &armkeyvault.KeyAttributes{
					Enabled: to.Ptr(true),
				},
				KeySize: to.Ptr[int32](2048),
				KeyOps: []*armkeyvault.JSONWebKeyOperation{
					to.Ptr(armkeyvault.JSONWebKeyOperationEncrypt),
					to.Ptr(armkeyvault.JSONWebKeyOperationDecrypt),
				},
				Kty: to.Ptr(armkeyvault.JSONWebKeyTypeRSA),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &secretResp.Key, nil
}

func diskEncryptionSets(ctx context.Context, vaultID, keyURL string) (*armcompute.DiskEncryptionSet, error) {

	pollerResp, err := diskEncryptionSetsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		diskEncryptionSetName,
		armcompute.DiskEncryptionSet{
			Location: to.Ptr(location),
			Identity: &armcompute.EncryptionSetIdentity{
				Type: to.Ptr(armcompute.DiskEncryptionSetIdentityTypeSystemAssigned),
			},
			Properties: &armcompute.EncryptionSetProperties{
				ActiveKey: &armcompute.KeyForDiskEncryptionSet{
					SourceVault: &armcompute.SourceVault{
						ID: to.Ptr(vaultID),
					},
					KeyURL: to.Ptr(keyURL),
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

	return &resp.DiskEncryptionSet, nil
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
