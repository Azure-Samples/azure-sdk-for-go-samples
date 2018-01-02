package sql

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

var (
	serverName = "sql-server-go-samples-" + helpers.GetRandomLetterSequence(10)
	dbName     = "sql-db1"
	dbLogin    = "sql-db-user1"
	dbPassword = "NoSoupForYou1!"
)

func init() {
	flag.StringVar(&serverName, "sqlServerName", serverName, "Provide a name for the SQL server to be created")
	flag.StringVar(&dbName, "sqlDbName", dbName, "Provide a name for the SQL database to be created")
	flag.StringVar(&dbLogin, "sqlDbUsername", dbLogin, "Provide a username for the SQL database.")
	flag.StringVar(&dbPassword, "sqlDbPassword", dbPassword, "Provide a password for the username.")
	helpers.ParseArgs()
}

// Example creates a SQL server and database, then creates a table and inserts a record.
func ExampleDatabaseQueries() {
	ctx := context.Background()

	defer resources.Cleanup(ctx)

	_, err := resources.CreateGroup(ctx, helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("resource group created")

	serverName = strings.ToLower(serverName)

	_, err = CreateServer(ctx, serverName, dbLogin, dbPassword)
	if err != nil {
		helpers.PrintAndLog(fmt.Sprintf("cannot create sql server: %v", err))
	}

	helpers.PrintAndLog("sql server created")

	_, err = CreateDB(ctx, serverName, dbName)
	if err != nil {
		helpers.PrintAndLog(fmt.Sprintf("cannot create sql database: %v", err))
	}
	helpers.PrintAndLog("database created")

	err = CreateFirewallRules(ctx, serverName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("database firewall rules set")

	err = DbOperations(serverName, dbName, dbLogin, dbPassword)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("database operations performed")

	// Output:
	// resource group created
	// sql server created
	// database created
	// database firewall rules set
	// database operations performed
}
