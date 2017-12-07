package sql

import (
	"flag"
	"strings"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/common"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/management"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
	"github.com/marstr/randname"
)

var (
	serverName string
	dbName     string
	dbLogin    string
	dbPassword string
)

func init() {
	management.GetStartParams()
	flag.StringVar(&serverName, "sqlServerName", "server"+randname.AdjNoun{}.Generate(), "Provide a name for the SQL server name to be created")
	flag.StringVar(&dbName, "sqlDbName", "db"+randname.AdjNoun{}.Generate(), "Provide a name for the SQL data basename to be created")
	flag.StringVar(&dbLogin, "sqlDbUserName", "user"+randname.AdjNoun{}.Generate(), "Provide a name for the SQL database username")
	flag.StringVar(&dbPassword, "sqlDbPassword", randname.AdjNoun{}.Generate(), "Provide a name for the SQL database password")
	flag.Parse()
}

// Example creates a SQL server and database, then creates a table and inserts a record.
func ExampleDatabaseQueries() {
	defer resources.Cleanup()

	_, err := resources.CreateGroup()
	if err != nil {
		common.PrintAndLog(err.Error())
	}
	common.PrintAndLog("resource group created")

	serverName = strings.ToLower(serverName)

	_, errC := CreateServer(serverName, dbLogin, dbPassword)
	err = <-errC
	if err != nil {
		common.PrintAndLog(err.Error())
	}
	common.PrintAndLog("sql server created")

	_, errC = CreateDb(serverName, dbName)
	err = <-errC
	if err != nil {
		common.PrintAndLog(err.Error())
	}
	common.PrintAndLog("database created")

	err = CreateFirewallRules(serverName)
	if err != nil {
		common.PrintAndLog(err.Error())
	}
	common.PrintAndLog("database firewall rules set")

	err = DbOperations(serverName, dbName, dbLogin, dbPassword)
	if err != nil {
		common.PrintAndLog(err.Error())
	}
	common.PrintAndLog("database operations performed")

	// Output:
	// resource group created
	// sql server created
	// database created
	// database firewall rules set
	// database operations performed
}
