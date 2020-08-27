package postgresql

import (
	"context"
	"fmt"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/gechris/azure-sdk-for-go-samples/internal/config"
	"github.com/gechris/azure-sdk-for-go-samples/internal/iam"
	pg "github.com/gechris/azure-sdk-for-go/services/preview/postgresql/mgmt/2020-02-14-preview/postgresqlflexibleservers"
)

// GetServersClient returns
func GetServersClient() pg.ServersClient {
	serversClient := pg.NewServersClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	serversClient.Authorizer = a
	serversClient.AddToUserAgent(config.UserAgent())
	return serversClient
}

// CreateServer creates a new PostgreSQL Server
func CreateServer(ctx context.Context, serversClient pg.ServersClient, serverName string, dbLogin string, dbPassword string) (server pg.Server, err error) {

	// Create the server
	future, err := serversClient.Create(
		ctx,
		config.GroupName(),
		serverName,
		pg.Server{
			Location: to.StringPtr(config.Location()),
			Sku: &pg.Sku{
				Name: to.StringPtr("Standard_D4s_v3"),
				Tier: "GeneralPurpose",
			},
			ServerProperties: &pg.ServerProperties{
				AdministratorLogin:         to.StringPtr(dbLogin),
				AdministratorLoginPassword: to.StringPtr(dbPassword),
				Version:                    pg.OneTwo,
				StorageProfile: &pg.StorageProfile{
					StorageMB: to.Int32Ptr(524288),
				},
			},
		})

	if err != nil {
		return server, fmt.Errorf("cannot create pg server: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, serversClient.Client)
	if err != nil {
		return server, fmt.Errorf("cannot get the pg server create or update future response: %v", err)
	}

	return future.Result(serversClient)
}

// UpdateServerStorageCapacity given the server name and the new storage capacity it updates the server's storage capacity.
func UpdateServerStorageCapacity(ctx context.Context, serversClient pg.ServersClient, serverName string, storageCapacity int32) (server pg.Server, err error) {

	future, err := serversClient.Update(
		ctx,
		config.GroupName(),
		serverName,
		pg.ServerForUpdate{
			Properties: &pg.ServerPropertiesForUpdate{
				StorageProfile: &pg.StorageProfile{
					StorageMB: &storageCapacity,
				},
			},
		},
	)
	if err != nil {
		return server, fmt.Errorf("cannot update pg server: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, serversClient.Client)
	if err != nil {
		return server, fmt.Errorf("cannot get the pg server update future response: %v", err)
	}

	return future.Result(serversClient)
}

// DeleteServer deletes the PostgreSQL server.
func DeleteServer(ctx context.Context, serversClient pg.ServersClient, serverName string) (resp autorest.Response, err error) {

	future, err := serversClient.Delete(ctx, config.GroupName(), serverName)
	if err != nil {
		return resp, fmt.Errorf("cannot delete the pg server: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, serversClient.Client)
	if err != nil {
		return resp, fmt.Errorf("cannot get the pg server update future response: %v", err)
	}

	return future.Result(serversClient)
}

// GetFwRulesClient returns the FirewallClient
func GetFwRulesClient() pg.FirewallRulesClient {
	fwrClient := pg.NewFirewallRulesClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	fwrClient.Authorizer = a
	fwrClient.AddToUserAgent(config.UserAgent())
	return fwrClient
}

// CreateOrUpdateFirewallRule given the firewallname and new properties it updates the firewall rule.
func CreateOrUpdateFirewallRule(ctx context.Context, fwrClient pg.FirewallRulesClient, serverName, firewallRuleName, startIPAddr, endIPAddr string) error {

	_, err := fwrClient.CreateOrUpdate(
		ctx,
		config.GroupName(),
		serverName,
		firewallRuleName,
		pg.FirewallRule{
			FirewallRuleProperties: &pg.FirewallRuleProperties{
				StartIPAddress: &startIPAddr,
				EndIPAddress:   &endIPAddr,
			},
		},
	)

	return err
}

// GetConfigurationsClient creates and returns the configuration client for the server.
func GetConfigurationsClient() pg.ConfigurationsClient {
	configClient := pg.NewConfigurationsClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	configClient.Authorizer = a
	configClient.AddToUserAgent(config.UserAgent())
	return configClient
}

// GetConfiguration given the server name and configuration name it returns the configuration.
func GetConfiguration(ctx context.Context, configClient pg.ConfigurationsClient, serverName string, configurationName string) (pg.Configuration, error) {

	// Get the configuration.
	configuration, err := configClient.Get(ctx, config.GroupName(), serverName, configurationName)

	if err != nil {
		return configuration, fmt.Errorf("cannot get the configuration with name %s", configurationName)
	}

	return configuration, err
}

// UpdateConfiguration given the name of the configuation and the configuration object it updates the configuration for the given server.
func UpdateConfiguration(ctx context.Context, configClient pg.ConfigurationsClient, serverName string, configurationName string, configuration pg.Configuration) (updatedConfig pg.Configuration, err error) {

	future, err := configClient.Update(ctx, config.GroupName(), serverName, configurationName, configuration)

	if err != nil {
		return updatedConfig, fmt.Errorf("cannot update the configuration with name %s", configurationName)
	}

	err = future.WaitForCompletionRef(ctx, configClient.Client)
	if err != nil {
		return updatedConfig, fmt.Errorf("cannot get the pg configuration update future response: %v", err)
	}

	return future.Result(configClient)
}
