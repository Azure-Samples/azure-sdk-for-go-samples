package keyvault

import (
	"context"
	"flag"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

var (
	vaultName = "vault-sample-go-" + helpers.GetRandomLetterSequence(5)
)

func init() {
	flag.StringVar(&vaultName, "vaultName", vaultName, "Specify name of vault to create.")

	err := helpers.ParseArgs()
	if err != nil {
		log.Fatalln("failed to parse args")
	}
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
