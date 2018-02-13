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

	"github.com/subosito/gotenv"
)

var (
	resourceGroupNamePrefix  string
	resourceGroupName        string
	location                 string
	subscriptionID           string
	servicePrincipalObjectID string
	keepResources            bool
	deviceFlow               bool

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
	err = LoadEnvVars()
	if err != nil {
		return err
	}

	resourceGroupNamePrefix = os.Getenv("AZ_RESOURCE_GROUP_PREFIX")
	servicePrincipalObjectID = os.Getenv("AZ_SP_OBJECT_ID")
	location = os.Getenv("AZ_LOCATION")
	if os.Getenv("AZ_SAMPLES_KEEP_RESOURCES") == "1" {
		keepResources = true
	}

	// flags override envvars
	flag.StringVar(&resourceGroupNamePrefix, "groupPrefix", GroupPrefix(), "Specify prefix name of resource group for sample resources.")
	flag.StringVar(&location, "location", location, "Provide the Azure location where the resources will be be created")
	flag.BoolVar(&keepResources, "keepResources", keepResources, "Keep resources created by samples.")
	flag.Parse()

	// defaults
	if !(len(resourceGroupNamePrefix) > 0) {
		resourceGroupNamePrefix = GroupPrefix()
	}

	if !(len(location) > 0) {
		location = "westus2" // lots of space, most new features
	}
	return nil
}

// ParseSubscriptionID gets the subscription id from either an env var, .env file or flag
// The caller should do flag.Parse()
func ParseSubscriptionID() error {
	err := LoadEnvVars()
	if err != nil {
		return err
	}

	subscriptionID = os.Getenv("AZ_SUBSCRIPTION_ID")
	flag.StringVar(&subscriptionID, "subscription", subscriptionID, "Subscription to use for deployment.")

	if !(len(subscriptionID) > 0) {
		return errors.New("subscription ID must be specified in env var, .env file or flag")
	}
	return nil
}

// ParseDeviceFlow parses the auth grant type to be used
// The caller should do flag.Parse()
func ParseDeviceFlow() error {
	err := LoadEnvVars()
	if err != nil {
		return err
	}

	if os.Getenv("AZ_AUTH_DEVICEFLOW") != "" {
		deviceFlow = true
	}
	flag.BoolVar(&deviceFlow, "deviceFlow", deviceFlow, "Use device flow for authentication. This flag should be used with -v flag. Default authentication is service principal.")
	return nil
}

// getters

// KeepResources indicates whether resources created by samples should be retained.
func KeepResources() bool {
	return keepResources
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
	return deviceFlow
}

// end getters

func SetPrefix(prefix string) {
	resourceGroupNamePrefix = prefix
}

func SetResourceGroupName(suffix string) {
	resourceGroupName = GroupPrefix() + "-" + suffix + "-" + GetRandomLetterSequence(5)
}

// OverrideLocation ovverrides the specified location where to create Azure resources.
// This can be used when the selection location does not have the desired resource provider available yet
func OverrideLocation(available []string) {
	// If location is not listed on all locations, don't override it. It might be a canary location
	if contains(allLocations, location) && !contains(available, location) && len(available) > 0 {
		log.Printf("Using location %s on this sample, because this service is not yet available on specified location %s\n", available[0], location)
		location = available[0]
	}
}

// LoadEnvVars loads environment variables.
func LoadEnvVars() error {
	err := gotenv.Load() // to allow use of .env file
	if err != nil && !strings.HasPrefix(err.Error(), "open .env:") {
		return err
	}
	return nil
}
