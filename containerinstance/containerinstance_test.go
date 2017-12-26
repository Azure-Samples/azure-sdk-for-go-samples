package containerinstance

import (
	"fmt"
	"log"
	"os"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
	"github.com/subosito/gotenv"
)

var (
	containerGroupName = "az-samples-go-container-group-" + helpers.GetRandomLetterSequence(10)
)

func init() {
	err := parseArgs()
	if err != nil {
		log.Fatalf("cannot parse arguments: %v", err)
	}

}

func parseArgs() error {
	err := gotenv.Load()
	if err != nil {
		return fmt.Errorf("cannot load env: %v", err)
	}

	err = helpers.ParseArgs()
	if err != nil {
		return fmt.Errorf("cannot parse args: %v", err)
	}

	containerGroupName := os.Getenv("AZ_CONTAINERINSTANCE_NAME")
	if !(len(containerGroupName) > 0) {
		containerGroupName = "az-samples-go-container-group-" + helpers.GetRandomLetterSequence(10)
	}

	return nil
}

func ExampleCreateContainerGroup() {
	defer resources.Cleanup()

	_, err := resources.CreateGroup(helpers.ResourceGroupName())
	if err != nil {
		log.Printf("cannot create resource group: %v", err)
	}
	helpers.PrintAndLog("created resource group")

	_, err = CreateContainerGroup(containerGroupName, helpers.Location(), helpers.ResourceGroupName())
	if err != nil {
		log.Fatalf("cannot create container group: %v", err)
	}

	helpers.PrintAndLog("created container group")

	// Output:
	// created resource group
	// created container group
}
