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
	flexibleservers "github.com/Azure/azure-sdk-for-go/services/preview/postgresql/mgmt/2020-02-14-preview/postgresqlflexibleservers"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/marstr/randname"
)

var (
	groupName  = config.GenerateGroupName("PgServerOperations")
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
	if !config.KeepResources() {
		// does not wait
		_, err := resources.DeleteGroup(context.Background(), groupName)
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

// Example_performServerOperations creates a postgresql server, updates it, add firewall rules and configurations and at the end it deletes it.
func Example_performServerOperations() {
	serverName = strings.ToLower(serverName)

	ctx := context.Background()
	defer resources.Cleanup(ctx)

	if _, err := resources.CreateGroup(ctx, groupName); err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("resource group created")

	// Create the server.
	if _, err := CreateServer(ctx, groupName, serverName, dbLogin, dbPassword); err != nil {
		util.LogAndPanic(fmt.Errorf("cannot create postgresql server: %+v", err))
	}
	util.PrintAndLog("postgresql server created")

	// Update the server's storage capacity field.
	if _, err := UpdateServerStorageCapacity(ctx, groupName, serverName, 1048576); err != nil {
		util.LogAndPanic(fmt.Errorf("cannot update postgresql server: %+v", err))
	}
	util.PrintAndLog("postgresql server's storage capacity updated")

	if _, err := CreateOrUpdateFirewallRule(ctx, groupName, serverName, "FirewallRuleName", "0.0.0.0", "0.0.0.0"); err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("firewall rule set created")

	if _, err := CreateOrUpdateFirewallRule(ctx, groupName, serverName, "FirewallRuleName", "0.0.0.0", "1.1.1.1"); err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("firewall rule updated")

	var configuration flexibleservers.Configuration

	configuration, err := GetConfiguration(ctx, groupName, serverName, "max_replication_slots")
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("got the max_replication_slots configuration")

	// Update the configuration Value.
	configuration.ConfigurationProperties.Value = to.StringPtr("20")
	configuration.ConfigurationProperties.Source = to.StringPtr("user-override")

	if _, err := UpdateConfiguration(ctx, groupName, serverName, "max_replication_slots", configuration); err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("max_replication_slots configuration updated")

	// Finally delete the server.
	if _, err := DeleteServer(ctx, groupName, serverName); err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("postgresql server deleted")

	// Output:
	// resource group created
	// postgresql server created
	// postgresql server's storage capacity updated
	// firewall rule set created
	// firewall rule updated
	// got the max_replication_slots configuration
	// max_replication_slots configuration updated
	// postgresql server deleted
}
