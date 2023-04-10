// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"log"
	"os"
)

var (
	subscriptionID         string
	TenantID               string
	location               = "westus"
	resourceGroupName      = "sample-resource-group"
	diskName               = "sample-disk"
	snapshotName           = "sample-snapshot"
	galleryName            = "sample_gallery"
	galleryApplicationName = "sample_gallery_application"
	galleryImageName       = "sample_gallery_image"
)

var (
	resourcesClientFactory *armresources.ClientFactory
	computeClientFactory   *armcompute.ClientFactory
)

var (
	resourceGroupClient *armresources.ResourceGroupsClient
	disksClient         *armcompute.DisksClient
	snapshotsClient     *armcompute.SnapshotsClient
	galleriesClient     *armcompute.GalleriesClient
	galleryImagesClient *armcompute.GalleryImagesClient
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

	resourcesClientFactory, err = armresources.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	resourceGroupClient = resourcesClientFactory.NewResourceGroupsClient()

	computeClientFactory, err = armcompute.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	disksClient = computeClientFactory.NewDisksClient()
	snapshotsClient = computeClientFactory.NewSnapshotsClient()
	galleriesClient = computeClientFactory.NewGalleriesClient()
	galleryImagesClient = computeClientFactory.NewGalleryImagesClient()

	resourceGroup, err := createResourceGroup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	disk, err := createDisk(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("disk:", *disk.ID)

	snapshot, err := createSnapshot(ctx, *disk.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("snapshot:", *snapshot.ID)

	gallery, err := createGallery(ctx, *disk.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("gallery:", *gallery.ID)

	galleryApplication, err := createGalleryApplication(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("gallery application:", *galleryApplication.ID)

	galleryImage, err := createGalleryImage(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("gallery image:", *galleryImage.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err := cleanup(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createDisk(ctx context.Context) (*armcompute.Disk, error) {

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

func createSnapshot(ctx context.Context, diskID string) (*armcompute.Snapshot, error) {

	pollerResp, err := snapshotsClient.BeginCreateOrUpdate(
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

func createGallery(ctx context.Context, diskID string) (*armcompute.Gallery, error) {

	pollerResp, err := galleriesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		galleryName,
		armcompute.Gallery{
			Location: to.Ptr(location),
			Properties: &armcompute.GalleryProperties{
				Description: to.Ptr("This is gallery description."),
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

	return &resp.Gallery, nil
}

func createGalleryApplication(ctx context.Context) (*armcompute.Gallery, error) {

	pollerResp, err := galleriesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		galleryName,
		armcompute.Gallery{
			Location: to.Ptr(location),
			Properties: &armcompute.GalleryProperties{
				Description: to.Ptr("This is the gallery application description."),
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

	return &resp.Gallery, nil
}

func createGalleryImage(ctx context.Context) (*armcompute.GalleryImage, error) {

	pollerResp, err := galleryImagesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		galleryName,
		galleryImageName,
		armcompute.GalleryImage{
			Location: to.Ptr(location),
			Properties: &armcompute.GalleryImageProperties{
				OSType:           to.Ptr(armcompute.OperatingSystemTypesWindows),
				OSState:          to.Ptr(armcompute.OperatingSystemStateTypesGeneralized),
				HyperVGeneration: to.Ptr(armcompute.HyperVGenerationV1),
				Identifier: &armcompute.GalleryImageIdentifier{
					Offer:     to.Ptr("myPublisherName"),
					Publisher: to.Ptr("myOfferName"),
					SKU:       to.Ptr("mySkuName"),
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

	return &resp.GalleryImage, nil
}

func createResourceGroup(ctx context.Context) (*armresources.ResourceGroup, error) {

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

func cleanup(ctx context.Context) error {

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
