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
	operationID       = "sample-api-operation"
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

	apiOperation, err := createApiOperation(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("api operation:", *apiOperation.ID)

	apiOperationPolicy, err := createApiOperationPolicy(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("api operation policy:", *apiOperationPolicy.ID)

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

func createApiOperation(ctx context.Context, cred azcore.TokenCredential) (*armapimanagement.OperationContract, error) {
	apiOperationClient, err := armapimanagement.NewAPIOperationClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := apiOperationClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		serviceName,
		apiID,
		operationID,
		armapimanagement.OperationContract{
			Properties: &armapimanagement.OperationContractProperties{
				DisplayName: to.Ptr("sample operation"),
				Method:      to.Ptr("GET"),
				URLTemplate: to.Ptr("/operation/customers/{uid}"),
				TemplateParameters: []*armapimanagement.ParameterContract{
					{
						Name:        to.Ptr("uid"),
						Type:        to.Ptr("string"),
						Description: to.Ptr("user id"),
					},
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.OperationContract, nil
}

var value = `<?xml version="1.0" encoding="utf-8"?>
<policies>
    <inbound>
        <base />
    </inbound>
    <backend>
        <base />
    </backend>
    <outbound>
        <base />
    </outbound>
    <on-error>
        <base />
    </on-error>
</policies>
`

func createApiOperationPolicy(ctx context.Context, cred azcore.TokenCredential) (*armapimanagement.PolicyContract, error) {
	apiOperationPolicyClient, err := armapimanagement.NewAPIOperationPolicyClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := apiOperationPolicyClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		serviceName,
		apiID,
		operationID,
		armapimanagement.PolicyIDNamePolicy,
		armapimanagement.PolicyContract{
			Properties: &armapimanagement.PolicyContractProperties{
				Format: to.Ptr(armapimanagement.PolicyContentFormatXML),
				Value:  to.Ptr(value),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.PolicyContract, nil
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
