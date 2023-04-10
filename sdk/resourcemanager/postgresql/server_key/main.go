// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"log"
	"os"
)

var (
	subscriptionID    string
	TenantID          string
	ObjectID          string
	location          = "eastus"
	resourceGroupName = "sample-resource-group"
	serverName        = "sampleXserver"
	vaultName         = "sample2vault"
	keyName           = "sample2key"
	serverKeyName     = "sample-postgresql-key"
)

var (
	resourcesClientFactory  *armresources.ClientFactory
	keyvaultClientFactory   *armkeyvault.ClientFactory
	postgresqlClientFactory *armpostgresql.ClientFactory
)

var (
	resourceGroupClient *armresources.ResourceGroupsClient
	vaultsClient        *armkeyvault.VaultsClient
	keysClient          *armkeyvault.KeysClient
	serversClient       *armpostgresql.ServersClient
	serverKeysClient    *armpostgresql.ServerKeysClient
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

	postgresqlClientFactory, err = armpostgresql.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	serversClient = postgresqlClientFactory.NewServersClient()
	serverKeysClient = postgresqlClientFactory.NewServerKeysClient()

	resourceGroup, err := createResourceGroup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	server, err := createServer(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("postgresql server:", *server.ID)

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
	log.Println("postgresql server key:", *serverKey.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createServer(ctx context.Context) (*armpostgresql.Server, error) {

	pollerResp, err := serversClient.BeginCreate(
		ctx,
		resourceGroupName,
		serverName,
		armpostgresql.ServerForCreate{
			Location: to.Ptr(location),
			Properties: &armpostgresql.ServerPropertiesForDefaultCreate{
				CreateMode:                 to.Ptr(armpostgresql.CreateModeDefault),
				InfrastructureEncryption:   to.Ptr(armpostgresql.InfrastructureEncryptionDisabled),
				PublicNetworkAccess:        to.Ptr(armpostgresql.PublicNetworkAccessEnumEnabled),
				Version:                    to.Ptr(armpostgresql.ServerVersionEleven),
				AdministratorLogin:         to.Ptr("dummylogin"),
				AdministratorLoginPassword: to.Ptr("QWE123!@#"),
			},
			SKU: &armpostgresql.SKU{
				Name: to.Ptr("B_Gen5_1"),
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

func createServerKey(ctx context.Context, keyID string) (*armpostgresql.ServerKey, error) {

	pollerResp, err := serverKeysClient.BeginCreateOrUpdate(
		ctx,
		serverName,
		serverKeyName,
		resourceGroupName,
		armpostgresql.ServerKey{
			Properties: &armpostgresql.ServerKeyProperties{
				ServerKeyType: to.Ptr(armpostgresql.ServerKeyTypeAzureKeyVault),
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
