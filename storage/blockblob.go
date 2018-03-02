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

func getBlockBlobURL(ctx context.Context, accountName, containerName, blobName string) blob.BlockBlobURL {
	container := getContainerURL(ctx, accountName, containerName)
	blob := container.NewBlockBlobURL(blobName)
	return blob
}

// CreateBlockBlob creates a new test blob in the container specified by env var
func CreateBlockBlob(ctx context.Context, accountName, containerName, blobName string) (blob.BlockBlobURL, error) {
	b := getBlockBlobURL(ctx, accountName, containerName, blobName)
	data := "blob created by Azure-Samples, okay to delete!"

	_, err := b.PutBlob(
		context.Background(),
		strings.NewReader(data),
		blob.BlobHTTPHeaders{
			ContentType: "text/plain",
		},
		blob.Metadata{},
		blob.BlobAccessConditions{},
	)

	return b, err
}
