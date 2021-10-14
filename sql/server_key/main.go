package main

import (
	"context"
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
	"github.com/Azure/azure-sdk-for-go/sdk/sql/armsql"
)

var (
	subscriptionID    string
	TenantID          string
	ObjectID          string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	serverName        = "sample2server"
	vaultName         = "sample2vault"
	keyName           = "sample2key"
	serverKeyName     = "sample-server-key"
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

	server, err := createServer(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("server:", *server.ID)

	vault, err := createVault(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("vault:", *vault.ID)

	key, err := createKey(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("key:", *key.ID)

	serverKey, err := createServerKey(ctx, conn, *key.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("server key:", *serverKey.ID)

	//keepResource := os.Getenv("KEEP_RESOURCE")
	//if len(keepResource) == 0 {
	//	_, err := cleanup(ctx, conn)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	log.Println("cleaned up successfully.")
	//}
}

func createServer(ctx context.Context, conn *arm.Connection) (*armsql.Server, error) {
	serversClient := armsql.NewServersClient(conn, subscriptionID)

	pollerResp, err := serversClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		armsql.Server{
			TrackedResource: armsql.TrackedResource{
				Location: to.StringPtr(location),
			},
			Identity: &armsql.ResourceIdentity{
				Type: armsql.IdentityTypeSystemAssigned.ToPtr(),
			},
			Properties: &armsql.ServerProperties{
				AdministratorLogin:         to.StringPtr("dummylogin"),
				AdministratorLoginPassword: to.StringPtr("QWE123!@#"),
				Version:                    to.StringPtr("12.0"),
				PublicNetworkAccess:        armsql.ServerNetworkAccessFlagEnabled.ToPtr(),
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

func createKey(ctx context.Context, conn *arm.Connection) (*armkeyvault.Key, error) {
	keysClient := armkeyvault.NewKeysClient(conn, subscriptionID)

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

func createServerKey(ctx context.Context, conn *arm.Connection, keyID string) (*armsql.ServerKey, error) {
	serverKeysClient := armsql.NewServerKeysClient(conn, subscriptionID)

	pollerResp, err := serverKeysClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		serverKeyName,
		armsql.ServerKey{
			Properties: &armsql.ServerKeyProperties{
				ServerKeyType: armsql.ServerKeyTypeAzureKeyVault.ToPtr(),
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
