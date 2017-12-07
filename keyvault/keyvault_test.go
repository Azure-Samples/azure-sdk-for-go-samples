package keyvault

import (
	"flag"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/common"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/management"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
	"github.com/marstr/randname"
)

var (
	keyValutName string
)

func init() {
	management.GetStartParams()
	flag.StringVar(&keyValutName, "keyValutName", "valut"+randname.AdjNoun{}.Generate(), "Provide a name for the keyvault to be created")
	flag.Parse()
}
func ExampleSetVaultPermissions() {
	defer resources.Cleanup()

	_, err := resources.CreateGroup()
	if err != nil {
		common.PrintAndLog(err.Error())
	}
	common.PrintAndLog("resource group created")

	_, err = CreateVault(keyValutName)
	if err != nil {
		common.PrintAndLog(err.Error())
	}
	common.PrintAndLog("keyvault created")

	_, err = SetVaultPermissions(keyValutName)
	if err != nil {
		common.PrintAndLog(err.Error())
	}
	common.PrintAndLog("set keyvault permissions")

	// Output:
	// resource group created
	// keyvault created
	// set keyvault permissions
}
