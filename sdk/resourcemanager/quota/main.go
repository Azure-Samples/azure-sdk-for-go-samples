// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	armPolicy "github.com/Azure/azure-sdk-for-go/sdk/azcore/arm/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/quota/armquota"
)

const (
	location          = "westus2"
	managementGroupId = "testMgIdRoot"
	resourceName      = "cores"
	provider          = "Microsoft.Compute"
	subscriptionId    = "65a85478-2333-4bbd-981b-1a818c944faf"
	tenantId          = "7be17286-b3ad-482d-ab84-4c772eebd529"
)

const (
	moduleName    = "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/quota/armquota"
	moduleVersion = "v1.1.0-beta.1"
)

func main() {
	//Group Quota Functions
	createGroupQuota("groupquota1", managementGroupId)
	getGroupQuota("groupquota1")

	//Subscription Functions
	addSubscription("groupquota1")
	deleteSubscription("groupquota1")

	//GroupQuotaLimit
	createGroupQuotaLimitRequest("groupquota1", provider, resourceName, location, 64)
	getGroupQuotaLimit("groupquota1", provider, resourceName, location)

	//SubscriptionQuotaAllocation
	createSubscriptionAllocationRequest("groupquota1", provider, resourceName, location, 10)
	getSubscriptionQuotaAllocation("groupquota1", resourceName, location)

	//Cleanup
	deleteSubscription("groupquota1")
	deleteGroupQuota("groupquota1")
}

func createClientFactory() *armquota.ClientFactory {
	//TODO: wrap in if statement for only corp tenant and only CLI
	credOptions := azidentity.DefaultAzureCredentialOptions{
		AdditionallyAllowedTenants: []string{tenantId},
	}

	cred, err := azidentity.NewDefaultAzureCredential(&credOptions)

	if err != nil {
		log.Fatalf("failed to obtain a credential: %v\n", err)
	}

	clientFactory, err := armquota.NewClientFactory(subscriptionId, cred, getArmClientOptions())

	if err != nil {
		log.Fatalf("failed to create client factory: %v", err)
	}
	return clientFactory
}

func getArmClientOptions() *armPolicy.ClientOptions {
	return &armPolicy.ClientOptions{
		ClientOptions: azcore.ClientOptions{
			Cloud: cloud.Configuration{
				ActiveDirectoryAuthorityHost: "https://login.microsoft.com/",
				Services: map[cloud.ServiceName]cloud.ServiceConfiguration{
					cloud.ResourceManager: {
						Audience: "https://management.azure.com/",
						Endpoint: "https://management.azure.com/",
					},
				},
			},
		},
		AuxiliaryTenants:      []string{},
		DisableRPRegistration: false,
	}
}
