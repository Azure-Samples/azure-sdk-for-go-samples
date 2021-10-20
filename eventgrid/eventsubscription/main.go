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
	subscriptionID        string
	location              = "westus"
	resourceGroupName     = "sample-resource-group"
	eventSubscriptionName = "sample-event-subscription"
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

	eventSubscription, err := createEventSubscription(ctx, conn, *resourceGroup.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("event subscription:", *eventSubscription.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createEventSubscription(ctx context.Context, conn *arm.Connection, resourceID string) (*armeventgrid.EventSubscription, error) {
	eventSubscriptionsClient := armeventgrid.NewEventSubscriptionsClient(conn, subscriptionID)

	pollerResp, err := eventSubscriptionsClient.BeginCreateOrUpdate(
		ctx,
		resourceID,
		eventSubscriptionName,
		armeventgrid.EventSubscription{
			Properties: &armeventgrid.EventSubscriptionProperties{
				DeadLetterDestination: &armeventgrid.StorageBlobDeadLetterDestination{
					DeadLetterDestination: armeventgrid.DeadLetterDestination{
						EndpointType: armeventgrid.DeadLetterEndPointTypeStorageBlob.ToPtr(),
					},
				},
				Destination: &armeventgrid.WebHookEventSubscriptionDestination{
					EventSubscriptionDestination: armeventgrid.EventSubscriptionDestination{
						EndpointType: armeventgrid.EndpointTypeWebHook.ToPtr(),
					},
					//Properties: &armeventgrid.WebHookEventSubscriptionDestinationProperties{
					//	EndpointURL: to.StringPtr("https://www.example.com"),
					//},
				},
				EventDeliverySchema: armeventgrid.EventDeliverySchemaEventGridSchema.ToPtr(),
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
	return &resp.EventSubscription, nil
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
