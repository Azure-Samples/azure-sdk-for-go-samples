package keyvault

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/services/keyvault/mgmt/2016-10-01/keyvault"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	uuid "github.com/satori/go.uuid"
)

func getVaultsClient() keyvault.VaultsClient {
	token, _ := iam.GetResourceManagementToken(iam.OAuthGrantTypeServicePrincipal)
	vaultsClient := keyvault.NewVaultsClient(helpers.SubscriptionID())
	vaultsClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return vaultsClient
}

// CreateVault creates a new vault
func CreateVault(ctx context.Context, vaultName string) (keyvault.Vault, error) {
	vaultsClient := getVaultsClient()
	tenantID, err := uuid.FromString(iam.TenantID())
	if err != nil {
		return keyvault.Vault{}, err
	}

	return vaultsClient.CreateOrUpdate(
		ctx,
		helpers.ResourceGroupName(),
		vaultName,
		keyvault.VaultCreateOrUpdateParameters{
			Location: to.StringPtr(helpers.Location()),
			Properties: &keyvault.VaultProperties{
				TenantID: &tenantID,
				Sku: &keyvault.Sku{
					Family: to.StringPtr("A"),
					Name:   keyvault.Standard,
				},
				AccessPolicies: &[]keyvault.AccessPolicyEntry{},
			},
		},
	)
}

// GetVault returns an existing vault
func GetVault(ctx context.Context, vaultName string) (keyvault.Vault, error) {
	vaultsClient := getVaultsClient()
	return vaultsClient.Get(ctx, helpers.ResourceGroupName(), vaultName)
}

// SetVaultPermissions adds an access policy permitting this app's Client ID to manage keys and secrets.
func SetVaultPermissions(ctx context.Context, vaultName string) (keyvault.Vault, error) {
	vaultsClient := getVaultsClient()
	tenantID, err := uuid.FromString(iam.TenantID())
	if err != nil {
		return keyvault.Vault{}, err
	}

	clientID := iam.ClientID()

	return vaultsClient.CreateOrUpdate(
		ctx,
		helpers.ResourceGroupName(),
		vaultName,
		keyvault.VaultCreateOrUpdateParameters{
			Location: to.StringPtr(helpers.Location()),
			Properties: &keyvault.VaultProperties{
				TenantID: &tenantID,
				Sku: &keyvault.Sku{
					Family: to.StringPtr("A"),
					Name:   keyvault.Standard,
				},
				AccessPolicies: &[]keyvault.AccessPolicyEntry{
					keyvault.AccessPolicyEntry{
						ObjectID: &clientID,
						TenantID: &tenantID,
						Permissions: &keyvault.Permissions{
							Keys: &[]keyvault.KeyPermissions{
								keyvault.KeyPermissionsGet,
								keyvault.KeyPermissionsList,
								keyvault.KeyPermissionsCreate,
							},
							Secrets: &[]keyvault.SecretPermissions{
								keyvault.SecretPermissionsGet,
								keyvault.SecretPermissionsList,
							},
						},
					},
				},
			},
		},
	)
}

// SetVaultPermissionsForDeployment updates a key vault to enable deployments and add permissions to the application")
func SetVaultPermissionsForDeployment(ctx context.Context, vaultName string) (keyvault.Vault, error) {
	vaultsClient := getVaultsClient()
	tenantID, err := uuid.FromString(iam.TenantID())
	if err != nil {
		return keyvault.Vault{}, err
	}
	clientID := iam.ClientID()

	return vaultsClient.CreateOrUpdate(
		ctx,
		helpers.ResourceGroupName(),
		vaultName,
		keyvault.VaultCreateOrUpdateParameters{
			Location: to.StringPtr(helpers.Location()),
			Properties: &keyvault.VaultProperties{
				TenantID:                     &tenantID,
				EnabledForDeployment:         to.BoolPtr(true),
				EnabledForTemplateDeployment: to.BoolPtr(true),
				Sku: &keyvault.Sku{
					Family: to.StringPtr("A"),
					Name:   keyvault.Standard,
				},
				AccessPolicies: &[]keyvault.AccessPolicyEntry{
					keyvault.AccessPolicyEntry{
						ObjectID: to.StringPtr(clientID),
						TenantID: &tenantID,
						Permissions: &keyvault.Permissions{
							Keys: &[]keyvault.KeyPermissions{
								keyvault.KeyPermissionsGet,
								keyvault.KeyPermissionsList,
								keyvault.KeyPermissionsCreate,
							},
							Secrets: &[]keyvault.SecretPermissions{
								keyvault.SecretPermissionsGet,
								keyvault.SecretPermissionsSet,
								keyvault.SecretPermissionsList,
							},
						},
					},
				},
			},
		},
	)
}

// GetVaults lists all key vaults in a subscrition
func GetVaults(ctx context.Context) {
	vaultsClient := getVaultsClient()

	fmt.Println("Getting all vaults in subscription")
	subList, err := vaultsClient.List(ctx, "resourceType eq 'Microsoft.KeyVault/vaults'", nil)
	if err != nil {
		log.Printf("failed to get list of vaults: %v", err)
	}
	for _, kv := range subList.Values() {
		fmt.Printf("\t%s\n", *kv.Name)
	}

	fmt.Println("Getting all vaults in resource group")
	rgList, err := vaultsClient.ListByResourceGroup(ctx, helpers.ResourceGroupName(), nil)
	if err != nil {
		log.Printf("failed to get list of vaults: %v", err)
	}
	for _, kv := range rgList.Values() {
		fmt.Printf("\t%s\n", *kv.Name)
	}
}

// DeleteVault deletes an existing vault
func DeleteVault(ctx context.Context, vaultName string) (autorest.Response, error) {
	vaultsClient := getVaultsClient()
	return vaultsClient.Delete(ctx, helpers.ResourceGroupName(), vaultName)
}
