package keyvault

import (
	"context"
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
	ctx := context.Background()

	defer resources.Cleanup(ctx)

	_, err := resources.CreateGroup(ctx, helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("resource group created")

	_, err = CreateVault(ctx, vaultName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("vault created")

	_, err = SetVaultPermissions(ctx, vaultName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("set vault permissions")

	// Output:
	// resource group created
	// vault created
	// set vault permissions
}
