package config

import (
	"log"
	"strings"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
)

const allAzureLocations = []string{
	"eastasia",
	"southeastasia",
	"centralus",
	"eastus",
	"eastus2",
	"westus",
	"northcentralus",
	"southcentralus",
	"northeurope",
	"westeurope",
	"japanwest",
	"japaneast",
	"brazilsouth",
	"australiaeast",
	"australiasoutheast",
	"southindia",
	"centralindia",
	"westindia",
	"canadacentral",
	"canadaeast",
	"uksouth",
	"ukwest",
	"westcentralus",
	"westus2",
	"koreacentral",
	"koreasouth",
}

const AzureCanaryLocation = "eastus2euap"

// Available returns whether a given resource type/version is available in the
// specifed location.
// TODO: use Azure SDK to dynamically check types/versions available in locations.
func Available(location, resourceType, resourceVersion string) (bool, error) {
	if !contains(allAzureLocations, location) {
		return false, errors.Formatf("invalid location specified: %s\n", location)
	}

}
