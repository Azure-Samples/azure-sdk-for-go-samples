// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventgrid/armeventgrid/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"log"
	"os"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	domainName        = "sample-domain"
	domainTopicName   = "sample-domain-topic"
)

var (
	resourcesClientFactory *armresources.ClientFactory
	eventgridClientFactory *armeventgrid.ClientFactory
)

var (
	resourceGroupClient *armresources.ResourceGroupsClient
	domainsClient       *armeventgrid.DomainsClient
	domainTopicsClient  *armeventgrid.DomainTopicsClient
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

	eventgridClientFactory, err = armeventgrid.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	domainsClient = eventgridClientFactory.NewDomainsClient()
	domainTopicsClient = eventgridClientFactory.NewDomainTopicsClient()

	resourceGroup, err := createResourceGroup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	domain, err := createDomain(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("domain:", *domain.ID)

	domainTopic, err := createDomainTopic(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("domain topic:", *domainTopic.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createDomain(ctx context.Context) (*armeventgrid.Domain, error) {

	pollerResp, err := domainsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		domainName,
		armeventgrid.Domain{
			Location: to.Ptr(location),
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
	return &resp.Domain, nil
}

func createDomainTopic(ctx context.Context) (*armeventgrid.DomainTopic, error) {

	pollerResp, err := domainTopicsClient.BeginCreateOrUpdate(ctx, resourceGroupName, domainName, domainTopicName, nil)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.DomainTopic, nil
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
