package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/armcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/to"
)

var (
	subscriptionID              string
	TenantID                    string
	location                    = "westus"
	resourceGroupName           = "sample-resource-group"
	proximityPlacementGroupName = "sample-proximity-placement"
)

func main() {
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set.")
	}

	TenantID = os.Getenv("AZURE_TENANT_ID")
	if len(TenantID) == 0 {
		log.Fatal("AZURE_TENANT_ID is not set.")
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}

	conn := armcore.NewDefaultConnection(cred, &armcore.ConnectionOptions{
		Logging: azcore.LogOptions{
			IncludeBody: true,
		},
	})
	ctx := context.Background()

	resourceGroup, err := createResourceGroup(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	proximityPlacement, err := createProximityPlacement(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("proximity placement group:", *proximityPlacement.ID)

	proximityPlacement, err = getProximityPlacement(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get proximity placement group:", *proximityPlacement.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createProximityPlacement(ctx context.Context, conn *armcore.Connection) (*armcompute.ProximityPlacementGroup, error) {
	proximityPlacementGroupClient := armcompute.NewProximityPlacementGroupsClient(conn, subscriptionID)

	resp, err := proximityPlacementGroupClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		proximityPlacementGroupName,
		armcompute.ProximityPlacementGroup{
			Resource: armcompute.Resource{
				Location: to.StringPtr(location),
			},
			Properties: &armcompute.ProximityPlacementGroupProperties{
				ProximityPlacementGroupType: armcompute.ProximityPlacementGroupTypeStandard.ToPtr(),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return resp.ProximityPlacementGroup, nil
}

func getProximityPlacement(ctx context.Context, conn *armcore.Connection) (*armcompute.ProximityPlacementGroup, error) {
	proximityPlacementGroupClient := armcompute.NewProximityPlacementGroupsClient(conn, subscriptionID)

	resp, err := proximityPlacementGroupClient.Get(ctx, resourceGroupName, proximityPlacementGroupName, nil)
	if err != nil {
		return nil, err
	}

	return resp.ProximityPlacementGroup, nil
}

func createResourceGroup(ctx context.Context, conn *armcore.Connection) (*armresources.ResourceGroup, error) {
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
	return resourceGroupResp.ResourceGroup, nil
}

func cleanup(ctx context.Context, conn *armcore.Connection) (*http.Response, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn, subscriptionID)

	pollerResp, err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
