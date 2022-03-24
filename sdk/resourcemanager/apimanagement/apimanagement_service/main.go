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

	// if happen soft-delete please use delete_service sample to delete
	apiManagementService, err := createApiManagementService(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("api management service:", *apiManagementService.ID)

	apiManagementService, err = getApiManagementService(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get api management service:", *apiManagementService.ID)

	ssoToken, err := getSsoToken(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("ssoToken:", *ssoToken.RedirectURI)

	domainOwnershipIdentifier, err := getDomainOwnershipIdentifier(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("domain owner ship Identifier:", *domainOwnershipIdentifier.DomainOwnershipIdentifier)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createApiManagementService(ctx context.Context, cred azcore.TokenCredential) (*armapimanagement.ServiceResource, error) {
	apiManagementServiceClient := armapimanagement.NewServiceClient(subscriptionID, cred, nil)

	pollerResp, err := apiManagementServiceClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serviceName,
		armapimanagement.ServiceResource{
			Location: to.StringPtr(location),
			Properties: &armapimanagement.ServiceProperties{
				PublisherName:  to.StringPtr("sample"),
				PublisherEmail: to.StringPtr("xxx@wircesoft.com"),
			},
			SKU: &armapimanagement.ServiceSKUProperties{
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
	return &resp.ServiceResource, nil
}

//The resource type 'getDomainOwnershipIdentifier' could not be found in the namespace 'Microsoft.ApiManagement' for api version '2021-04-01-preview'. The supported api-versions are '2020-12-01,2021-01-01-preview'."}
func getDomainOwnershipIdentifier(ctx context.Context, cred azcore.TokenCredential) (*armapimanagement.ServiceGetDomainOwnershipIdentifierResult, error) {
	apiManagementServiceClient := armapimanagement.NewServiceClient(subscriptionID, cred, nil)

	resp, err := apiManagementServiceClient.GetDomainOwnershipIdentifier(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ServiceGetDomainOwnershipIdentifierResult, nil
}

func getSsoToken(ctx context.Context, cred azcore.TokenCredential) (*armapimanagement.ServiceGetSsoTokenResult, error) {
	apiManagementServiceClient := armapimanagement.NewServiceClient(subscriptionID, cred, nil)

	resp, err := apiManagementServiceClient.GetSsoToken(ctx, resourceGroupName, serviceName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ServiceGetSsoTokenResult, nil
}

func getApiManagementService(ctx context.Context, cred azcore.TokenCredential) (*armapimanagement.ServiceResource, error) {
	apiManagementServiceClient := armapimanagement.NewServiceClient(subscriptionID, cred, nil)

	resp, err := apiManagementServiceClient.Get(ctx, resourceGroupName, serviceName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ServiceResource, nil
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
