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
	containerGroupName string
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

	containerGroupName = os.Getenv("AZ_CONTAINERINSTANCE_CONTAINER_GROUP_NAME")
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

func ExampleGetContainerGroup() {
	defer resources.Cleanup()

	c, err := GetContainerGroup(helpers.ResourceGroupName(), containerGroupName)
	if err != nil {
		log.Fatalf("cannot get container group %v from resource group %v", containerGroupName, helpers.ResourceGroupName())
	}

	if *c.Name != containerGroupName {
		log.Fatalf("incorrect name of container group: expected %v, got %v", containerGroupName, *c.Name)
	}

	helpers.PrintAndLog("retrieved container group")

	// Output:
	// retrieved container group
}

func ExampleUpdateContainerGroup() {
	defer resources.Cleanup()

	_, err := UpdateContainerGroup(helpers.ResourceGroupName(), containerGroupName)
	if err != nil {
		log.Fatalf("cannot upate container group: %v", err)
	}

	helpers.PrintAndLog("updated container group")

	// Output:
	// updated container group
}

func ExampleDeleteContainerGroup() {
	defer resources.Cleanup()

	_, err := DeleteContainerGroup(helpers.ResourceGroupName(), containerGroupName)
	if err != nil {
		log.Fatalf("cannot delete container group %v from resource group %v: %v", containerGroupName, helpers.ResourceGroupName(), err)
	}

	helpers.PrintAndLog("deleted container group")

	// Output:
	// deleted container group
}
