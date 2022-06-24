package main

import (
	"context"
	"log"
	"os"

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
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createIotHubResource(ctx context.Context, cred azcore.TokenCredential) (*armiothub.Description, error) {
	iotHubResourceClient, err := armiothub.NewResourceClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := iotHubResourceClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		iotHubName,
		armiothub.Description{
			Location: to.Ptr(location),
			SKU: &armiothub.SKUInfo{
				Name:     to.Ptr(armiothub.IotHubSKUS1),
				Capacity: to.Ptr[int64](1),
			},
			Properties: &armiothub.Properties{
				EnableFileUploadNotifications: to.Ptr(false),
				MinTLSVersion:                 to.Ptr("1.2"),
				EventHubEndpoints: map[string]*armiothub.EventHubProperties{
					"events": {
						RetentionTimeInDays: to.Ptr[int64](1),
						PartitionCount:      to.Ptr[int32](4),
					},
				},
				StorageEndpoints: map[string]*armiothub.StorageEndpointProperties{
					"$default": {
						SasTTLAsIso8601: to.Ptr("PT1H"),
					},
				},
				MessagingEndpoints: map[string]*armiothub.MessagingEndpointProperties{
					"fileNotifications": {
						LockDurationAsIso8601: to.Ptr("PT5S"),
						TTLAsIso8601:          to.Ptr("PT1H"),
						MaxDeliveryCount:      to.Ptr[int32](10),
					},
				},
				CloudToDevice: &armiothub.CloudToDeviceProperties{
					MaxDeliveryCount:    to.Ptr[int32](10),
					DefaultTTLAsIso8601: to.Ptr("PT1H"),
					Feedback: &armiothub.FeedbackProperties{
						LockDurationAsIso8601: to.Ptr("PT5S"),
						TTLAsIso8601:          to.Ptr("PT1H"),
						MaxDeliveryCount:      to.Ptr[int32](10),
					},
				},
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

	return &resp.Description, nil
}

func getIotHubResource(ctx context.Context, cred azcore.TokenCredential) (*armiothub.Description, error) {
	iotHubResourceClient, err := armiothub.NewResourceClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := iotHubResourceClient.Get(ctx, resourceGroupName, iotHubName, nil)
	if err != nil {
		return nil, err
	}

	return &resp.Description, nil
}

func getIotHubEndpointHealth(ctx context.Context, cred azcore.TokenCredential) []*armiothub.EndpointHealthData {
	iotHubResourceClient, err := armiothub.NewResourceClient(subscriptionID, cred, nil)
	if err != nil {
		return nil
	}

	pager := iotHubResourceClient.NewGetEndpointHealthPager(resourceGroupName, iotHubName, nil)
	results := make([]*armiothub.EndpointHealthData, 0)
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return nil
		}
		results = append(results, resp.Value...)
	}

	return results
}

func getIotHubStats(ctx context.Context, cred azcore.TokenCredential) (*armiothub.RegistryStatistics, error) {
	iotHubResourceClient, err := armiothub.NewResourceClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := iotHubResourceClient.GetStats(ctx, resourceGroupName, iotHubName, nil)
	if err != nil {
		return nil, err
	}

	return &resp.RegistryStatistics, nil
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

	_, err = pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}
	return nil
}
