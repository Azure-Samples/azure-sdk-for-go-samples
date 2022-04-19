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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	serviceName       = "sample-api-service"
	apiID             = "sample-api"
	tagID             = "sample-tag"
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

	apiManagementService, err := createApiManagementService(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("api management service:", *apiManagementService.ID)

	api, err := createApi(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("api:", *api.ID)

	tag, err := createTag(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("tag:", *tag.ID)

	apiTagDescription, err := createApiTagDescription(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("api tag description:", *apiTagDescription.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createApiManagementService(ctx context.Context, cred azcore.TokenCredential) (*armapimanagement.ServiceResource, error) {
	apiManagementServiceClient, err := armapimanagement.NewServiceClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := apiManagementServiceClient.BeginCreateOrUpdate(
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
	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.ServiceResource, nil
}

func createApi(ctx context.Context, cred azcore.TokenCredential) (*armapimanagement.APIContract, error) {
	APIClient, err := armapimanagement.NewAPIClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := APIClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serviceName,
		apiID,
		armapimanagement.APICreateOrUpdateParameter{
			Properties: &armapimanagement.APICreateOrUpdateProperties{
				Path:        to.Ptr("test"),
				DisplayName: to.Ptr("sample-sample"),
				Protocols: []*armapimanagement.Protocol{
					to.Ptr(armapimanagement.ProtocolHTTP),
					to.Ptr(armapimanagement.ProtocolHTTPS),
				},
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
	return &resp.APIContract, nil
}

func createTag(ctx context.Context, cred azcore.TokenCredential) (*armapimanagement.TagClientCreateOrUpdateResponse, error) {
	tagClient, err := armapimanagement.NewTagClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := tagClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		serviceName,
		tagID,
		armapimanagement.TagCreateUpdateParameters{
			Properties: &armapimanagement.TagContractProperties{
				DisplayName: to.Ptr("sample-tag"),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func createApiTagDescription(ctx context.Context, cred azcore.TokenCredential) (*armapimanagement.TagDescriptionContract, error) {
	apiTagDescriptionClient, err := armapimanagement.NewAPITagDescriptionClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := apiTagDescriptionClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		serviceName,
		apiID,
		tagID,
		armapimanagement.TagDescriptionCreateParameters{
			Properties: &armapimanagement.TagDescriptionBaseProperties{
				Description: to.Ptr("sample tag description"),
				//ExternalDocsDescription: to.Ptr(""),
				//ExternalDocsURL: to.Ptr(""),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.TagDescriptionContract, nil
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
