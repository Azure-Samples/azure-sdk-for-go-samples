package sql

import (
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/management"
	"github.com/Azure/azure-sdk-for-go/services/sql/mgmt/2015-05-01-preview/sql"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

// Servers

func getServersClient() sql.ServersClient {
	serversClient := sql.NewServersClient(management.GetSubID())
	serversClient.Authorizer = management.GetToken()
	return serversClient
}

func CreateServer(serverName, dbLogin, dbPassword string) (<-chan sql.Server, <-chan error) {
	serversClient := getServersClient()
	return serversClient.CreateOrUpdate(
		management.GetResourceGroup(),
		serverName,
		sql.Server{
			Location: to.StringPtr(management.GetLocation()),
			ServerProperties: &sql.ServerProperties{
				AdministratorLogin:         to.StringPtr(dbLogin),
				AdministratorLoginPassword: to.StringPtr(dbPassword),
			},
		},
		nil)
}

// Databases

func getDbClient() sql.DatabasesClient {
	dbClient := sql.NewDatabasesClient(management.GetSubID())
	dbClient.Authorizer = management.GetToken()
	return dbClient
}

func CreateDb(serverName, dbName string) (<-chan sql.Database, <-chan error) {
	dbClient := getDbClient()
	return dbClient.CreateOrUpdate(
		management.GetResourceGroup(),
		serverName,
		dbName,
		sql.Database{
			Location: to.StringPtr(management.GetLocation())},
		nil)
}

func DeleteDb(serverName, dbName string) (autorest.Response, error) {
	dbClient := getDbClient()
	return dbClient.Delete(
		management.GetResourceGroup(),
		serverName,
		dbName)
}

// Firewall rukes

func getFwRulesClient() sql.FirewallRulesClient {
	fwrClient := sql.NewFirewallRulesClient(management.GetSubID())
	fwrClient.Authorizer = management.GetToken()
	return fwrClient
}

func CreateFirewallRules(serverName string) error {
	fwrClient := getFwRulesClient()

	_, err := fwrClient.CreateOrUpdate(
		management.GetResourceGroup(),
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
		management.GetResourceGroup(),
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
