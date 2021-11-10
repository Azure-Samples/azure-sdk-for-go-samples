package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
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
	versionSetID      = "sample-api-version"
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

	conn := arm.NewDefaultConnection(cred, &arm.ConnectionOptions{
		Logging: policy.LogOptions{
			IncludeBody: true,
		},
	})
	ctx := context.Background()

	resourceGroup, err := createResourceGroup(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	apiManagementService, err := createApiManagementService(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("api management service:", *apiManagementService.ID)

	apiVersionSet, err := createApiVersionSet(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("api version set:", *apiVersionSet.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createApiManagementService(ctx context.Context, conn *arm.Connection) (*armapimanagement.APIManagementServiceResource, error) {
	apiManagementServiceClient := armapimanagement.NewAPIManagementServiceClient(conn, subscriptionID)

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

func createApiVersionSet(ctx context.Context, conn *arm.Connection) (*armapimanagement.APIVersionSetContract, error) {
	apiVersionSetClient := armapimanagement.NewAPIVersionSetClient(conn, subscriptionID)

	resp, err := apiVersionSetClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		serviceName,
		versionSetID,
		armapimanagement.APIVersionSetContract{
			Properties: &armapimanagement.APIVersionSetContractProperties{
				DisplayName:      to.StringPtr("sample api version set description"),
				VersioningScheme: armapimanagement.VersioningSchemeQuery.ToPtr(),
				APIVersionSetEntityBase: armapimanagement.APIVersionSetEntityBase{
					VersionQueryName: to.StringPtr("sample-version-query"),
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.APIVersionSetContract, nil
}

func createResourceGroup(ctx context.Context, conn *arm.Connection) (*armresources.ResourceGroup, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn, subscriptionID)

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

func cleanup(ctx context.Context, conn *arm.Connection) (*http.Response, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn, subscriptionID)

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
