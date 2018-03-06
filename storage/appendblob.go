// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package storage

import (
	"context"
	"strings"

	blob "github.com/Azure/azure-storage-blob-go/2016-05-31/azblob"
)

func getAppendBlobURL(ctx context.Context, accountName, containerName, blobName string) blob.AppendBlobURL {
	container := getContainerURL(ctx, accountName, containerName)
	blob := container.NewAppendBlobURL(blobName)
	return blob
}

// CreateAppendBlob creates an empty append blob
func CreateAppendBlob(ctx context.Context, accountName, containerName, blobName string) (blob.AppendBlobURL, error) {
	b := getAppendBlobURL(ctx, accountName, containerName, blobName)
	_, err := b.Create(ctx, blob.BlobHTTPHeaders{}, blob.Metadata{}, blob.BlobAccessConditions{})
	return b, err
}

// AppendToBlob appends new data to the specified append blob
func AppendToBlob(ctx context.Context, accountName, containerName, blobName, message string) error {
	b := getAppendBlobURL(ctx, accountName, containerName, blobName)
	_, err := b.AppendBlock(ctx, strings.NewReader(message), blob.BlobAccessConditions{})
	return err
}
