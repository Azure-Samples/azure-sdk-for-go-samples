// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package storage

import (
	"bytes"
	"context"

	blob "github.com/Azure/azure-storage-blob-go/2016-05-31/azblob"
)

func getPageBlobURL(ctx context.Context, accountName, containerName, blobName string) blob.PageBlobURL {
	container := getContainerURL(ctx, accountName, containerName)
	blob := container.NewPageBlobURL(blobName)
	return blob
}

// CreatePageBlob creates a new test blob in the container specified by env var
func CreatePageBlob(ctx context.Context, accountName, containerName, blobName string, pages int) (blob.PageBlobURL, error) {
	b := getPageBlobURL(ctx, accountName, containerName, blobName)

	_, err := b.Create(
		ctx,
		int64(blob.PageBlobPageBytes*pages),
		0,
		blob.BlobHTTPHeaders{
			ContentType: "text/plain",
		},
		blob.Metadata{},
		blob.BlobAccessConditions{},
	)
	return b, err
}

func PutPage(ctx context.Context, accountName, containerName, blobName, message string, page int) error {
	b := getPageBlobURL(ctx, accountName, containerName, blobName)

	fullMessage := make([]byte, blob.PageBlobPageBytes)
	for i, e := range []byte(message) {
		fullMessage[i] = e
	}

	_, err := b.PutPages(ctx,
		blob.PageRange{
			Start: int32(page * blob.PageBlobPageBytes),
			End:   int32((page+1)*blob.PageBlobPageBytes - 1),
		},
		bytes.NewReader(fullMessage),
		blob.BlobAccessConditions{},
	)
	return err
}

func ClearPage(ctx context.Context, accountName, containerName, blobName string, page int) error {
	b := getPageBlobURL(ctx, accountName, containerName, blobName)

	_, err := b.ClearPages(ctx,
		blob.PageRange{
			Start: int32(page * blob.PageBlobPageBytes),
			End:   int32((page+1)*blob.PageBlobPageBytes - 1),
		},
		blob.BlobAccessConditions{},
	)
	return err
}

func GetPageRanges(ctx context.Context, accountName, containerName, blobName string, pages int) (*blob.PageList, error) {
	b := getPageBlobURL(ctx, accountName, containerName, blobName)
	return b.GetPageRanges(
		ctx,
		blob.BlobRange{
			Offset: 0 * blob.PageBlobPageBytes,
			Count:  int64(pages*blob.PageBlobPageBytes - 1),
		},
		blob.BlobAccessConditions{},
	)
}
