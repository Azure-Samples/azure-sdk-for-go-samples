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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventgrid/armeventgrid"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID       string
	location             = "westus"
	resourceGroupName    = "sample-resource-group"
	partnerNamespaceName = "sample-partner-namespace"
	eventChannelName     = "sample-event-channel"
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

	partnerNamespace, err := createPartnerNamespace(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("partner namespace:", *partnerNamespace.ID)

	eventChannel, err := createEventChannel(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("event channel:", *eventChannel.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createPartnerNamespace(ctx context.Context, conn *arm.Connection) (*armeventgrid.PartnerNamespace, error) {
	partnerNamespacesClient := armeventgrid.NewPartnerNamespacesClient(conn, subscriptionID)

	pollerResp, err := partnerNamespacesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		partnerNamespaceName,
		armeventgrid.PartnerNamespace{
			TrackedResource: armeventgrid.TrackedResource{
				Location: to.StringPtr(location),
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
	return &resp.PartnerNamespace, nil
}

func createEventChannel(ctx context.Context, conn *arm.Connection) (*armeventgrid.EventChannel, error) {
	eventChannelsClient := armeventgrid.NewEventChannelsClient(conn, subscriptionID)

	resp, err := eventChannelsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		partnerNamespaceName,
		eventChannelName,
		armeventgrid.EventChannel{
			Properties: &armeventgrid.EventChannelProperties{
				Destination: &armeventgrid.EventChannelDestination{
					AzureSubscriptionID: to.StringPtr(subscriptionID),
					PartnerTopicName:    to.StringPtr(partnerNamespaceName),
					ResourceGroup:       to.StringPtr(resourceGroupName),
				},
				Source: &armeventgrid.EventChannelSource{
					Source: to.StringPtr("ContosoCorp.Accounts.User1"),
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.EventChannel, nil
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
