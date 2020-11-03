package postgresql

import (
	"context"
	"fmt"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
	flexibleservers "github.com/Azure/azure-sdk-for-go/services/preview/postgresql/mgmt/2020-02-14-preview/postgresqlflexibleservers"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

// GetServersClient returns
func getServersClient() flexibleservers.ServersClient {
	serversClient := flexibleservers.NewServersClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	serversClient.Authorizer = a
	serversClient.AddToUserAgent(config.UserAgent())
	return serversClient
}

// CreateServer creates a new PostgreSQL Server
func CreateServer(ctx context.Context, resourceGroup, serverName, dbLogin, dbPassword string) (server flexibleservers.Server, err error) {
	serversClient := getServersClient()

	// Create the server
	future, err := serversClient.Create(
		ctx,
		resourceGroup,
		serverName,
		flexibleservers.Server{
			Location: to.StringPtr(config.Location()),
			Sku: &flexibleservers.Sku{
				Name: to.StringPtr("Standard_D4s_v3"),
				Tier: "GeneralPurpose",
			},
			ServerProperties: &flexibleservers.ServerProperties{
				AdministratorLogin:         to.StringPtr(dbLogin),
				AdministratorLoginPassword: to.StringPtr(dbPassword),
				Version:                    flexibleservers.OneTwo,
				StorageProfile: &flexibleservers.StorageProfile{
					StorageMB: to.Int32Ptr(524288),
				},
			},
		})

	if err != nil {
		return server, fmt.Errorf("cannot create pg server: %+v", err)
	}

	if err := future.WaitForCompletionRef(ctx, serversClient.Client); err != nil {
		return server, fmt.Errorf("cannot get the pg server create or update future response: %+v", err)
	}

	return future.Result(serversClient)
}

// UpdateServerStorageCapacity given the server name and the new storage capacity it updates the server's storage capacity.
func UpdateServerStorageCapacity(ctx context.Context, resourceGroup, serverName string, storageCapacity int32) (server flexibleservers.Server, err error) {
	serversClient := getServersClient()

	future, err := serversClient.Update(
		ctx,
		resourceGroup,
		serverName,
		flexibleservers.ServerForUpdate{
			ServerPropertiesForUpdate: &flexibleservers.ServerPropertiesForUpdate{
				StorageProfile: &flexibleservers.StorageProfile{
					StorageMB: &storageCapacity,
				},
			},
		},
	)
	if err != nil {
		return server, fmt.Errorf("cannot update pg server: %+v", err)
	}

	if err := future.WaitForCompletionRef(ctx, serversClient.Client); err != nil {
		return server, fmt.Errorf("cannot get the pg server update future response: %+v", err)
	}

	return future.Result(serversClient)
}

// DeleteServer deletes the PostgreSQL server.
func DeleteServer(ctx context.Context, resourceGroup, serverName string) (resp autorest.Response, err error) {
	serversClient := getServersClient()

	future, err := serversClient.Delete(ctx, resourceGroup, serverName)
	if err != nil {
		return resp, fmt.Errorf("cannot delete the pg server: %+v", err)
	}

	if err := future.WaitForCompletionRef(ctx, serversClient.Client); err != nil {
		return resp, fmt.Errorf("cannot get the pg server update future response: %+v", err)
	}

	return future.Result(serversClient)
}

// GetFwRulesClient returns the FirewallClient
func getFwRulesClient() flexibleservers.FirewallRulesClient {
	fwrClient := flexibleservers.NewFirewallRulesClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	fwrClient.Authorizer = a
	fwrClient.AddToUserAgent(config.UserAgent())
	return fwrClient
}

// CreateOrUpdateFirewallRule given the firewallname and new properties it updates the firewall rule.
func CreateOrUpdateFirewallRule(ctx context.Context, resourceGroup, serverName, firewallRuleName, startIPAddr, endIPAddr string) (rule flexibleservers.FirewallRule, err error) {
	fwrClient := getFwRulesClient()

	future, err := fwrClient.CreateOrUpdate(
		ctx,
		resourceGroup,
		serverName,
		firewallRuleName,
		flexibleservers.FirewallRule{
			FirewallRuleProperties: &flexibleservers.FirewallRuleProperties{
				StartIPAddress: &startIPAddr,
				EndIPAddress:   &endIPAddr,
			},
		},
	)
	if err != nil {
		return rule, fmt.Errorf("cannot create the firewall rule: %+v", err)
	}
	if err := future.WaitForCompletionRef(ctx, fwrClient.Client); err != nil {
		return rule, fmt.Errorf("cannot get the firewall rule create or update future response: %+v", err)
	}

	return future.Result(fwrClient)
}

// GetConfigurationsClient creates and returns the configuration client for the server.
func getConfigurationsClient() flexibleservers.ConfigurationsClient {
	configClient := flexibleservers.NewConfigurationsClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	configClient.Authorizer = a
	configClient.AddToUserAgent(config.UserAgent())
	return configClient
}

// GetConfiguration given the server name and configuration name it returns the configuration.
func GetConfiguration(ctx context.Context, resourceGroup, serverName, configurationName string) (flexibleservers.Configuration, error) {
	configClient := getConfigurationsClient()
	return configClient.Get(ctx, resourceGroup, serverName, configurationName)
}

// UpdateConfiguration given the name of the configuation and the configuration object it updates the configuration for the given server.
func UpdateConfiguration(ctx context.Context, resourceGroup, serverName string, configurationName string, configuration flexibleservers.Configuration) (updatedConfig flexibleservers.Configuration, err error) {
	configClient := getConfigurationsClient()

	future, err := configClient.Update(ctx, resourceGroup, serverName, configurationName, configuration)
	if err != nil {
		return updatedConfig, fmt.Errorf("cannot update the configuration with name %s: %+v", configurationName, err)
	}

	if err := future.WaitForCompletionRef(ctx, configClient.Client); err != nil {
		return updatedConfig, fmt.Errorf("cannot get the pg configuration update future response: %+v", err)
	}

	return future.Result(configClient)
}
