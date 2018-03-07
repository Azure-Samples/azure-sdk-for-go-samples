// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package keyvault

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal"
	"github.com/Azure/azure-sdk-for-go/services/keyvault/mgmt/2016-10-01/keyvault"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	uuid "github.com/satori/go.uuid"
)

func getVaultsClient() keyvault.VaultsClient {
	token, _ := iam.GetResourceManagementToken(iam.AuthGrantType())
	vaultsClient := keyvault.NewVaultsClient(internal.SubscriptionID())
	vaultsClient.Authorizer = autorest.NewBearerAuthorizer(token)
	vaultsClient.AddToUserAgent(internal.UserAgent())
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
		internal.ResourceGroupName(),
		vaultName,
		keyvault.VaultCreateOrUpdateParameters{
			Location: to.StringPtr(internal.Location()),
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
	return vaultsClient.Get(ctx, internal.ResourceGroupName(), vaultName)
}

// CreateComplexKeyVault creates a new vault which grants access to the the current user and the service principal in use
func CreateComplexKeyVault(ctx context.Context, vaultName, userID string) (vault keyvault.Vault, err error) {
	vaultsClient := getVaultsClient()

	tenantID, err := uuid.FromString(iam.TenantID())
	if err != nil {
		return
	}

	apList := []keyvault.AccessPolicyEntry{}
	ap := keyvault.AccessPolicyEntry{
		TenantID: &tenantID,
		Permissions: &keyvault.Permissions{
			Keys: &[]keyvault.KeyPermissions{
				keyvault.KeyPermissionsCreate,
			},
			Secrets: &[]keyvault.SecretPermissions{
				keyvault.SecretPermissionsSet,
			},
		},
	}
	if userID != "" {
		ap.ObjectID = to.StringPtr(userID)
		apList = append(apList, ap)
	}
	if internal.ServicePrincipalObjectID() != "" {
		// This is the SP object ID, which is not the same as the AD app object ID
		// SP appID and AD app ID are the same values, aka, the client ID
		// You can get the SP objectID on the Azure CLI like this
		// az ad sp list --spn <AD app appID>
		ap.ObjectID = to.StringPtr(internal.ServicePrincipalObjectID())
		apList = append(apList, ap)
	}

	return vaultsClient.CreateOrUpdate(
		ctx,
		internal.ResourceGroupName(),
		vaultName,
		keyvault.VaultCreateOrUpdateParameters{
			Location: to.StringPtr(internal.Location()),
			Properties: &keyvault.VaultProperties{
				AccessPolicies:           &apList,
				EnabledForDiskEncryption: to.BoolPtr(true),
				Sku: &keyvault.Sku{
					Family: to.StringPtr("A"),
					Name:   keyvault.Standard,
				},
				TenantID: &tenantID,
			},
		})

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
		internal.ResourceGroupName(),
		vaultName,
		keyvault.VaultCreateOrUpdateParameters{
			Location: to.StringPtr(internal.Location()),
			Properties: &keyvault.VaultProperties{
				TenantID: &tenantID,
				Sku: &keyvault.Sku{
					Family: to.StringPtr("A"),
					Name:   keyvault.Standard,
				},
				AccessPolicies: &[]keyvault.AccessPolicyEntry{
					{
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
		internal.ResourceGroupName(),
		vaultName,
		keyvault.VaultCreateOrUpdateParameters{
			Location: to.StringPtr(internal.Location()),
			Properties: &keyvault.VaultProperties{
				TenantID:                     &tenantID,
				EnabledForDeployment:         to.BoolPtr(true),
				EnabledForTemplateDeployment: to.BoolPtr(true),
				Sku: &keyvault.Sku{
					Family: to.StringPtr("A"),
					Name:   keyvault.Standard,
				},
				AccessPolicies: &[]keyvault.AccessPolicyEntry{
					{
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

// GetVaults lists all key vaults in a subscription
func GetVaults() {
	vaultsClient := getVaultsClient()

	fmt.Println("Getting all vaults in subscription")
	for subList, err := vaultsClient.ListComplete(context.Background(), nil); subList.NotDone(); err = subList.Next() {
		if err != nil {
			log.Printf("failed to get list of vaults: %v", err)
		}
		fmt.Printf("\t%s\n", *subList.Value().Name)
	}

	fmt.Println("Getting all vaults in resource group")
	for rgList, err := vaultsClient.ListByResourceGroupComplete(context.Background(), internal.ResourceGroupName(), nil); rgList.NotDone(); err = rgList.Next() {
		if err != nil {
			log.Printf("failed to get list of vaults: %v", err)
		}
		fmt.Printf("\t%s\n", *rgList.Value().Name)
	}
}

// DeleteVault deletes an existing vault
func DeleteVault(ctx context.Context, vaultName string) (autorest.Response, error) {
	vaultsClient := getVaultsClient()
	return vaultsClient.Delete(ctx, internal.ResourceGroupName(), vaultName)
}
