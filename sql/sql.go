package sql

import (
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/services/sql/mgmt/2015-05-01-preview/sql"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

// Servers

func getServersClient() sql.ServersClient {
	token, _ := iam.GetResourceManagementToken(iam.OAuthGrantTypeServicePrincipal)
	serversClient := sql.NewServersClient(helpers.SubscriptionID())
	serversClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return serversClient
}

func CreateServer(serverName, dbLogin, dbPassword string) (<-chan sql.Server, <-chan error) {
	serversClient := getServersClient()
	return serversClient.CreateOrUpdate(
		helpers.ResourceGroupName(),
		serverName,
		sql.Server{
			Location: to.StringPtr(helpers.Location()),
			ServerProperties: &sql.ServerProperties{
				AdministratorLogin:         to.StringPtr(dbLogin),
				AdministratorLoginPassword: to.StringPtr(dbPassword),
			},
		},
		nil)
}

// Databases

func getDbClient() sql.DatabasesClient {
	token, _ := iam.GetResourceManagementToken(iam.OAuthGrantTypeServicePrincipal)
	dbClient := sql.NewDatabasesClient(helpers.SubscriptionID())
	dbClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return dbClient
}

func CreateDb(serverName, dbName string) (<-chan sql.Database, <-chan error) {
	dbClient := getDbClient()
	return dbClient.CreateOrUpdate(
		helpers.ResourceGroupName(),
		serverName,
		dbName,
		sql.Database{
			Location: to.StringPtr(helpers.Location()),
		},
		nil)
}

func DeleteDb(serverName, dbName string) (autorest.Response, error) {
	dbClient := getDbClient()
	return dbClient.Delete(
		helpers.ResourceGroupName(),
		serverName,
		dbName,
	)
}

// Firewall rukes

func getFwRulesClient() sql.FirewallRulesClient {
	token, _ := iam.GetResourceManagementToken(iam.OAuthGrantTypeServicePrincipal)
	fwrClient := sql.NewFirewallRulesClient(helpers.SubscriptionID())
	fwrClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return fwrClient
}

func CreateFirewallRules(serverName string) error {
	fwrClient := getFwRulesClient()

	_, err := fwrClient.CreateOrUpdate(
		helpers.ResourceGroupName(),
		serverName,
		"unsafe open to world",
		sql.FirewallRule{
			FirewallRuleProperties: &sql.FirewallRuleProperties{
				StartIPAddress: to.StringPtr("0.0.0.0"),
				EndIPAddress:   to.StringPtr("255.255.255.255"),
			},
		},
	)
	if err != nil {
		return err
	}

	_, err = fwrClient.CreateOrUpdate(
		helpers.ResourceGroupName(),
		serverName,
		"open to Azure internal",
		sql.FirewallRule{
			FirewallRuleProperties: &sql.FirewallRuleProperties{
				StartIPAddress: to.StringPtr("0.0.0.0"),
				EndIPAddress:   to.StringPtr("0.0.0.0"),
			},
		},
	)

	return err
}

func PrintInfo() {
	log.Printf("user agent string: %s\n", sql.UserAgent())
	log.Printf("SQL ARM Client version: %s\n", sql.Version())
}
