// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	clientFactory       *armapimanagement.ClientFactory
	resourceGroupClient *armresources.ResourceGroupsClient

	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	serviceName       = "sample-api-service"
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

	clientFactory, err = armapimanagement.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}

	resourcesClientFactory, err := armresources.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	resourceGroupClient = resourcesClientFactory.NewResourceGroupsClient()

	resourceGroup, err := createResourceGroup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	// if happen soft-delete please use delete_service sample to delete
	apiManagementService, err := createApiManagementService(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("api management service:", *apiManagementService.ID)

	apiManagementService, err = getApiManagementService(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get api management service:", *apiManagementService.ID)

	ssoToken, err := getSsoToken(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("ssoToken:", *ssoToken.RedirectURI)

	domainOwnershipIdentifier, err := getDomainOwnershipIdentifier(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("domain owner ship Identifier:", *domainOwnershipIdentifier.DomainOwnershipIdentifier)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createApiManagementService(ctx context.Context) (*armapimanagement.ServiceResource, error) {

	pollerResp, err := clientFactory.NewServiceClient().BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serviceName,
		armapimanagement.ServiceResource{
			Location: to.Ptr(location),
			Properties: &armapimanagement.ServiceProperties{
				PublisherName:  to.Ptr("sample"),
				PublisherEmail: to.Ptr("xxx@wircesoft.com"),
			},
			SKU: &armapimanagement.ServiceSKUProperties{
				Name:     to.Ptr(armapimanagement.SKUTypeStandard),
				Capacity: to.Ptr[int32](2),
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

// The resource type 'getDomainOwnershipIdentifier' could not be found in the namespace 'Microsoft.ApiManagement' for api version '2021-04-01-preview'. The supported api-versions are '2020-12-01,2021-01-01-preview'."}
func getDomainOwnershipIdentifier(ctx context.Context) (*armapimanagement.ServiceGetDomainOwnershipIdentifierResult, error) {

	resp, err := clientFactory.NewServiceClient().GetDomainOwnershipIdentifier(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ServiceGetDomainOwnershipIdentifierResult, nil
}

func getSsoToken(ctx context.Context) (*armapimanagement.ServiceGetSsoTokenResult, error) {

	resp, err := clientFactory.NewServiceClient().GetSsoToken(ctx, resourceGroupName, serviceName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ServiceGetSsoTokenResult, nil
}

func getApiManagementService(ctx context.Context) (*armapimanagement.ServiceResource, error) {

	resp, err := clientFactory.NewServiceClient().Get(ctx, resourceGroupName, serviceName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ServiceResource, nil
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
