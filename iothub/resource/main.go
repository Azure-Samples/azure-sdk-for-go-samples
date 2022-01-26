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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/iothub/armiothub"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	location          = "eastus"
	resourceGroupName = "sample-resource-group"
	iotHubName        = "sample-iothub"
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

	iothubResource, err := createIotHubResource(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("iothub resource:", *iothubResource.ID)

	iothubResource, err = getIotHubResource(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get iothub resource:", *iothubResource.ID)

	iothubStats, err := getIotHubStats(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("get iothub resource stats:%v\n", *iothubStats)

	endpointHealths := getIotHubEndpointHealth(ctx, cred)
	log.Println("get iothub resource endpoint health:", endpointHealths)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createIotHubResource(ctx context.Context, cred azcore.TokenCredential) (*armiothub.Description, error) {
	iotHubResourceClient := armiothub.NewResourceClient(subscriptionID, cred, nil)

	pollerResp, err := iotHubResourceClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		iotHubName,
		armiothub.Description{
			Location: to.StringPtr(location),
			SKU: &armiothub.SKUInfo{
				Name:     armiothub.IotHubSKUS1.ToPtr(),
				Capacity: to.Int64Ptr(1),
			},
			Properties: &armiothub.Properties{
				EnableFileUploadNotifications: to.BoolPtr(false),
				MinTLSVersion:                 to.StringPtr("1.2"),
				EventHubEndpoints: map[string]*armiothub.EventHubProperties{
					"events": {
						RetentionTimeInDays: to.Int64Ptr(1),
						PartitionCount:      to.Int32Ptr(4),
					},
				},
				StorageEndpoints: map[string]*armiothub.StorageEndpointProperties{
					"$default": {
						SasTTLAsIso8601: to.StringPtr("PT1H"),
					},
				},
				MessagingEndpoints: map[string]*armiothub.MessagingEndpointProperties{
					"fileNotifications": {
						LockDurationAsIso8601: to.StringPtr("PT5S"),
						TTLAsIso8601:          to.StringPtr("PT1H"),
						MaxDeliveryCount:      to.Int32Ptr(10),
					},
				},
				CloudToDevice: &armiothub.CloudToDeviceProperties{
					MaxDeliveryCount:    to.Int32Ptr(10),
					DefaultTTLAsIso8601: to.StringPtr("PT1H"),
					Feedback: &armiothub.FeedbackProperties{
						LockDurationAsIso8601: to.StringPtr("PT5S"),
						TTLAsIso8601:          to.StringPtr("PT1H"),
						MaxDeliveryCount:      to.Int32Ptr(10),
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
	return &resp.Description, nil
}

func getIotHubResource(ctx context.Context, cred azcore.TokenCredential) (*armiothub.Description, error) {
	iotHubResourceClient := armiothub.NewResourceClient(subscriptionID, cred, nil)

	resp, err := iotHubResourceClient.Get(ctx, resourceGroupName, iotHubName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Description, nil
}

func getIotHubEndpointHealth(ctx context.Context, cred azcore.TokenCredential) []*armiothub.EndpointHealthData {
	iotHubResourceClient := armiothub.NewResourceClient(subscriptionID, cred, nil)

	pager := iotHubResourceClient.GetEndpointHealth(resourceGroupName, iotHubName, nil)
	results := make([]*armiothub.EndpointHealthData, 0)
	for pager.NextPage(ctx) {
		resp := pager.PageResponse()
		results = append(results, resp.Value...)
	}
	return results
}

func getIotHubStats(ctx context.Context, cred azcore.TokenCredential) (*armiothub.RegistryStatistics, error) {
	iotHubResourceClient := armiothub.NewResourceClient(subscriptionID, cred, nil)

	resp, err := iotHubResourceClient.GetStats(ctx, resourceGroupName, iotHubName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.RegistryStatistics, nil
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
