package config

import (
	"log"
	"strings"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
)

var allLocations = []string{
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

var locationOverrideTemplate = "overriding default location %s for this package because service is not available there. using location %s instead.\n"

// OverrideCanaryLocation overrides the specified canary location where to create Azure resources.
func OverrideCanaryLocation(usableLocation string) {
	if strings.HasSuffix(location, "euap") {
		log.Printf(locationOverrideTemplate, usableLocation, location)
		location = usableLocation
	}
}

// OverrideLocation overrides the specified location where to create Azure resources.
// This can be used when the selection location does not have the desired resource provider available yet
func OverrideLocation(available []string) {
	// If location is not listed on all locations, don't override it. It might be a canary location
	if util.Contains(allLocations, location) && !util.Contains(available, location) && len(available) > 0 {
		log.Printf(locationOverrideTemplate, available[0], location)
		location = available[0]
	}
}
