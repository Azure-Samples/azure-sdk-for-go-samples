// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package storage

import (
	"bytes"
	"context"

	"github.com/Azure/azure-storage-blob-go/azblob"
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
	copy(newPage, page)

	_, err := b.UploadPages(ctx, int64(pages*azblob.PageBlobPageBytes),
		bytes.NewReader(newPage),
		azblob.PageBlobAccessConditions{},
		nil,
	)
	return err
}

// ClearPage clears the specified page in the page blob
func ClearPage(ctx context.Context, accountName, accountGroupName, containerName, blobName string, pageNumber int) error {
	b := getPageBlobURL(ctx, accountName, accountGroupName, containerName, blobName)

	_, err := b.ClearPages(ctx,
		int64(pageNumber*azblob.PageBlobPageBytes),
		int64(azblob.PageBlobPageBytes),
		azblob.PageBlobAccessConditions{},
	)
	return err
}

// GetPageRanges gets a list of valid page ranges in the page blob
func GetPageRanges(ctx context.Context, accountName, accountGroupName, containerName, blobName string, pages int) (*azblob.PageList, error) {
	b := getPageBlobURL(ctx, accountName, accountGroupName, containerName, blobName)
	return b.GetPageRanges(
		ctx,
		0*azblob.PageBlobPageBytes,
		int64(pages*azblob.PageBlobPageBytes-1),
		azblob.BlobAccessConditions{},
	)
}
