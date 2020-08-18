package mysqlsamples

import (
	"context"
	"fmt"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/gechris/azure-sdk-for-go-samples/internal/config"
	"github.com/gechris/azure-sdk-for-go-samples/internal/iam"
	"github.com/gechris/azure-sdk-for-go/services/preview/mysql/mgmt/flexible-servers/2020-07-01-privatepreview/mysql"
)

// GetServersClient returns
func GetServersClient() mysql.ServersClient {
	serversClient := mysql.NewServersClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	serversClient.Authorizer = a
	serversClient.AddToUserAgent(config.UserAgent())
	return serversClient
}

// CreateServer creates a new PostgreSQL Server
func CreateServer(ctx context.Context, serversClient mysql.ServersClient, serverName string, dbLogin string, dbPassword string) (server mysql.Server, err error) {

	// Create the server
	future, err := serversClient.Create(
		ctx,
		config.GroupName(),
		serverName,
		mysql.Server{
			Location: to.StringPtr(config.Location()),
			Sku: &mysql.Sku{
				Name: to.StringPtr("Standard_D4s_v3"),
			},
			ServerProperties: &mysql.ServerProperties{
				AdministratorLogin:         to.StringPtr(dbLogin),
				AdministratorLoginPassword: to.StringPtr(dbPassword),
				Version:                    mysql.FiveFullStopSeven, // 5.7
				StorageProfile: &mysql.StorageProfile{
					StorageMB: to.Int32Ptr(524288),
				},
			},
		})

	if err != nil {
		return server, fmt.Errorf("cannot create mysql server: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, serversClient.Client)
	if err != nil {
		return server, fmt.Errorf("cannot get the mysql server create or update future response: %v", err)
	}

	return future.Result(serversClient)
}

// UpdateServerStorageCapacity given the server name and the new storage capacity it updates the server's storage capacity.
func UpdateServerStorageCapacity(ctx context.Context, serversClient mysql.ServersClient, serverName string, storageCapacity int32) (server mysql.Server, err error) {

	future, err := serversClient.Update(
		ctx,
		config.GroupName(),
		serverName,
		mysql.ServerForUpdate{
			ServerPropertiesForUpdate: &mysql.ServerPropertiesForUpdate{
				StorageProfile: &mysql.StorageProfile{
					StorageMB: &storageCapacity,
				},
			},
		},
	)
	if err != nil {
		return server, fmt.Errorf("cannot update mysql server: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, serversClient.Client)
	if err != nil {
		return server, fmt.Errorf("cannot get the mysql server update future response: %v", err)
	}

	return future.Result(serversClient)
}

// DeleteServer deletes the PostgreSQL server.
func DeleteServer(ctx context.Context, serversClient mysql.ServersClient, serverName string) (resp autorest.Response, err error) {

	future, err := serversClient.Delete(ctx, config.GroupName(), serverName)
	if err != nil {
		return resp, fmt.Errorf("cannot delete the mysql server: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, serversClient.Client)
	if err != nil {
		return resp, fmt.Errorf("cannot get the mysql server update future response: %v", err)
	}

	return future.Result(serversClient)
}

// Firewall rules
func getFwRulesClient() mysql.FirewallRulesClient {
	fwrClient := mysql.NewFirewallRulesClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	fwrClient.Authorizer = a
	fwrClient.AddToUserAgent(config.UserAgent())
	return fwrClient
}

// CreateFirewallRules creates new firewall rules for a given server
func CreateFirewallRules(ctx context.Context, serverName string) error {
	fwrClient := getFwRulesClient()

	_, err := fwrClient.CreateOrUpdate(
		ctx,
		config.GroupName(),
		serverName,
		"unsafe open to world",
		mysql.FirewallRule{
			FirewallRuleProperties: &mysql.FirewallRuleProperties{
				StartIPAddress: to.StringPtr("0.0.0.0"),
				EndIPAddress:   to.StringPtr("255.255.255.255"),
			},
		},
	)
	if err != nil {
		return err
	}

	_, err = fwrClient.CreateOrUpdate(
		ctx,
		config.GroupName(),
		serverName,
		"open to Azure internal",
		mysql.FirewallRule{
			FirewallRuleProperties: &mysql.FirewallRuleProperties{
				StartIPAddress: to.StringPtr("0.0.0.0"),
				EndIPAddress:   to.StringPtr("0.0.0.0"),
			},
		},
	)

	return err
}

// GetConfigurationsClient creates and returns the configuration client for the server.
func GetConfigurationsClient() mysql.ConfigurationsClient {
	configClient := mysql.NewConfigurationsClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	configClient.Authorizer = a
	configClient.AddToUserAgent(config.UserAgent())
	return configClient
}

// GetConfiguration given the server name and configuration name it returns the configuration.
func GetConfiguration(ctx context.Context, configClient mysql.ConfigurationsClient, serverName string, configurationName string) (mysql.Configuration, error) {

	// Get the configuration.
	configuration, err := configClient.Get(ctx, config.GroupName(), serverName, configurationName)

	if err != nil {
		return configuration, fmt.Errorf("cannot get the configuration with name %s", configurationName)
	}

	return configuration, err
}

// UpdateConfiguration given the name of the configuation and the configuration object it updates the configuration for the given server.
func UpdateConfiguration(ctx context.Context, configClient mysql.ConfigurationsClient, serverName string, configurationName string, configuration mysql.Configuration) (updatedConfig mysql.Configuration, err error) {

	future, err := configClient.Update(ctx, config.GroupName(), serverName, configurationName, configuration)

	if err != nil {
		return updatedConfig, fmt.Errorf("cannot update the configuration with name %s", configurationName)
	}

	err = future.WaitForCompletionRef(ctx, configClient.Client)
	if err != nil {
		return updatedConfig, fmt.Errorf("cannot get the mysql configuration update future response: %v", err)
	}

	return future.Result(configClient)
}
