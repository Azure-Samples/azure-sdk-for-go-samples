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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID       string
	location             = "westus"
	resourceGroupName    = "sample-resource-group"
	activityLogAlertName = "sample-activity-log-alert"
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

	activityLogAlert, err := createActivityLogAlert(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("activity log alert:", *activityLogAlert.ID)

	activityLogAlert, err = getActivityLogAlert(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get activity log alert:", *activityLogAlert.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createActivityLogAlert(ctx context.Context, cred azcore.TokenCredential) (*armmonitor.ActivityLogAlertResource, error) {
	activityLogAlert := armmonitor.NewActivityLogAlertsClient(subscriptionID, cred, nil)

	resp, err := activityLogAlert.CreateOrUpdate(
		ctx,
		resourceGroupName,
		activityLogAlertName,
		armmonitor.ActivityLogAlertResource{
			Resource: armmonitor.Resource{
				Location: to.StringPtr("global"),
			},
			Properties: &armmonitor.ActivityLogAlert{
				Scopes: []*string{
					to.StringPtr("subscriptions/" + subscriptionID),
				},
				Enabled: to.BoolPtr(true),
				Condition: &armmonitor.ActivityLogAlertAllOfCondition{
					AllOf: []*armmonitor.ActivityLogAlertLeafCondition{
						{
							Field:  to.StringPtr("category"),
							Equals: to.StringPtr("Administrative"),
						},
						{
							Field:  to.StringPtr("level"),
							Equals: to.StringPtr("Error"),
						},
					},
				},
				Actions: &armmonitor.ActivityLogAlertActionList{
					ActionGroups: []*armmonitor.ActivityLogAlertActionGroup{},
				},
				Description: to.StringPtr("Sample activity log alert description"),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.ActivityLogAlertResource, nil
}

func getActivityLogAlert(ctx context.Context, cred azcore.TokenCredential) (*armmonitor.ActivityLogAlertResource, error) {
	activityLogAlert := armmonitor.NewActivityLogAlertsClient(subscriptionID, cred, nil)

	resp, err := activityLogAlert.Get(ctx, resourceGroupName, activityLogAlertName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ActivityLogAlertResource, nil
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
