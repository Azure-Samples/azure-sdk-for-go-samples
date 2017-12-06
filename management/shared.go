package management

import (
	"flag"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/common"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/go-autorest/autorest"
	"github.com/marstr/randname"
	"github.com/subosito/gotenv"
)

var (
	resourceGroupName string
	location          string
	subscriptionID    string
	tenantId          string
	token             autorest.Authorizer
	keepResources     bool
)

func GetStartParams() {
	gotenv.Load() // read from .env file

	subscriptionID = common.GetEnvVarOrFail("AZURE_SUBSCRIPTION_ID")
	tenantId = common.GetEnvVarOrFail("AZURE_TENANT_ID")

	flag.BoolVar(&keepResources, "keepResources", false, "Set the sample to keep or delete the created resources")
	flag.StringVar(&resourceGroupName, "rgName", "rg"+randname.AdjNoun{}.Generate(), "Provide a name for the resource group to be created")
	flag.StringVar(&location, "location", "westus", "Provide the Azure location where the resources will be be created")

	armToken, err := iam.GetResourceManagementToken(iam.OAuthGrantTypeServicePrincipal)
	if err != nil {
		log.Fatalf("%s: %v", "failed to get auth token", err)
	}
	token = autorest.NewBearerAuthorizer(armToken)
}

func KeepResources() bool {
	return keepResources
}

func GetToken() autorest.Authorizer {
	return token
}

func GetSubID() string {
	return subscriptionID
}

func GetResourceGroup() string {
	return resourceGroupName
}

func GetLocation() string {
	return location
}
