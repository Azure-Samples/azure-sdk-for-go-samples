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
	subscriptionID    string
	TenantID          string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	diskName          = "sample-disk"
	snapshotName      = "sample-snapshot"
	imageName         = "sample-image"
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

	disk, err := createDisk(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("virtual disk:", *disk.ID)

	snapshot, err := createSnapshot(ctx, conn, *disk.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("snapshot:", *snapshot.ID)

	image, err := createImage(ctx, conn, *snapshot.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("image:", *image.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createDisk(ctx context.Context, conn *armcore.Connection) (*armcompute.Disk, error) {
	disksClient := armcompute.NewDisksClient(conn, subscriptionID)

	pollerResp, err := disksClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		diskName,
		armcompute.Disk{
			Resource: armcompute.Resource{
				Location: to.StringPtr(location),
			},
			SKU: &armcompute.DiskSKU{
				Name: armcompute.DiskStorageAccountTypesStandardLRS.ToPtr(),
			},
			Properties: &armcompute.DiskProperties{
				CreationData: &armcompute.CreationData{
					CreateOption: armcompute.DiskCreateOptionEmpty.ToPtr(),
				},
				DiskSizeGB: to.Int32Ptr(64),
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

	return resp.Disk, nil
}

func createSnapshot(ctx context.Context, conn *armcore.Connection, diskID string) (*armcompute.Snapshot, error) {
	snapshotClient := armcompute.NewSnapshotsClient(conn, subscriptionID)

	pollerResp, err := snapshotClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		snapshotName,
		armcompute.Snapshot{
			Resource: armcompute.Resource{
				Location: to.StringPtr(location),
			},
			Properties: &armcompute.SnapshotProperties{
				CreationData: &armcompute.CreationData{
					CreateOption:     armcompute.DiskCreateOptionCopy.ToPtr(),
					SourceResourceID: to.StringPtr(diskID),
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

	return resp.Snapshot, nil
}

func createImage(ctx context.Context, conn *armcore.Connection, snapshotID string) (*armcompute.Image, error) {
	snapshotClient := armcompute.NewImagesClient(conn, subscriptionID)

	pollerResp, err := snapshotClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		imageName,
		armcompute.Image{
			Resource: armcompute.Resource{
				Location: to.StringPtr(location),
			},
			Properties: &armcompute.ImageProperties{
				StorageProfile: &armcompute.ImageStorageProfile{
					OSDisk: &armcompute.ImageOSDisk{
						OSType: armcompute.OperatingSystemTypesWindows.ToPtr(),
						ImageDisk: armcompute.ImageDisk{
							Snapshot: &armcompute.SubResource{
								ID: to.StringPtr(snapshotID),
							},
						},
						OSState: armcompute.OperatingSystemStateTypesGeneralized.ToPtr(),
					},
					ZoneResilient: to.BoolPtr(false),
				},
				HyperVGeneration: armcompute.HyperVGenerationTypesV1.ToPtr(),
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

	return resp.Image, nil
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
