package management

import (
	"log"

	"github.com/Azure/azure-sdk-for-go/profiles/preview/sql/mgmt/sql"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/subosito/gotenv"
)

func init() {
	gotenv.Load() // read from .env file
}

func CreateServer(serverName, dbLogin, dbPassword string) (<-chan sql.Server, <-chan error) {
	serversClient := sql.NewServersClient(subscriptionId)
	serversClient.Authorizer = token

	return serversClient.CreateOrUpdate(
		resourceGroupName,
		serverName,
		sql.Server{
			Location: to.StringPtr(location),
			ServerProperties: &sql.ServerProperties{
				AdministratorLogin:         to.StringPtr(dbLogin),
				AdministratorLoginPassword: to.StringPtr(dbPassword),
			},
		},
		nil)
}

func CreateDb(serverName, dbName string) (<-chan sql.Database, <-chan error) {
	dbClient := sql.NewDatabasesClient(subscriptionId)
	dbClient.Authorizer = token

	return dbClient.CreateOrUpdate(
		resourceGroupName,
		serverName,
		dbName,
		sql.Database{
			Location: to.StringPtr(location)},
		nil)
}

func OpenDbPort(serverName string) error {
	fwRulesClient := sql.NewFirewallRulesClient(subscriptionId)
	fwRulesClient.Authorizer = token

	_, _ = fwRulesClient.CreateOrUpdate(
		resourceGroupName,
		serverName,
		"unsafe open to world",
		sql.FirewallRule{
			FirewallRuleProperties: &sql.FirewallRuleProperties{
				StartIPAddress: to.StringPtr("0.0.0.0"),
				EndIPAddress:   to.StringPtr("255.255.255.255")}})

	_, err2 := fwRulesClient.CreateOrUpdate(
		resourceGroupName,
		serverName,
		"open to Azure internal",
		sql.FirewallRule{
			FirewallRuleProperties: &sql.FirewallRuleProperties{
				StartIPAddress: to.StringPtr("0.0.0.0"),
				EndIPAddress:   to.StringPtr("0.0.0.0")}})

	return err2
}

func DeleteDb(serverName, dbName string) (autorest.Response, error) {
	dbClient := sql.NewDatabasesClient(subscriptionId)
	dbClient.Authorizer = token

	return dbClient.Delete(
		resourceGroupName,
		serverName,
		dbName)
}

func PrintInfo() {
	log.Printf("user agent string: %s\n", sql.UserAgent())
	log.Printf("SQL ARM Client version: %s\n", sql.Version())
}
