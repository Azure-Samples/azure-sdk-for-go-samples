// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package storage

import (
	"context"
	"fmt"
	"net/url"

	"github.com/Azure/azure-storage-blob-go/2016-05-31/azblob"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
)

var (
	blobFormatString = `https://%s.blob.core.windows.net`
)

func getContainerURL(ctx context.Context, accountName, accountGroupName, containerName string) azblob.ContainerURL {
	key := getAccountPrimaryKey(ctx, accountName, accountGroupName)
	c := azblob.NewSharedKeyCredential(accountName, key)
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

	_, err := c.GetPropertiesAndMetadata(ctx, azblob.LeaseAccessConditions{})
	return c, err
}

// DeleteContainer deletes the named container.
func DeleteContainer(ctx context.Context, accountName, accountGroupName, containerName string) error {
	c := getContainerURL(ctx, accountName, accountGroupName, containerName)

	_, err := c.Delete(ctx, azblob.ContainerAccessConditions{})
	return err
}

// ListBlobs lists blobs on the specified container
func ListBlobs(ctx context.Context, accountName, accountGroupName, containerName string) (*azblob.ListBlobsResponse, error) {
	c := getContainerURL(ctx, accountName, accountGroupName, containerName)
	return c.ListBlobs(
		ctx,
		azblob.Marker{},
		azblob.ListBlobsOptions{
			Details: azblob.BlobListingDetails{
				Snapshots: true,
			},
		})
}
