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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	TenantID          string
	ObjectID          string
	location          = "eastus"
	resourceGroupName = "sample-resource-group2"
	serverName        = "sampleXserver"
	vaultName         = "sample2vaultxx"
	keyName           = "sample2keyxx"
	serverKeyName     = "sample-postgresql-key"
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

	resourceGroup, err := createResourceGroup(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	server, err := createServer(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("postgresql server:", *server.ID)

	vault, err := createVault(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("vault:", *vault.ID)

	key, err := createKey(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("key:", *key.ID)

	serverKey, err := createServerKey(ctx, cred, *key.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("postgresql server key:", *serverKey.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createServer(ctx context.Context, cred azcore.TokenCredential) (*armpostgresql.Server, error) {
	serversClient := armpostgresql.NewServersClient(subscriptionID, cred, nil)

	pollerResp, err := serversClient.BeginCreate(
		ctx,
		resourceGroupName,
		serverName,
		armpostgresql.ServerForCreate{
			Location: to.StringPtr(location),
			Properties: &armpostgresql.ServerPropertiesForDefaultCreate{
				ServerPropertiesForCreate: armpostgresql.ServerPropertiesForCreate{
					CreateMode:               armpostgresql.CreateModeDefault.ToPtr(),
					InfrastructureEncryption: armpostgresql.InfrastructureEncryptionDisabled.ToPtr(),
					PublicNetworkAccess:      armpostgresql.PublicNetworkAccessEnumEnabled.ToPtr(),
					Version:                  armpostgresql.ServerVersionEleven.ToPtr(),
				},
				AdministratorLogin:         to.StringPtr("dummylogin"),
				AdministratorLoginPassword: to.StringPtr("QWE123!@#"),
			},
			SKU: &armpostgresql.SKU{
				Name: to.StringPtr("B_Gen5_1"),
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
	return &resp.Server, nil
}

func createVault(ctx context.Context, cred azcore.TokenCredential) (*armkeyvault.Vault, error) {
	vaultsClient := armkeyvault.NewVaultsClient(subscriptionID, cred, nil)

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
							Storage: []*armkeyvault.StoragePermissions{
								armkeyvault.StoragePermissionsGet.ToPtr(),
								armkeyvault.StoragePermissionsList.ToPtr(),
								armkeyvault.StoragePermissionsDelete.ToPtr(),
								armkeyvault.StoragePermissionsSet.ToPtr(),
							},
						},
					},
				},
				EnabledForDiskEncryption:  to.BoolPtr(true),
				EnableSoftDelete:          to.BoolPtr(true),
				SoftDeleteRetentionInDays: to.Int32Ptr(90),
				NetworkACLs: &armkeyvault.NetworkRuleSet{
					Bypass:              armkeyvault.NetworkRuleBypassOptionsAzureServices.ToPtr(),
					DefaultAction:       armkeyvault.NetworkRuleActionAllow.ToPtr(),
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

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.Vault, nil
}

func createKey(ctx context.Context, cred azcore.TokenCredential) (*armkeyvault.Key, error) {
	keysClient := armkeyvault.NewKeysClient(subscriptionID, cred, nil)

	secretResp, err := keysClient.CreateIfNotExist(
		ctx,
		resourceGroupName,
		vaultName,
		keyName,
		armkeyvault.KeyCreateParameters{
			Properties: &armkeyvault.KeyProperties{
				Attributes: &armkeyvault.KeyAttributes{
					Enabled: to.BoolPtr(true),
				},
				KeySize: to.Int32Ptr(2048),
				KeyOps: []*armkeyvault.JSONWebKeyOperation{
					armkeyvault.JSONWebKeyOperationEncrypt.ToPtr(),
					armkeyvault.JSONWebKeyOperationDecrypt.ToPtr(),
				},
				Kty: armkeyvault.JSONWebKeyTypeRSA.ToPtr(),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &secretResp.Key, nil
}

func createServerKey(ctx context.Context, cred azcore.TokenCredential, keyID string) (*armpostgresql.ServerKey, error) {
	serverKeysClient := armpostgresql.NewServerKeysClient(subscriptionID, cred, nil)

	pollerResp, err := serverKeysClient.BeginCreateOrUpdate(
		ctx,
		serverName,
		serverKeyName,
		resourceGroupName,
		armpostgresql.ServerKey{
			Properties: &armpostgresql.ServerKeyProperties{
				ServerKeyType: armpostgresql.ServerKeyTypeAzureKeyVault.ToPtr(),
				URI:           to.StringPtr(keyID),
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
	return &resp.ServerKey, nil
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
