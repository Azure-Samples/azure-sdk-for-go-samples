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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	registryName      = "sample2registry"
	webhookName       = "sample2webhook"
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

	registry, err := createRegistry(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("registry:", *registry.ID)

	webhook, err := createWebhook(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("webhook:", *webhook.ID)

	webhook, err = getWebhook(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get webhook:", *webhook.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createRegistry(ctx context.Context, cred azcore.TokenCredential) (*armcontainerregistry.Registry, error) {
	registriesClient, err := armcontainerregistry.NewRegistriesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := registriesClient.BeginCreate(
		ctx,
		resourceGroupName,
		registryName,
		armcontainerregistry.Registry{
			Location: to.Ptr(location),
			Tags: map[string]*string{
				"key": to.Ptr("value"),
			},
			SKU: &armcontainerregistry.SKU{
				Name: to.Ptr(armcontainerregistry.SKUNameStandard),
			},
			Properties: &armcontainerregistry.RegistryProperties{
				AdminUserEnabled: to.Ptr(true),
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
	return &resp.Registry, nil
}

func createWebhook(ctx context.Context, cred azcore.TokenCredential) (*armcontainerregistry.Webhook, error) {
	webhooksClient, err := armcontainerregistry.NewWebhooksClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := webhooksClient.BeginCreate(
		ctx,
		resourceGroupName,
		registryName,
		webhookName,
		armcontainerregistry.WebhookCreateParameters{
			Location: to.Ptr(location),
			Properties: &armcontainerregistry.WebhookPropertiesCreateParameters{
				Actions: []*armcontainerregistry.WebhookAction{
					to.Ptr(armcontainerregistry.WebhookActionPush),
				},
				ServiceURI: to.Ptr("https://www.microsoft.com"),
				Status:     to.Ptr(armcontainerregistry.WebhookStatusEnabled),
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
	return &resp.Webhook, nil
}

func getWebhook(ctx context.Context, cred azcore.TokenCredential) (*armcontainerregistry.Webhook, error) {
	webhooksClient, err := armcontainerregistry.NewWebhooksClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := webhooksClient.Get(ctx, resourceGroupName, registryName, webhookName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Webhook, nil
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
