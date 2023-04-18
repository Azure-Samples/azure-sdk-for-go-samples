// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
	"log"
	"os"
)

var (
	subscriptionID    string
	TenantID          string
	ObjectID          string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	serverName        = "sample2server"
	vaultName         = "sample2vault223"
	keyName           = "sample2key223"
	serverKeyName     = "sample-server-key"
)

var (
	resourcesClientFactory *armresources.ClientFactory
	keyvaultClientFactory  *armkeyvault.ClientFactory
	sqlClientFactory       *armsql.ClientFactory
)

var (
	resourceGroupClient *armresources.ResourceGroupsClient
	vaultsClient        *armkeyvault.VaultsClient
	keysClient          *armkeyvault.KeysClient
	serversClient       *armsql.ServersClient
	serverKeysClient    *armsql.ServerKeysClient
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
	keysClient = keyvaultClientFactory.NewKeysClient()

	sqlClientFactory, err = armsql.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	serversClient = sqlClientFactory.NewServersClient()
	serverKeysClient = sqlClientFactory.NewServerKeysClient()

	resourceGroup, err := createResourceGroup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	server, err := createServer(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("server:", *server.ID)

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

	serverKey, err := createServerKey(ctx, *key.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("server key:", *serverKey.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createServer(ctx context.Context) (*armsql.Server, error) {

	pollerResp, err := serversClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		armsql.Server{
			Location: to.Ptr(location),
			Identity: &armsql.ResourceIdentity{
				Type: to.Ptr(armsql.IdentityTypeSystemAssigned),
			},
			Properties: &armsql.ServerProperties{
				AdministratorLogin:         to.Ptr("dummylogin"),
				AdministratorLoginPassword: to.Ptr("QWE123!@#"),
				Version:                    to.Ptr("12.0"),
				PublicNetworkAccess:        to.Ptr(armsql.ServerNetworkAccessFlagEnabled),
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
	return &resp.Server, nil
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
							Storage: []*armkeyvault.StoragePermissions{
								to.Ptr(armkeyvault.StoragePermissionsGet),
								to.Ptr(armkeyvault.StoragePermissionsList),
								to.Ptr(armkeyvault.StoragePermissionsDelete),
								to.Ptr(armkeyvault.StoragePermissionsSet),
							},
						},
					},
				},
				EnabledForDiskEncryption:  to.Ptr(true),
				EnableSoftDelete:          to.Ptr(true),
				SoftDeleteRetentionInDays: to.Ptr[int32](90),
				NetworkACLs: &armkeyvault.NetworkRuleSet{
					Bypass:              to.Ptr(armkeyvault.NetworkRuleBypassOptionsAzureServices),
					DefaultAction:       to.Ptr(armkeyvault.NetworkRuleActionAllow),
					IPRules:             []*armkeyvault.IPRule{},
					VirtualNetworkRules: []*armkeyvault.VirtualNetworkRule{},
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

func createServerKey(ctx context.Context, keyID string) (*armsql.ServerKey, error) {

	pollerResp, err := serverKeysClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		serverKeyName,
		armsql.ServerKey{
			Properties: &armsql.ServerKeyProperties{
				ServerKeyType: to.Ptr(armsql.ServerKeyTypeAzureKeyVault),
				URI:           to.Ptr(keyID),
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
	return &resp.ServerKey, nil
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
