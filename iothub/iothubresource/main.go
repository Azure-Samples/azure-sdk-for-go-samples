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
	resourceGroupName = "sample-resource-group3"
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

	//keepResource := os.Getenv("KEEP_RESOURCE")
	//if len(keepResource) == 0 {
	//	_, err := cleanup(ctx, cred)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	log.Println("cleaned up successfully.")
	//}
}

func createIotHubResource(ctx context.Context, cred azcore.TokenCredential) (*armiothub.IotHubDescription, error) {
	iotHubResourceClient := armiothub.NewIotHubResourceClient(subscriptionID, cred, nil)

	pollerResp, err := iotHubResourceClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		iotHubName,
		armiothub.IotHubDescription{
			Resource: armiothub.Resource{
				Location: to.StringPtr(location),
			},
			SKU: &armiothub.IotHubSKUInfo{
				Name:     armiothub.IotHubSKUS1.ToPtr(),
				Capacity: to.Int64Ptr(1),
			},
			Properties: &armiothub.IotHubProperties{
				EnableFileUploadNotifications: to.BoolPtr(false),
				Features:                      armiothub.CapabilitiesNone.ToPtr(),
				//NetworkRuleSets: &armiothub.NetworkRuleSetProperties{
				//	DefaultAction:                  armiothub.DefaultActionDeny.ToPtr(),
				//	ApplyToBuiltInEventHubEndpoint: to.BoolPtr(true),
				//	IPRules: []*armiothub.NetworkRuleSetIPRule{
				//		{
				//			FilterName: to.StringPtr("rule1"),
				//			Action:     armiothub.NetworkRuleIPActionAllow.ToPtr(),
				//			IPMask:     to.StringPtr("0.0.0.0"),
				//		},
				//	},
				//},
				//EventHubEndpoints: map[string]*armiothub.EventHubProperties{
				//	"events": {
				//		RetentionTimeInDays: to.Int64Ptr(1),
				//		PartitionCount:      to.Int32Ptr(2),
				//	},
				//},
				//StorageEndpoints: map[string]*armiothub.StorageEndpointProperties{
				//	"$default": {
				//		SasTTLAsIso8601:  to.StringPtr("PT1H"),
				//		ConnectionString: to.StringPtr(""),
				//		ContainerName:    to.StringPtr(""),
				//	},
				//},
				//MessagingEndpoints: map[string]*armiothub.MessagingEndpointProperties{
				//	"fileNotifications": {
				//		LockDurationAsIso8601: to.StringPtr("PT1M"),
				//		TTLAsIso8601:          to.StringPtr("PT1H"),
				//		MaxDeliveryCount:      to.Int32Ptr(10),
				//	},
				//},
				CloudToDevice: &armiothub.CloudToDeviceProperties{
					MaxDeliveryCount:    to.Int32Ptr(10),
					DefaultTTLAsIso8601: to.StringPtr("PT1H"),
					Feedback: &armiothub.FeedbackProperties{
						LockDurationAsIso8601: to.StringPtr("PT1M"),
						TTLAsIso8601:          to.StringPtr("PT1H"),
						MaxDeliveryCount:      to.Int32Ptr(10),
					},
				},
			},
		},
		nil,
	)
	if err != nil {
		log.Println("x")
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		log.Println("y")
		return nil, err
	}
	return &resp.IotHubDescription, nil
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
