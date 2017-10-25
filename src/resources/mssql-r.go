package resources

import (
  "fmt"
  "github.com/joshgav/az-go/src/util"
  "github.com/subosito/gotenv"

  "github.com/Azure/azure-sdk-for-go/arm/sql"
  "github.com/Azure/go-autorest/autorest"
  "github.com/Azure/go-autorest/autorest/to"
)

var (
  serverName string
  dbName string
  dbLogin string
  dbPassword string
)

func init() {
  gotenv.Load() // read from .env file

	serverName        = util.GetEnvVarOrFail("AZURE_SQL_SERVERNAME")
	dbName            = util.GetEnvVarOrFail("AZURE_SQL_DBNAME")
	dbLogin           = util.GetEnvVarOrFail("AZURE_SQL_DBUSER")
  dbPassword        = util.GetEnvVarOrFail("AZURE_SQL_DBPASSWORD")
}

func CreateServer() (<-chan sql.Server, <-chan error) {
  serversClient := sql.NewServersClient(SubscriptionId)
  serversClient.Authorizer = autorest.NewBearerAuthorizer(Token)

  return serversClient.CreateOrUpdate(
    ResourceGroupName,
    serverName,
    sql.Server{
      Location: to.StringPtr(Location),
      ServerProperties: &sql.ServerProperties{
        AdministratorLogin: to.StringPtr(dbLogin),
        AdministratorLoginPassword: to.StringPtr(dbPassword)}},
    nil)
}

func CreateDb() (<-chan sql.Database, <-chan error) {
  dbClient := sql.NewDatabasesClient(SubscriptionId)
  dbClient.Authorizer = autorest.NewBearerAuthorizer(Token)

  return dbClient.CreateOrUpdate(
    ResourceGroupName,
    serverName,
    dbName,
    sql.Database{
      Location: to.StringPtr(Location)},
    nil)
}

func OpenDbPort() (sql.FirewallRule, error) {
  fwRulesClient := sql.NewFirewallRulesClient(SubscriptionId)
  fwRulesClient.Authorizer = autorest.NewBearerAuthorizer(Token)

  return fwRulesClient.CreateOrUpdate(
    ResourceGroupName,
    serverName,
    "unsafe open to world",
    sql.FirewallRule{
      FirewallRuleProperties: &sql.FirewallRuleProperties{
        StartIPAddress: to.StringPtr("0.0.0.0"),
        EndIPAddress: to.StringPtr("255.255.255.255")}})
}

func DeleteDb() (autorest.Response, error) {
  dbClient := sql.NewDatabasesClient(SubscriptionId)
  dbClient.Authorizer = autorest.NewBearerAuthorizer(Token)

  return dbClient.Delete(
    ResourceGroupName,
    serverName,
    dbName)
}

func PrintInfo() {
  fmt.Printf("user agent string: %s\n", sql.UserAgent())
  fmt.Printf("SQL ARM Client version: %s\n", sql.Version())
}

