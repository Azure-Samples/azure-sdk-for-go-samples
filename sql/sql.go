package mssql

import (
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"

	"github.com/Azure/azure-sdk-for-go/services/sql/mgmt/2015-05-01-preview/sql"
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

	serverName = helpers.GetEnvVarOrFail("AZURE_SQL_SERVERNAME")
	dbName = helpers.GetEnvVarOrFail("AZURE_SQL_DBNAME")
	dbLogin = helpers.GetEnvVarOrFail("AZURE_SQL_DBUSER")
	dbPassword = helpers.GetEnvVarOrFail("AZURE_SQL_DBPASSWORD")
}

func CreateServer() (<-chan sql.Server, <-chan error) {
	token, err := iam.GetResourceManagementToken(iam.OAuthGrantTypeServicePrincipal)
	if err != nil {
		log.Fatalf("%s: %v", "failed to get auth token", err)
	}

	serversClient := sql.NewServersClient(helpers.SubscriptionID)
	serversClient.Authorizer = autorest.NewBearerAuthorizer(token)

	return serversClient.CreateOrUpdate(
		helpers.ResourceGroupName,
		serverName,
		sql.Server{
			Location: to.StringPtr(helpers.Location),
			ServerProperties: &sql.ServerProperties{
				AdministratorLogin:         to.StringPtr(dbLogin),
				AdministratorLoginPassword: to.StringPtr(dbPassword)}},
		nil)
}

func CreateDb() (<-chan sql.Database, <-chan error) {
	token, err := iam.GetResourceManagementToken(iam.OAuthGrantTypeServicePrincipal)
	if err != nil {
		log.Fatalf("%s: %v", "failed to get auth token", err)
	}

	dbClient := sql.NewDatabasesClient(helpers.SubscriptionID)
	dbClient.Authorizer = autorest.NewBearerAuthorizer(token)

	return dbClient.CreateOrUpdate(
		helpers.ResourceGroupName,
		serverName,
		dbName,
		sql.Database{
			Location: to.StringPtr(helpers.Location)},
		nil)
}

func OpenDbPort() error {
	token, err := iam.GetResourceManagementToken(iam.OAuthGrantTypeServicePrincipal)
	if err != nil {
		log.Fatalf("%s: %v", "failed to get auth token", err)
	}

	fwRulesClient := sql.NewFirewallRulesClient(helpers.SubscriptionID)
	fwRulesClient.Authorizer = autorest.NewBearerAuthorizer(token)

	_, _ = fwRulesClient.CreateOrUpdate(
		helpers.ResourceGroupName,
		serverName,
		"unsafe open to world",
		sql.FirewallRule{
			FirewallRuleProperties: &sql.FirewallRuleProperties{
				StartIPAddress: to.StringPtr("0.0.0.0"),
				EndIPAddress:   to.StringPtr("255.255.255.255")}})

	_, err2 := fwRulesClient.CreateOrUpdate(
		helpers.ResourceGroupName,
		serverName,
		"open to Azure internal",
		sql.FirewallRule{
			FirewallRuleProperties: &sql.FirewallRuleProperties{
				StartIPAddress: to.StringPtr("0.0.0.0"),
				EndIPAddress:   to.StringPtr("0.0.0.0")}})

	return err2
}

func DeleteDb() (autorest.Response, error) {
	token, err := iam.GetResourceManagementToken(iam.OAuthGrantTypeServicePrincipal)
	if err != nil {
		log.Fatalf("%s: %v", "failed to get auth token", err)
	}

	dbClient := sql.NewDatabasesClient(helpers.SubscriptionID)
	dbClient.Authorizer = autorest.NewBearerAuthorizer(token)

	return dbClient.Delete(
		helpers.ResourceGroupName,
		serverName,
		dbName)
}

func PrintInfo() {
	log.Printf("user agent string: %s\n", sql.UserAgent())
	log.Printf("SQL ARM Client version: %s\n", sql.Version())
}
