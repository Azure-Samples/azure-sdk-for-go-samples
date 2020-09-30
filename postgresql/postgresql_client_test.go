package postgresql

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
	pg "github.com/Azure/azure-sdk-for-go/services/preview/postgresql/mgmt/2020-02-14-preview/postgresqlflexibleservers"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/marstr/randname"
)

var (
	serverName = generateName("gosdkpostgresql")
	dbName     = "postgresqldb1"
	dbLogin    = "postgresqldbuser1"
	dbPassword = generatePassword("postgresqldbuserpass!1")
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

	flag.StringVar(&serverName, "pgsqlServerName", serverName, "Name for PostgreSQL server.")
	flag.StringVar(&dbName, "pgsqlDbName", dbName, "Name for PostgreSQL database.")
	flag.StringVar(&dbLogin, "pgsqlDbUsername", dbLogin, "Username for PostgreSQL login.")
	flag.StringVar(&dbPassword, "pgsqlDbPassword", dbPassword, "Password for PostgreSQL login.")

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

// Just add 5 random digits at the end of the prefix password.
func generatePassword(pass string) string {
	return randname.GenerateWithPrefix(pass, 5)
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

// TestPerformServerOperations creates a postgresql server, updates it, add firewall rules and configurations and at the end it deletes it.
func TestPerformServerOperations(t *testing.T) {
	var groupName = config.GenerateGroupName("PgServerOperations")
	config.SetGroupName(groupName)

	serverName = strings.ToLower(serverName)

	ctx := context.Background()
	defer resources.Cleanup(ctx)

	_, err := resources.CreateGroup(ctx, config.GroupName())
	if err != nil {
		util.LogAndPanic(err)
	}

	// Get the ServersClient.
	serversClient := GetServersClient()

	// Create the server.
	_, err = CreateServer(ctx, serversClient, serverName, dbLogin, dbPassword)
	if err != nil {
		util.LogAndPanic(fmt.Errorf("cannot create postgresql server: %+v", err))
	}
	util.PrintAndLog("postgresql server created")

	// Update the server's storage capacity field.
	_, err = UpdateServerStorageCapacity(ctx, serversClient, serverName, 1048576)
	if err != nil {
		util.LogAndPanic(fmt.Errorf("cannot update postgresql server: %+v", err))
	}
	util.PrintAndLog("postgresql server's storage capacity updated.")

	fwrClient := GetFwRulesClient()

	err = CreateOrUpdateFirewallRule(ctx, fwrClient, serverName, "FirewallRuleName", "0.0.0.0", "0.0.0.0")
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("Firewall rule set")

	err = CreateOrUpdateFirewallRule(ctx, fwrClient, serverName, "FirewallRuleName", "0.0.0.0", "1.1.1.1")
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("Firewall rule updated")

	configClient := GetConfigurationsClient()

	var configuration pg.Configuration

	configuration, err = GetConfiguration(ctx, configClient, serverName, "array_nulls")
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("Got the array_nulls configuration")

	// Update the configuration Value.
	configuration.ConfigurationProperties.Value = to.StringPtr("on")

	_, err = UpdateConfiguration(ctx, configClient, serverName, "array_nulls", configuration)
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
