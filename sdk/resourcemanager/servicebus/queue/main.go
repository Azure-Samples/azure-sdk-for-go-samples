// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicebus/armservicebus"
	"log"
	"os"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	namespaceName     = "sample-sb-namespace"
	queueName         = "sample-sb-queue"
)

var (
	resourcesClientFactory  *armresources.ClientFactory
	servicebusClientFactory *armservicebus.ClientFactory
)

var (
	resourceGroupClient *armresources.ResourceGroupsClient
	namespacesClient    *armservicebus.NamespacesClient
	queuesClient        *armservicebus.QueuesClient
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
	servicebusClientFactory, err = armservicebus.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	namespacesClient = servicebusClientFactory.NewNamespacesClient()
	queuesClient = servicebusClientFactory.NewQueuesClient()

	resourceGroup, err := createResourceGroup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	namespace, err := createNamespace(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus namespace:", *namespace.ID)

	queue, err := createQueue(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus queue:", *queue.ID)

	queueGet, err := getQueue(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get service bus queue:", *queueGet.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createNamespace(ctx context.Context) (*armservicebus.SBNamespace, error) {

	pollerResp, err := namespacesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		namespaceName,
		armservicebus.SBNamespace{
			Location: to.Ptr(location),
			SKU: &armservicebus.SBSKU{
				Name: to.Ptr(armservicebus.SKUNamePremium),
				Tier: to.Ptr(armservicebus.SKUTierPremium),
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
	return &resp.SBNamespace, nil
}

func createQueue(ctx context.Context) (*armservicebus.SBQueue, error) {

	resp, err := queuesClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		namespaceName,
		queueName,
		armservicebus.SBQueue{
			Properties: &armservicebus.SBQueueProperties{
				EnablePartitioning: to.Ptr(true),
			},
		},
		nil,
	)
	if err != nil {
		return nil, nil
	}

	return &resp.SBQueue, nil
}

func getQueue(ctx context.Context) (*armservicebus.SBQueue, error) {

	resp, err := queuesClient.Get(ctx, resourceGroupName, namespaceName, queueName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.SBQueue, nil
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
