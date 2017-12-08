package resources

import (
	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
)

func init() {
	helpers.ParseArgs()
}

func ExampleCreateGroup() {
	defer Cleanup()

	_, err := CreateGroup(helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("resource group created")

	// Output:
	// resource group created
}
