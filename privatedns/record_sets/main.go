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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID        string
	location              = "westus"
	resourceGroupName     = "sample-resource-group"
	privateZoneName       = "sample-private-zone"
	relativeRecordSetName = "sample-relative-record-set"
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

	privateZone, err := createPrivateZone(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("private zone:", *privateZone.ID)

	recordSets, err := createRecordSets(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("record sets:", *recordSets.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createPrivateZone(ctx context.Context, conn *arm.Connection) (*armprivatedns.PrivateZone, error) {
	privateZonesClient := armprivatedns.NewPrivateZonesClient(conn, subscriptionID)

	pollersResp, err := privateZonesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		privateZoneName,
		armprivatedns.PrivateZone{
			TrackedResource: armprivatedns.TrackedResource{
				Location: to.StringPtr(location),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollersResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.PrivateZone, nil
}

func createRecordSets(ctx context.Context, conn *arm.Connection) (*armprivatedns.RecordSet, error) {
	recordSets := armprivatedns.NewRecordSetsClient(conn, subscriptionID)

	resp, err := recordSets.CreateOrUpdate(
		ctx,
		resourceGroupName,
		privateZoneName,
		armprivatedns.RecordTypeA,
		relativeRecordSetName,
		armprivatedns.RecordSet{
			Properties: &armprivatedns.RecordSetProperties{
				ARecords: []*armprivatedns.ARecord{},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.RecordSet, nil
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
