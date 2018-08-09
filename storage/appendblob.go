// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package storage

import (
	"context"
	"strings"

	"github.com/Azure/azure-storage-blob-go/2016-05-31/azblob"
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
	_, err := b.AppendBlock(ctx, strings.NewReader(message), azblob.BlobAccessConditions{})
	return err
}
