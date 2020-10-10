package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
	"github.com/marstr/randname"
)

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

// test helpers
func generateName(prefix string) string {
	return strings.ToLower(randname.GenerateWithPrefix(prefix, 5))
}

// Just add 5 random digits at the end of the prefix password.
func generateResourceName(pass string) string {
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
	serviceClient := GetManagementServiceClient()
	var response, error = serviceClient.ListBySubscription(context.TODO())
	if(error!=nil){
		for _,resource := range response.Values(){
			fmt.Println("Name: ",  *resource.Name);
			fmt.Println("Provisioning state:", resource.ServiceProperties.ProvisioningState)
			fmt.Println("ImmutableResourceId", *resource.ServiceProperties.ImmutableResourceID)
		}
	}
}
