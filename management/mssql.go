package management

import (
	"github.com/joshgav/az-go/common"
	"log"

	"github.com/Azure/azure-sdk-for-go/profiles/preview/sql/mgmt/sql"
	"github.com/subosito/gotenv"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

var (
	serverName string
	dbName     string
	dbLogin    string
	dbPassword string
)

func init() {
	gotenv.Load() // read from .env file

	serverName = common.GetEnvVarOrFail("AZURE_SQL_SERVERNAME")
	dbName = common.GetEnvVarOrFail("AZURE_SQL_DBNAME")
	dbLogin = common.GetEnvVarOrFail("AZURE_SQL_DBUSER")
	dbPassword = common.GetEnvVarOrFail("AZURE_SQL_DBPASSWORD")
}

func CreateServer() (<-chan sql.Server, <-chan error) {
	token, err := common.GetResourceManagementToken(common.OAuthGrantTypeServicePrincipal)
	if err != nil {
		log.Fatalf("%s: %v", "failed to get auth token", err)
	}

	serversClient := sql.NewServersClient(subscriptionId)
	serversClient.Authorizer = autorest.NewBearerAuthorizer(token)

	return serversClient.CreateOrUpdate(
		resourceGroupName,
		serverName,
		sql.Server{
			Location: to.StringPtr(location),
			ServerProperties: &sql.ServerProperties{
				AdministratorLogin:         to.StringPtr(dbLogin),
				AdministratorLoginPassword: to.StringPtr(dbPassword)}},
		nil)
}

func CreateDb() (<-chan sql.Database, <-chan error) {
	token, err := common.GetResourceManagementToken(common.OAuthGrantTypeServicePrincipal)
	if err != nil {
		log.Fatalf("%s: %v", "failed to get auth token", err)
	}

	dbClient := sql.NewDatabasesClient(subscriptionId)
	dbClient.Authorizer = autorest.NewBearerAuthorizer(token)

	return dbClient.CreateOrUpdate(
		resourceGroupName,
		serverName,
		dbName,
		sql.Database{
			Location: to.StringPtr(location)},
		nil)
}

func OpenDbPort() error {
	token, err := common.GetResourceManagementToken(common.OAuthGrantTypeServicePrincipal)
	if err != nil {
		log.Fatalf("%s: %v", "failed to get auth token", err)
	}

	fwRulesClient := sql.NewFirewallRulesClient(subscriptionId)
	fwRulesClient.Authorizer = autorest.NewBearerAuthorizer(token)

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

func DeleteDb() (autorest.Response, error) {
	token, err := common.GetResourceManagementToken(common.OAuthGrantTypeServicePrincipal)
	if err != nil {
		log.Fatalf("%s: %v", "failed to get auth token", err)
	}

	dbClient := sql.NewDatabasesClient(subscriptionId)
	dbClient.Authorizer = autorest.NewBearerAuthorizer(token)

	return dbClient.Delete(
		resourceGroupName,
		serverName,
		dbName)
}

func PrintInfo() {
	log.Printf("user agent string: %s\n", sql.UserAgent())
	log.Printf("SQL ARM Client version: %s\n", sql.Version())
}
