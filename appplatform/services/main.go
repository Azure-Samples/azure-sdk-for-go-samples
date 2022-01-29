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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appplatform/armappplatform"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	serviceName       = "sample-spring-cloud"
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

	service, err := createSpringCloudService(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("app platform service:", *service.ID)

	service, err = getSpringCloudService(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get app platform service:", *service.ID)

	testKey, err := regenerateTestKey(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("app platform test key:", *testKey.PrimaryKey)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createSpringCloudService(ctx context.Context, cred azcore.TokenCredential) (*armappplatform.ServiceResource, error) {
	servicesClient := armappplatform.NewServicesClient(subscriptionID, cred, nil)
	pollerResp, err := servicesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serviceName,
		armappplatform.ServiceResource{
			Location: to.StringPtr(location),
			SKU: &armappplatform.SKU{
				Name: to.StringPtr("S0"),
				Tier: to.StringPtr("Standard"),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, 30*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.ServiceResource, nil
}

func getSpringCloudService(ctx context.Context, cred azcore.TokenCredential) (*armappplatform.ServiceResource, error) {
	servicesClient := armappplatform.NewServicesClient(subscriptionID, cred, nil)
	resp, err := servicesClient.Get(ctx, resourceGroupName, serviceName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ServiceResource, nil
}

func regenerateTestKey(ctx context.Context, cred azcore.TokenCredential) (*armappplatform.TestKeys, error) {
	servicesClient := armappplatform.NewServicesClient(subscriptionID, cred, nil)
	resp, err := servicesClient.RegenerateTestKey(ctx, resourceGroupName, serviceName, armappplatform.RegenerateTestKeyRequestPayload{armappplatform.TestKeyTypePrimary.ToPtr()}, nil)
	if err != nil {
		return nil, err
	}
	return &resp.TestKeys, nil
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
