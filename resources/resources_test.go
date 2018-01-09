package resources

import (
	"context"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
)

func init() {
	err := helpers.ParseArgs()
	if err != nil {
		log.Fatalf("cannot parse arguments: %v", err)
	}
}

func ExampleCreateGroup() {
	defer Cleanup(context.Background())

	_, err := CreateGroup(context.Background(), helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("resource group created")

	// Output:
	// resource group created
}
