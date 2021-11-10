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
	apiID             = "sample-api"
	schemaID          = "sample-api-schema"
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

	api, err := createApi(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("api:", *api.ID)

	apiSchema, err := createApiSchema(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("api schema:", *apiSchema.ID)

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

func createApi(ctx context.Context, conn *arm.Connection) (*armapimanagement.APIContract, error) {
	APIClient := armapimanagement.NewAPIClient(conn, subscriptionID)

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

func createApiSchema(ctx context.Context, conn *arm.Connection) (*armapimanagement.SchemaContract, error) {
	apiSchemaClient := armapimanagement.NewAPISchemaClient(conn, subscriptionID)

	pollerResp, err := apiSchemaClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serviceName,
		apiID,
		schemaID,
		armapimanagement.SchemaContract{
			Properties: &armapimanagement.SchemaContractProperties{
				ContentType: to.StringPtr("application/json"),
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
	return &resp.SchemaContract, nil
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
