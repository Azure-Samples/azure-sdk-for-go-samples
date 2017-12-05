package management

import (
	"flag"
	"log"
	"os"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/common"
	"github.com/Azure/go-autorest/autorest"
	"github.com/subosito/gotenv"
)

var (
	subscriptionId    string
	tenantId          string
	resourceGroupName string
	location          string
	token             autorest.Authorizer
	keepResources     bool
)

func GetStartParams() {
	gotenv.Load() // read from .env file

	subscriptionId = common.GetEnvVarOrFail("AZURE_SUBSCRIPTION_ID")
	tenantId = common.GetEnvVarOrFail("AZURE_TENANT_ID")

	flag.StringVar(&resourceGroupName, "rgName", "rgname", "Provide a name for the resource group to be created")
	flag.StringVar(&location, "location", "westus", "Provide the Azure location where the resources will be be created")
	flag.BoolVar(&keepResources, "keepResources", false, "Set the sample to keep or delete the created resources")

	armToken, err := common.GetResourceManagementToken(common.OAuthGrantTypeServicePrincipal)
	if err != nil {
		log.Fatalf("%s: %v", "failed to get auth token", err)
	}
	token = autorest.NewBearerAuthorizer(armToken)
}

func KeepResourcesAndExit() {
	if keepResources {
		log.Printf("keeping resources and exit\n")
		os.Exit(0)
	}
}
