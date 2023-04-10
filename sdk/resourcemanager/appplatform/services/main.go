// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appplatform/armappplatform"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	serviceName       = "sample-spring-cloud"
)

var (
	resourcesClientFactory   *armresources.ClientFactory
	appplatformClientFactory *armappplatform.ClientFactory
)

var (
	resourceGroupClient *armresources.ResourceGroupsClient
	servicesClient      *armappplatform.ServicesClient
)

func main() {
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set.")
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	resourcesClientFactory, err = armresources.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	resourceGroupClient = resourcesClientFactory.NewResourceGroupsClient()

	appplatformClientFactory, err = armappplatform.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	servicesClient = appplatformClientFactory.NewServicesClient()

	resourceGroup, err := createResourceGroup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	service, err := createSpringCloudService(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("app platform service:", *service.ID)

	service, err = getSpringCloudService(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get app platform service:", *service.ID)

	testKey, err := regenerateTestKey(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("app platform test key:", *testKey.PrimaryKey)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createSpringCloudService(ctx context.Context) (*armappplatform.ServiceResource, error) {

	pollerResp, err := servicesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serviceName,
		armappplatform.ServiceResource{
			Location: to.Ptr(location),
			SKU: &armappplatform.SKU{
				Name: to.Ptr("S0"),
				Tier: to.Ptr("Standard"),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &resp.ServiceResource, nil
}

func getSpringCloudService(ctx context.Context) (*armappplatform.ServiceResource, error) {

	resp, err := servicesClient.Get(ctx, resourceGroupName, serviceName, nil)
	if err != nil {
		return nil, err
	}

	return &resp.ServiceResource, nil
}

func regenerateTestKey(ctx context.Context) (*armappplatform.TestKeys, error) {

	resp, err := servicesClient.RegenerateTestKey(
		ctx,
		resourceGroupName,
		serviceName,
		armappplatform.RegenerateTestKeyRequestPayload{
			KeyType: to.Ptr(armappplatform.TestKeyTypePrimary),
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &resp.TestKeys, nil
}

func createResourceGroup(ctx context.Context) (*armresources.ResourceGroup, error) {

	resourceGroupResp, err := resourceGroupClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		armresources.ResourceGroup{
			Location: to.Ptr(location),
		},
		nil)
	if err != nil {
		return nil, err
	}

	return &resourceGroupResp.ResourceGroup, nil
}

func cleanup(ctx context.Context) error {

	pollerResp, err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return err
	}

	_, err = pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}
