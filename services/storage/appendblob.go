// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package storage

import (
	"context"
	"strings"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

func getAppendBlobURL(ctx context.Context, accountName, accountGroupName, containerName, blobName string) azblob.AppendBlobURL {
	container := getContainerURL(ctx, accountName, accountGroupName, containerName)
	blob := container.NewAppendBlobURL(blobName)
	return blob
}

// CreateAppendBlob creates an empty append blob
func CreateAppendBlob(ctx context.Context, accountName, accountGroupName, containerName, blobName string) (azblob.AppendBlobURL, error) {
	b := getAppendBlobURL(ctx, accountName, accountGroupName, containerName, blobName)
	_, err := b.Create(ctx, azblob.BlobHTTPHeaders{}, azblob.Metadata{}, azblob.BlobAccessConditions{})
	return b, err
}

// AppendToBlob appends new data to the specified append blob
func AppendToBlob(ctx context.Context, accountName, accountGroupName, containerName, blobName, message string) error {
	b := getAppendBlobURL(ctx, accountName, accountGroupName, containerName, blobName)
	_, err := b.AppendBlock(ctx, strings.NewReader(message), azblob.AppendBlobAccessConditions{}, nil)
	return err
}
