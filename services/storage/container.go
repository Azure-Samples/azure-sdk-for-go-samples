// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package storage

import (
	"context"
	"fmt"
	"net/url"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/internal/config"
	"github.com/Azure/azure-storage-blob-go/azblob"
)

var (
	blobFormatString = `https://%s.blob.core.windows.net`
)

func getContainerURL(ctx context.Context, accountName, accountGroupName, containerName string) azblob.ContainerURL {
	key := getAccountPrimaryKey(ctx, accountName, accountGroupName)
	c, _ := azblob.NewSharedKeyCredential(accountName, key)
	p := azblob.NewPipeline(c, azblob.PipelineOptions{
		Telemetry: azblob.TelemetryOptions{Value: config.UserAgent()},
	})
	u, _ := url.Parse(fmt.Sprintf(blobFormatString, accountName))
	service := azblob.NewServiceURL(*u, p)
	container := service.NewContainerURL(containerName)
	return container
}

// CreateContainer creates a new container with the specified name in the specified account
func CreateContainer(ctx context.Context, accountName, accountGroupName, containerName string) (azblob.ContainerURL, error) {
	c := getContainerURL(ctx, accountName, accountGroupName, containerName)

	_, err := c.Create(
		ctx,
		azblob.Metadata{},
		azblob.PublicAccessContainer)
	return c, err
}

// GetContainer gets info about an existing container.
func GetContainer(ctx context.Context, accountName, accountGroupName, containerName string) (azblob.ContainerURL, error) {
	c := getContainerURL(ctx, accountName, accountGroupName, containerName)

	_, err := c.GetProperties(ctx, azblob.LeaseAccessConditions{})
	return c, err
}

// DeleteContainer deletes the named container.
func DeleteContainer(ctx context.Context, accountName, accountGroupName, containerName string) error {
	c := getContainerURL(ctx, accountName, accountGroupName, containerName)

	_, err := c.Delete(ctx, azblob.ContainerAccessConditions{})
	return err
}

// ListBlobs lists blobs on the specified container
func ListBlobs(ctx context.Context, accountName, accountGroupName, containerName string) (*azblob.ListBlobsFlatSegmentResponse, error) {
	c := getContainerURL(ctx, accountName, accountGroupName, containerName)
	return c.ListBlobsFlatSegment(
		ctx,
		azblob.Marker{},
		azblob.ListBlobsSegmentOptions{
			Details: azblob.BlobListingDetails{
				Snapshots: true,
			},
		})
}
