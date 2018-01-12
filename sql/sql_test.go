// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package sql

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

var (
	serverName = "sql-server-go-samples-" + helpers.GetRandomLetterSequence(10)
	dbName     = "sql-db1"
	dbLogin    = "sql-db-user1"
	dbPassword = "NoSoupForYou1!"
)

func TestMain(m *testing.M) {
	flag.StringVar(&serverName, "sqlServerName", serverName, "Provide a name for the SQL server to be created")
	flag.StringVar(&dbName, "sqlDbName", dbName, "Provide a name for the SQL database to be created")
	flag.StringVar(&dbLogin, "sqlDbUsername", dbLogin, "Provide a username for the SQL database.")
	flag.StringVar(&dbPassword, "sqlDbPassword", dbPassword, "Provide a password for the username.")

	err := helpers.ParseArgs()
	if err != nil {
		log.Fatalln("failed to parse args")
	}

	ctx := context.Background()
	defer resources.Cleanup(ctx)

	_, err = resources.CreateGroup(ctx, helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog(fmt.Sprintf("resource group created on location: %s", helpers.Location()))

	os.Exit(m.Run())
}

// Example creates a SQL server and database, then creates a table and inserts a record.
func ExampleDatabaseQueries() {
	ctx := context.Background()
	serverName = strings.ToLower(serverName)

	_, err := CreateServer(ctx, serverName, dbLogin, dbPassword)
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
	// sql server created
	// database created
	// database firewall rules set
	// database operations performed
}
