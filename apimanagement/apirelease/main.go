package main

import (
	"context"
	"log"
	"net/http"
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
	releaseID         = "sample-api-release"
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

	apiRelease, err := createApiRelease(ctx, cred, *api.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("api release:", *apiRelease.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createApiManagementService(ctx context.Context, cred azcore.TokenCredential) (*armapimanagement.APIManagementServiceResource, error) {
	apiManagementServiceClient := armapimanagement.NewAPIManagementServiceClient(subscriptionID, cred, nil)

	pollerResp, err := apiManagementServiceClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serviceName,
		armapimanagement.APIManagementServiceResource{
			Location: to.StringPtr(location),
			Properties: &armapimanagement.APIManagementServiceProperties{
				PublisherName:  to.StringPtr("sample"),
				PublisherEmail: to.StringPtr("xxx@wircesoft.com"),
			},
			SKU: &armapimanagement.APIManagementServiceSKUProperties{
				Name:     armapimanagement.SKUTypeStandard.ToPtr(),
				Capacity: to.Int32Ptr(2),
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
	return &resp.APIManagementServiceResource, nil
}

func createApi(ctx context.Context, cred azcore.TokenCredential) (*armapimanagement.APIContract, error) {
	APIClient := armapimanagement.NewAPIClient(subscriptionID, cred, nil)

	pollerResp, err := APIClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serviceName,
		apiID,
		armapimanagement.APICreateOrUpdateParameter{
			Properties: &armapimanagement.APICreateOrUpdateProperties{
				APIContractProperties: armapimanagement.APIContractProperties{
					Path:        to.StringPtr("test"),
					DisplayName: to.StringPtr("sample-sample"),
					Protocols: []*armapimanagement.Protocol{
						armapimanagement.ProtocolHTTP.ToPtr(),
						armapimanagement.ProtocolHTTPS.ToPtr(),
					},
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

func createApiRelease(ctx context.Context, cred azcore.TokenCredential, apiId string) (*armapimanagement.APIReleaseContract, error) {
	apiReleaseClient := armapimanagement.NewAPIReleaseClient(subscriptionID, cred, nil)

	resp, err := apiReleaseClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		serviceName,
		apiID,
		releaseID,
		armapimanagement.APIReleaseContract{
			Properties: &armapimanagement.APIReleaseContractProperties{
				APIID: to.StringPtr(apiId),
				Notes: to.StringPtr("sample api release"),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.APIReleaseContract, nil
}

func createResourceGroup(ctx context.Context, cred azcore.TokenCredential) (*armresources.ResourceGroup, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)

	resourceGroupResp, err := resourceGroupClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		armresources.ResourceGroup{
			Location: to.StringPtr(location),
		},
		nil)
	if err != nil {
		return nil, err
	}
	return &resourceGroupResp.ResourceGroup, nil
}

func cleanup(ctx context.Context, cred azcore.TokenCredential) (*http.Response, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)

	pollerResp, err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp.RawResponse, nil
}
