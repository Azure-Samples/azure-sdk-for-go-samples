// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package storage

import (
	"bytes"
	"context"

	"github.com/Azure/azure-storage-blob-go/2016-05-31/azblob"
)

func getPageBlobURL(ctx context.Context, accountName, accountGroupName, containerName, blobName string) azblob.PageBlobURL {
	container := getContainerURL(ctx, accountName, accountGroupName, containerName)
	blob := container.NewPageBlobURL(blobName)
	return blob
}

// CreatePageBlob creates a new test blob in the container specified.
func CreatePageBlob(ctx context.Context, accountName, accountGroupName, containerName, blobName string, pages int) (azblob.PageBlobURL, error) {
	b := getPageBlobURL(ctx, accountName, accountGroupName, containerName, blobName)

	_, err := b.Create(
		ctx,
		int64(pages*azblob.PageBlobPageBytes),
		0,
		azblob.BlobHTTPHeaders{
			ContentType: "text/plain",
		},
		azblob.Metadata{},
		azblob.BlobAccessConditions{},
	)
	return b, err
}

// PutPage adds a page to the page blob
// TODO: page should be []byte
func PutPage(ctx context.Context, accountName, accountGroupName, containerName, blobName, page string, pages int) error {
	b := getPageBlobURL(ctx, accountName, accountGroupName, containerName, blobName)

	newPage := make([]byte, azblob.PageBlobPageBytes)
	for i, c := range []byte(page) {
		newPage[i] = c
	}

	_, err := b.PutPages(ctx,
		azblob.PageRange{
			Start: int32(pages * azblob.PageBlobPageBytes),
			End:   int32((pages+1)*azblob.PageBlobPageBytes - 1),
		},
		bytes.NewReader(newPage),
		azblob.BlobAccessConditions{},
	)
	return err
}

// ClearPage clears the specified page in the page blob
func ClearPage(ctx context.Context, accountName, accountGroupName, containerName, blobName string, pageNumber int) error {
	b := getPageBlobURL(ctx, accountName, accountGroupName, containerName, blobName)

	_, err := b.ClearPages(ctx,
		azblob.PageRange{
			Start: int32(pageNumber * azblob.PageBlobPageBytes),
			End:   int32((pageNumber+1)*azblob.PageBlobPageBytes - 1),
		},
		azblob.BlobAccessConditions{},
	)
	return err
}

// GetPageRanges gets a list of valid page ranges in the page blob
func GetPageRanges(ctx context.Context, accountName, accountGroupName, containerName, blobName string, pages int) (*azblob.PageList, error) {
	b := getPageBlobURL(ctx, accountName, accountGroupName, containerName, blobName)
	return b.GetPageRanges(
		ctx,
		azblob.BlobRange{
			Offset: 0 * azblob.PageBlobPageBytes,
			Count:  int64(pages*azblob.PageBlobPageBytes - 1),
		},
		azblob.BlobAccessConditions{},
	)
}
