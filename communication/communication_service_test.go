package communication

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
	"github.com/marstr/randname"
)

var (
	serviceName = generateName("comm-service1")
)

// Example_createCommunicationService creates a group needed for a communication service, then
// creates a communication service.
func Example_createCommunicationService() {
	groupName := config.GenerateGroupName("communication")
	config.SetGroupName(groupName)

	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()
	defer resources.Cleanup(ctx)

	if _, err := resources.CreateGroup(ctx, groupName); err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created resource group")

	if _, err := CreateCommunicationService(ctx, groupName, serviceName); err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created communication service")

	// Output:
	// created resource group
	// created communication service
}

// test helpers
func generateName(prefix string) string {
	return strings.ToLower(randname.GenerateWithPrefix(prefix, 5))
}

func addLocalEnvAndParse() error {
	// parse env at top-level (also controls dotenv load)
	err := config.ParseEnvironment()
	if err != nil {
		return fmt.Errorf("failed to add top-level env: %+v", err)
	}
	return nil
}

func setup() error {
	var err error
	err = addLocalEnvAndParse()
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
