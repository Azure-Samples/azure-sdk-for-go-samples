package keyvault

import (
	"flag"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

var (
	vaultName string
)

func init() {
	flag.StringVar(&vaultName, "vaultName", "vault-sample-go", "Specify name of vault to create.")
	helpers.ParseArgs()
}

func ExampleSetVaultPermissions() {
	defer resources.Cleanup()

	_, err := resources.CreateGroup(helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("resource group created")

	_, err = CreateVault(vaultName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("vault created")

	_, err = SetVaultPermissions(vaultName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("set vault permissions")

	// Output:
	// resource group created
	// vault created
	// set vault permissions
}
