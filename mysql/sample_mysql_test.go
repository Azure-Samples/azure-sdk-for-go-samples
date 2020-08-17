package mysqlsamples

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/gechris/azure-sdk-for-go/services/preview/mysql/mgmt/flexible-servers/2020-07-01-privatepreview/mysql"
	"github.com/marstr/randname"
)

var (
	serverName = generateName("gosdkmysql")
	dbName     = "mysqldb1"
	dbLogin    = "mysqldbuser1"
	dbPassword = "mysqldbuserpass!1"
)

func addLocalEnvAndParse() error {
	// parse env at top-level (also controls dotenv load)
	err := config.ParseEnvironment()
	if err != nil {
		return fmt.Errorf("failed to add top-level env: %+v", err)
	}
	return nil
}

func addLocalFlagsAndParse() error {
	// add top-level flags
	err := config.AddFlags()
	if err != nil {
		return fmt.Errorf("failed to add top-level flags: %+v", err)
	}

	flag.StringVar(&serverName, "sqlServerName", serverName, "Name for SQL server.")
	flag.StringVar(&dbName, "sqlDbName", dbName, "Name for SQL database.")
	flag.StringVar(&dbLogin, "sqlDbUsername", dbLogin, "Username for SQL login.")
	flag.StringVar(&dbPassword, "sqlDbPassword", dbPassword, "Password for SQL login.")

	// parse all flags
	flag.Parse()
	return nil
}

func setup() error {
	var err error
	err = addLocalEnvAndParse()
	if err != nil {
		return err
	}
	err = addLocalFlagsAndParse()
	if err != nil {
		return err
	}

	return nil
}

func teardown() error {
	if config.KeepResources() == false {
		// does not wait
		_, err := resources.DeleteGroup(context.Background(), config.GroupName())
		if err != nil {
			return err
		}
	}
	return nil
}

// test helpers
func generateName(prefix string) string {
	return strings.ToLower(randname.GenerateWithPrefix(prefix, 5))
}

// TestMain sets up the environment and initiates tests.
func TestMain(m *testing.M) {
	var err error
	var code int

	err = setup()
	if err != nil {
		log.Fatalf("could not set up environment: %+v", err)
	}

	code = m.Run()

	err = teardown()
	if err != nil {
		log.Fatalf(
			"could not tear down environment: %v\n; original exit code: %v\n",
			err, code)
	}

	os.Exit(code)
}

// Example_createDatabase creates a MySQL server and database, then creates a table and inserts a record.
func Example_PerformServerOperations() {
	var groupName = config.GenerateGroupName("DatabaseQueries")
	config.SetGroupName(groupName)

	serverName = strings.ToLower(serverName)

	ctx := context.Background()
	defer resources.Cleanup(ctx)

	_, err := resources.CreateGroup(ctx, config.GroupName())
	if err != nil {
		util.LogAndPanic(err)
	}

	serversClient := GetServersClient()

	_, err = CreateServer(ctx, serversClient, serverName, dbLogin, dbPassword)
	if err != nil {
		util.LogAndPanic(fmt.Errorf("cannot create mysql server: %+v", err))
	}
	util.PrintAndLog("mysql server created")

	_, err = UpdateServerStorageCapacity(ctx, serversClient, serverName, 1048576)
	if err != nil {
		util.LogAndPanic(fmt.Errorf("cannot update mysql server: %+v", err))
	}
	util.PrintAndLog("updated mysql server's storage capacity")

	err = CreateFirewallRules(ctx, serverName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("database firewall rules set")

	configClient := GetConfigurationsClient()

	var configuration mysql.Configuration

	configuration, err = GetConfiguration(ctx, configClient, serverName, "event_scheduler")
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("Got the event_scheduler configuration")

	// Update the configuration Value.
	configuration.ConfigurationProperties.Value = to.StringPtr("on")

	_, err = UpdateConfiguration(ctx, configClient, serverName, "event_scheduler", configuration)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("Updated the event_scheduler configuration")

	// Finally delete the server.
	_, err = DeleteServer(ctx, serversClient, serverName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("Successfully deleted the server")
}
