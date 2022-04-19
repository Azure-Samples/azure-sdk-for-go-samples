// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicebus/armservicebus"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	namespaceName     = "sample-sb-namespace"
	topicName         = "sample-sb-topic"
	subscriptionName  = "sample-subscription"
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

	resourceGroup, err := createResourceGroup(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	namespace, err := createNamespace(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus namespace:", *namespace.ID)

	topic, err := createTopic(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus topic:", *topic.ID)

	subscription, err := createSubscription(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus subscription:", *subscription.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createNamespace(ctx context.Context, cred azcore.TokenCredential) (*armservicebus.SBNamespace, error) {
	namespacesClient, err := armservicebus.NewNamespacesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

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

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.SBNamespace, nil
}

func createTopic(ctx context.Context, cred azcore.TokenCredential) (*armservicebus.SBTopic, error) {
	topicsClient, err := armservicebus.NewTopicsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

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

func createSubscription(ctx context.Context, cred azcore.TokenCredential) (*armservicebus.SBSubscription, error) {
	subscriptionsClient, err := armservicebus.NewSubscriptionsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

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

func createResourceGroup(ctx context.Context, cred azcore.TokenCredential) (*armresources.ResourceGroup, error) {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

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

func cleanup(ctx context.Context, cred azcore.TokenCredential) error {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	pollerResp, err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return err
	}

	_, err = pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return err
	}
	return nil
}
