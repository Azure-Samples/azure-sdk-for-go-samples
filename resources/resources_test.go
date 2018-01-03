package resources

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
)

func init() {
	helpers.ParseArgs()
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
