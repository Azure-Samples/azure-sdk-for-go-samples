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
	topicName         = "sample-sb-topic"
	subscriptionName  = "sample-subscription"
)

var (
	resourcesClientFactory  *armresources.ClientFactory
	servicebusClientFactory *armservicebus.ClientFactory
)

var (
	resourceGroupClient *armresources.ResourceGroupsClient
	namespacesClient    *armservicebus.NamespacesClient
	topicsClient        *armservicebus.TopicsClient
	subscriptionsClient *armservicebus.SubscriptionsClient
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
	topicsClient = servicebusClientFactory.NewTopicsClient()
	subscriptionsClient = servicebusClientFactory.NewSubscriptionsClient()

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

	topic, err := createTopic(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus topic:", *topic.ID)

	subscription, err := createSubscription(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus subscription:", *subscription.ID)

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

func createTopic(ctx context.Context) (*armservicebus.SBTopic, error) {

	resp, err := topicsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		namespaceName,
		topicName,
		armservicebus.SBTopic{
			Properties: &armservicebus.SBTopicProperties{},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.SBTopic, nil
}

func createSubscription(ctx context.Context) (*armservicebus.SBSubscription, error) {

	resp, err := subscriptionsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		namespaceName,
		topicName,
		subscriptionName,
		armservicebus.SBSubscription{},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.SBSubscription, nil
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
