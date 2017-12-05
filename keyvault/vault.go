package keyvault

import (
	"fmt"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/services/keyvault/mgmt/2016-10-01/keyvault"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/marstr/randname"
	"github.com/satori/go.uuid"
	"github.com/subosito/gotenv"
)

var (
	resourceGroupName = helpers.ResourceGroupName
	vaultName1        = "kv-" + randname.AdjNoun{}.Generate()
	vaultName2        = "kv-" + randname.AdjNoun{}.Generate()

	clientID     string
	clientSecret string
)

func init() {
	gotenv.Load() // read from .env file

	clientID = helpers.GetEnvVarOrFail("AZURE_CLIENT_ID")
	clientSecret = helpers.GetEnvVarOrFail("AZURE_CLIENT_SECRET")
}

func getVaultsClient() keyvault.VaultsClient {
	token, err := iam.GetResourceManagementToken(iam.OAuthGrantTypeServicePrincipal)
	if err != nil {
		log.Fatalf("%s: %v", "failed to get auth token", err)
	}
	vaultsClient := keyvault.NewVaultsClient(helpers.SubscriptionID)
	vaultsClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return vaultsClient
}

func CreateVault() (keyvault.Vault, error) {
	vaultsClient := getVaultsClient()
	_tenantID, _ := uuid.FromString(helpers.TenantID)

	return vaultsClient.CreateOrUpdate(
		helpers.ResourceGroupName,
		vaultName1,
		keyvault.VaultCreateOrUpdateParameters{
			Location: to.StringPtr(helpers.Location),
			Properties: &keyvault.VaultProperties{
				TenantID: &_tenantID,
				Sku: &keyvault.Sku{
					Family: to.StringPtr("A"),
					Name:   keyvault.Standard,
				},
				AccessPolicies: &[]keyvault.AccessPolicyEntry{},
			},
		},
	)
}

func GetVault() (keyvault.Vault, error) {
	vaultsClient := getVaultsClient()
	return vaultsClient.Get(helpers.ResourceGroupName, vaultName1)
}

// SetVaultPermissions adds an access policy permitting this app's Client ID to manage keys and secrets.
func SetVaultPermissions() (keyvault.Vault, error) {
	vaultsClient := getVaultsClient()
	_tenantID, _ := uuid.FromString(helpers.TenantID)

	return vaultsClient.CreateOrUpdate(
		helpers.ResourceGroupName,
		vaultName1,
		keyvault.VaultCreateOrUpdateParameters{
			Location: to.StringPtr(helpers.Location),
			Properties: &keyvault.VaultProperties{
				TenantID: &_tenantID,
				Sku: &keyvault.Sku{
					Family: to.StringPtr("A"),
					Name:   keyvault.Standard,
				},
				AccessPolicies: &[]keyvault.AccessPolicyEntry{
					keyvault.AccessPolicyEntry{
						ObjectID: &clientID,
						TenantID: &_tenantID,
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
func SetVaultPermissionsForDeployment() (keyvault.Vault, error) {
	vaultsClient := getVaultsClient()
	_tenantID, _ := uuid.FromString(helpers.TenantID)

	return vaultsClient.CreateOrUpdate(
		helpers.ResourceGroupName,
		vaultName1,
		keyvault.VaultCreateOrUpdateParameters{
			Location: to.StringPtr(helpers.Location),
			Properties: &keyvault.VaultProperties{
				TenantID:                     &_tenantID,
				EnabledForDeployment:         to.BoolPtr(true),
				EnabledForTemplateDeployment: to.BoolPtr(true),
				Sku: &keyvault.Sku{
					Family: to.StringPtr("A"),
					Name:   keyvault.Standard,
				},
				AccessPolicies: &[]keyvault.AccessPolicyEntry{
					keyvault.AccessPolicyEntry{
						ObjectID: to.StringPtr(clientID),
						TenantID: &_tenantID,
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
	rgList, err := vaultsClient.ListByResourceGroup(helpers.ResourceGroupName, nil)
	if err != nil {
		log.Printf("failed to get list of vaults: %v", err)
	}
	for _, kv := range *rgList.Value {
		fmt.Printf("\t%s\n", *kv.Name)
	}
}

func DeleteVault() (autorest.Response, error) {
	vaultsClient := getVaultsClient()
	return vaultsClient.Delete(helpers.ResourceGroupName, vaultName1)
}
