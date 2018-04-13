// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package helpers

import (
	"errors"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/Azure/go-autorest/autorest/to"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/subosito/gotenv"
)

var (
	targetEnv                = azure.PublicCloud.Name
	resourceGroupNamePrefix  string
	resourceGroupName        string
	location                 string
	subscriptionID           string
	servicePrincipalObjectID string
	keepResourcesPtr         *bool
	deviceFlow               *bool

	allLocations = []string{
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

	locationOverrideTemplate = "Using location %s on this sample, because this service is not yet available on specified location %s\n"
)

// ParseArgs picks up shared env vars and flags and finishes parsing flags
// Other packages should declare their flags then call helpers.ParseArgs()
func ParseArgs() error {
	err := ParseSubscriptionID()
	if err != nil {
		return err
	}

	err = ParseDeviceFlow()
	if err != nil {
		return err
	}

	// flags are prioritized over env vars,
	// so read from env vars first, then check flags
	err = ReadEnvFile()
	if err != nil {
		return err
	}

	servicePrincipalObjectID = os.Getenv("AZURE_SP_OBJECT_ID")

	// flags override envvars
	if resourceGroupNamePrefix == "" {
		resourceGroupNamePrefix = os.Getenv("AZURE_RESOURCE_GROUP_PREFIX")
		flag.StringVar(&resourceGroupNamePrefix, "groupPrefix", GroupPrefix(), "Specify prefix name of resource group for sample resources.")
	}

	if location == "" {
		location = os.Getenv("AZURE_LOCATION")
		if location == "" {
			location = "westus2" // lots of space, most new features
		}
		flag.StringVar(&location, "location", location, "Provide the Azure location where the resources will be be created.")
	}

	if keepResourcesPtr == nil {
		keepResources := false
		if os.Getenv("AZURE_SAMPLES_KEEP_RESOURCES") == "1" {
			keepResources = true
		}
		flag.BoolVar(&keepResources, "keepResources", keepResources, "Keep resources created by samples.")
		keepResourcesPtr = &keepResources
	}

	if targetEnv == "" {
		targetEnv = os.Getenv("AZURE_ENVIRONMENT")
		if targetEnv == "" {
			targetEnv = azure.PublicCloud.Name
		}
		flag.StringVar(&targetEnv, "environment", targetEnv, "Azure environment.")
	}

	flag.Parse()
	return nil
}

// ParseSubscriptionID gets the subscription id from either an env var, .env file or flag
// The caller should do flag.Parse()
func ParseSubscriptionID() error {
	if subscriptionID != "" {
		return nil
	}
	err := ReadEnvFile()
	if err != nil {
		return err
	}

	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	flag.StringVar(&subscriptionID, "subscription", subscriptionID, "Subscription to use for deployment.")

	if !(len(subscriptionID) > 0) {
		return errors.New("subscription ID must be specified in env var, .env file or flag")
	}
	return nil
}

// ParseDeviceFlow parses the auth grant type to be used
// The caller should do flag.Parse()
func ParseDeviceFlow() error {
	if deviceFlow != nil {
		return nil
	}
	err := ReadEnvFile()
	if err != nil {
		return err
	}

	deviceFlow = to.BoolPtr(false)

	if os.Getenv("AZURE_AUTH_DEVICEFLOW") != "" {
		deviceFlow = to.BoolPtr(true)
	}
	flag.BoolVar(deviceFlow, "deviceFlow", *deviceFlow, "Use device flow for authentication. This flag should be used with -v flag. Default authentication is service principal.")
	return nil
}

// getters

// KeepResources indicates whether resources created by samples should be retained.
func KeepResources() bool {
	if keepResourcesPtr == nil {
		return false
	}
	return *keepResourcesPtr
}

// SubscriptionID returns the ID of the subscription to use.
func SubscriptionID() string {
	return subscriptionID
}

// ServicePrincipalObjectID returns the object ID of the service principal in use.
func ServicePrincipalObjectID() string {
	return servicePrincipalObjectID
}

// ResourceGroupName returns the name of the resource group to use.
func ResourceGroupName() string {
	return resourceGroupName
}

// Location specifies the Azure region to use.
func Location() string {
	return location
}

// GroupPrefix specifies the prefix sample resource groups should have
func GroupPrefix() string {
	if resourceGroupNamePrefix == "" {
		return "azure-samples-go"
	}
	return resourceGroupNamePrefix
}

// DeviceFlow returns if device flow has been set as auth grant type
func DeviceFlow() bool {
	if deviceFlow == nil {
		return false
	}
	return *deviceFlow
}

// Environment gets the Azure environment
func Environment() azure.Environment {
	env, err := azure.EnvironmentFromName(targetEnv)
	if err != nil {
		log.Fatalf("failed to get environment from name (defaulting to Public Cloud): %s\n", err)
		return azure.PublicCloud
	}
	return env
}

// ArmEndpoint specifies resource manager URI
func ArmEndpoint() string {
	return Environment().ResourceManagerEndpoint
}

// end getters

// SetPrefix sets a prefix for resource group names
func SetPrefix(prefix string) {
	resourceGroupNamePrefix = prefix
}

// SetResourceGroupName sets a name for the resource group. It takes into account the
// resource group prefix, and adds some random letters to ensure uniqueness
func SetResourceGroupName(suffix string) {
	resourceGroupName = GroupPrefix() + "-" + suffix + "-" + GetRandomLetterSequence(5)
}

// OverrideCanaryLocation ovverrides the specified canary location where to create Azure resources.
func OverrideCanaryLocation(usableLocation string) {
	if strings.HasSuffix(location, "euap") {
		log.Printf(locationOverrideTemplate, usableLocation, location)
		location = usableLocation
	}
}

// OverrideLocation ovverrides the specified location where to create Azure resources.
// This can be used when the selection location does not have the desired resource provider available yet
func OverrideLocation(available []string) {
	// If location is not listed on all locations, don't override it. It might be a canary location
	if contains(allLocations, location) && !contains(available, location) && len(available) > 0 {
		log.Printf(locationOverrideTemplate, available[0], location)
		location = available[0]
	}
}

// ReadEnvFile reads the .env file and loads its environment variables.
func ReadEnvFile() error {
	err := gotenv.Load() // to allow use of .env file
	if err != nil && !strings.HasPrefix(err.Error(), "open .env:") {
		return err
	}
	return nil
}
