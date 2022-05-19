// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"log"
	"os"
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
	ctx := context.Background()

	resourceGroup, err := createResourceGroup(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	disk, err := createDisk(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("virtual disk:", *disk.ID)

	snapshot, err := createSnapshot(ctx, cred, *disk.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("snapshot:", *snapshot.ID)

	image, err := createImage(ctx, cred, *snapshot.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("image:", *image.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createDisk(ctx context.Context, cred azcore.TokenCredential) (*armcompute.Disk, error) {
	disksClient, err := armcompute.NewDisksClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := disksClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		diskName,
		armcompute.Disk{
			Location: to.Ptr(location),
			SKU: &armcompute.DiskSKU{
				Name: to.Ptr(armcompute.DiskStorageAccountTypesStandardLRS),
			},
			Properties: &armcompute.DiskProperties{
				CreationData: &armcompute.CreationData{
					CreateOption: to.Ptr(armcompute.DiskCreateOptionEmpty),
				},
				DiskSizeGB: to.Ptr[int32](64),
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

	return &resp.Disk, nil
}

func createSnapshot(ctx context.Context, cred azcore.TokenCredential, diskID string) (*armcompute.Snapshot, error) {
	snapshotClient, err := armcompute.NewSnapshotsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := snapshotClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		snapshotName,
		armcompute.Snapshot{
			Location: to.Ptr(location),
			Properties: &armcompute.SnapshotProperties{
				CreationData: &armcompute.CreationData{
					CreateOption:     to.Ptr(armcompute.DiskCreateOptionCopy),
					SourceResourceID: to.Ptr(diskID),
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

	return &resp.Snapshot, nil
}

func createImage(ctx context.Context, cred azcore.TokenCredential, snapshotID string) (*armcompute.Image, error) {
	snapshotClient, err := armcompute.NewImagesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := snapshotClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		imageName,
		armcompute.Image{
			Location: to.Ptr(location),
			Properties: &armcompute.ImageProperties{
				StorageProfile: &armcompute.ImageStorageProfile{
					OSDisk: &armcompute.ImageOSDisk{
						OSType: to.Ptr(armcompute.OperatingSystemTypesWindows),
						Snapshot: &armcompute.SubResource{
							ID: to.Ptr(snapshotID),
						},
						OSState: to.Ptr(armcompute.OperatingSystemStateTypesGeneralized),
					},
					ZoneResilient: to.Ptr(false),
				},
				HyperVGeneration: to.Ptr(armcompute.HyperVGenerationTypesV1),
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

	return &resp.Image, nil
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
