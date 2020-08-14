package postgresqlsamples

import (
	"context"
	"fmt"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
	"github.com/Azure/go-autorest/autorest/to"
	pg "github.com/gechris/azure-sdk-for-go/services/preview/postgresql/mgmt/flexible-servers/2020-02-14-privatepreview/postgresql"
)

// GetServersClient returns
func GetServersClient() pg.ServersClient {
	serversClient := pg.NewServersClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	serversClient.Authorizer = a
	serversClient.AddToUserAgent(config.UserAgent())
	return serversClient
}

// CreateServer creates a new SQL Server
func CreateServer(ctx context.Context, serverName, dbLogin, dbPassword string) (server pg.Server, err error) {
	serversClient := GetServersClient()

	future, err := serversClient.Create(
		ctx,
		config.GroupName(),
		serverName,
		pg.Server{
			Location: to.StringPtr(config.Location()),
			ServerProperties: &pg.ServerProperties{
				AdministratorLogin:         to.StringPtr(dbLogin),
				AdministratorLoginPassword: to.StringPtr(dbPassword),
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

/* func getDbClient() pg.FlexibleServerDatabasesClient {
	dbClient := pg.NewDatabasesClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	dbClient.Authorizer = a
	dbClient.AddToUserAgent(config.UserAgent())
	return dbClient
} */

// CreateDB creates a new SQL Database on a given server
/* func CreateDB(ctx context.Context, serverName, dbName string) (db pg.Database, err error) {
	dbClient := getDbClient()
	future, err := dbClient.CreateOrUpdate(
		ctx,
		config.GroupName(),
		serverName,
		dbName,
		pg.Database{})
	if err != nil {
		return db, fmt.Errorf("cannot create sql database: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, dbClient.Client)
	if err != nil {
		return db, fmt.Errorf("cannot get the pg database create or update future response: %v", err)
	}

	return future.Result(dbClient)
}

// DeleteDB deletes an existing database from a server
func DeleteDB(ctx context.Context, serverName, dbName string) (autorest.Response, error) {
	dbClient := getDbClient()
	future, err := dbClient.Delete(
		ctx,
		config.GroupName(),
		serverName,
		dbName,
	)

	if err != nil {
		return autorest.Response{}, fmt.Errorf("cannot delete the database.")
	}

	err = future.WaitForCompletionRef(ctx, dbClient.Client)
	if err != nil {
		return autorest.Response{}, fmt.Errorf("cannot get the pg server create or update future response: %v", err)
	}

	return future.Result(dbClient)
} */

// Firewall rules
func getFwRulesClient() pg.FirewallRulesClient {
	fwrClient := pg.NewFirewallRulesClient(config.SubscriptionID())
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
		pg.FirewallRule{
			FirewallRuleProperties: &pg.FirewallRuleProperties{
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
		pg.FirewallRule{
			FirewallRuleProperties: &pg.FirewallRuleProperties{
				StartIPAddress: to.StringPtr("0.0.0.0"),
				EndIPAddress:   to.StringPtr("0.0.0.0"),
			},
		},
	)

	return err
}
