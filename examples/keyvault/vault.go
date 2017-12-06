package keyvault

import (
	"fmt"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/management"
	"github.com/Azure/azure-sdk-for-go/services/keyvault/mgmt/2016-10-01/keyvault"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/satori/go.uuid"
)

func getVaultsClient() keyvault.VaultsClient {
	vaultsClient := keyvault.NewVaultsClient(management.GetSubID())
	vaultsClient.Authorizer = management.GetToken()
	return vaultsClient
}

func CreateVault(vaultName string) (keyvault.Vault, error) {
	vaultsClient := getVaultsClient()
	tenantID, err := uuid.FromString(iam.GetTenantID())
	if err != nil {
		return keyvault.Vault{}, err
	}

	return vaultsClient.CreateOrUpdate(
		management.GetResourceGroup(),
		vaultName,
		keyvault.VaultCreateOrUpdateParameters{
			Location: to.StringPtr(management.GetLocation()),
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

func GetVault(vaultName string) (keyvault.Vault, error) {
	vaultsClient := getVaultsClient()
	return vaultsClient.Get(management.GetResourceGroup(), vaultName)
}

// SetVaultPermissions adds an access policy permitting this app's Client ID to manage keys and secrets.
func SetVaultPermissions(vaultName string) (keyvault.Vault, error) {
	vaultsClient := getVaultsClient()
	tenantID, err := uuid.FromString(iam.GetTenantID())
	if err != nil {
		return keyvault.Vault{}, err
	}

	clientID := iam.GetClientID()

	return vaultsClient.CreateOrUpdate(
		management.GetResourceGroup(),
		vaultName,
		keyvault.VaultCreateOrUpdateParameters{
			Location: to.StringPtr(management.GetLocation()),
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
func SetVaultPermissionsForDeployment(vaultName string) (keyvault.Vault, error) {
	vaultsClient := getVaultsClient()
	tenantID, err := uuid.FromString(iam.GetTenantID())
	if err != nil {
		return keyvault.Vault{}, err
	}
	clientID := iam.GetClientID()

	return vaultsClient.CreateOrUpdate(
		management.GetResourceGroup(),
		vaultName,
		keyvault.VaultCreateOrUpdateParameters{
			Location: to.StringPtr(management.GetLocation()),
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
func GetVaults() {
	vaultsClient := getVaultsClient()

	fmt.Println("Getting all vaults in subscription")
	subList, err := vaultsClient.List("resourceType eq 'Microsoft.KeyVault/vaults'", nil)
	if err != nil {
		log.Printf("failed to get list of vaults: %v", err)
	}
	for _, kv := range *subList.Value {
		fmt.Printf("\t%s\n", *kv.Name)
	}

	fmt.Println("Getting all vaults in resource group")
	rgList, err := vaultsClient.ListByResourceGroup(management.GetResourceGroup(), nil)
	if err != nil {
		log.Printf("failed to get list of vaults: %v", err)
	}
	for _, kv := range *rgList.Value {
		fmt.Printf("\t%s\n", *kv.Name)
	}
}

func DeleteVault(vaultName string) (autorest.Response, error) {
	vaultsClient := getVaultsClient()
	return vaultsClient.Delete(management.GetResourceGroup(), vaultName)
}
