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
)

var (
	resourcesClientFactory *armresources.ClientFactory
	eventgridClientFactory *armeventgrid.ClientFactory
)

var (
	resourceGroupClient *armresources.ResourceGroupsClient
	domainsClient       *armeventgrid.DomainsClient
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

	domain, err = getDomain(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get domain:", *domain.ID)

	keys, err := regenerateKey(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("regenerate key:", *keys.Key1, *keys.Key2)

	domains, err := listDomain(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for _, d := range domains {
		log.Println(*d.Name, *d.ID)
	}

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

func getDomain(ctx context.Context) (*armeventgrid.Domain, error) {

	resp, err := domainsClient.Get(ctx, resourceGroupName, domainName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Domain, nil
}

func regenerateKey(ctx context.Context) (*armeventgrid.DomainSharedAccessKeys, error) {

	resp, err := domainsClient.RegenerateKey(
		ctx,
		resourceGroupName,
		domainName,
		armeventgrid.DomainRegenerateKeyRequest{
			KeyName: to.Ptr("key1"),
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.DomainSharedAccessKeys, nil
}

func listDomain(ctx context.Context) ([]*armeventgrid.Domain, error) {

	pager := domainsClient.NewListByResourceGroupPager(resourceGroupName, nil)

	domains := make([]*armeventgrid.Domain, 0)
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		domains = append(domains, resp.Value...)
	}
	return domains, nil
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
