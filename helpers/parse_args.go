package helpers

import (
	"errors"
	"flag"
	"os"

	"github.com/subosito/gotenv"
)

var (
	resourceGroupName string
	location          string
	subscriptionID    string
	keepResources     bool
)

// ParseArgs picks up shared env vars and flags and finishes parsing flags
// Other packages should declare their flags then call helpers.ParseArgs()
func ParseArgs() error {
	// flags are prioritized over env vars,
	// so read from env vars first, then check flags
	err := gotenv.Load() // to allow use of .env file
	if err != nil && err.Error() != "open .env: The system cannot find the file specified." {
		return err
	}

	subscriptionID = os.Getenv("AZ_SUBSCRIPTION_ID")
	resourceGroupName = os.Getenv("AZ_RESOURCE_GROUP_NAME")
	location = os.Getenv("AZ_LOCATION")
	if os.Getenv("AZ_SAMPLES_KEEP_RESOURCES") == "1" {
		keepResources = true
	} else {
		keepResources = false
	}

	// flags override envvars
	flag.StringVar(&resourceGroupName, "groupName", resourceGroupName, "Specify name of resource group for sample resources.")
	flag.StringVar(&location, "location", location, "Provide the Azure location where the resources will be be created")
	flag.StringVar(&subscriptionID, "subscription", subscriptionID, "Subscription to use for deployment.")
	flag.BoolVar(&keepResources, "keepResources", keepResources, "Keep resources created by samples.")
	flag.Parse()

	// defaults
	if !(len(subscriptionID) > 0) {
		return errors.New("subscription ID must be specified in env var, .env file or flag")
	}

	if !(len(resourceGroupName) > 0) {
		resourceGroupName = "group-azure-samples-go" + GetRandomLetterSequence(10)
	}

	if !(len(location) > 0) {
		location = "westus2" // lots of space, most new features
	}
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

// ResourceGroupName returns the name of the resource group to use.
func ResourceGroupName() string {
	return resourceGroupName
}

// Location specifies the Azure region to use.
func Location() string {
	return location
}

// end getters
