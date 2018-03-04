// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package storage

import (
	"context"
	"encoding/base64"
	"strings"

	blob "github.com/Azure/azure-storage-blob-go/2016-05-31/azblob"
)

func getBlockBlobURL(ctx context.Context, accountName, containerName, blobName string) blob.BlockBlobURL {
	container := getContainerURL(ctx, accountName, containerName)
	blob := container.NewBlockBlobURL(blobName)
	return blob
}

// CreateBlockBlob creates a new block blob
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

// PutBlockOnBlob adds a block to a block blob. It does not commit the block.
func PutBlockOnBlob(ctx context.Context, accountName, containerName, blobName, message string, blockNum int) error {
	b := getBlockBlobURL(ctx, accountName, containerName, blobName)
	id := base64.StdEncoding.EncodeToString([]byte(string(blockNum)))
	_, err := b.PutBlock(ctx, id, strings.NewReader(message), blob.LeaseAccessConditions{})
	return err
}

// GetUncommitedBlocks gets a list of uncommited blobs 
func GetUncommitedBlocks(ctx context.Context, accountName, containerName, blobName string) (*blob.BlockList, error) {
	b := getBlockBlobURL(ctx, accountName, containerName, blobName)
	return b.GetBlockList(ctx, blob.BlockListUncommitted, blob.LeaseAccessConditions{})
}

// CommitBlocks commits the uncommitted blocks to the blob
func CommitBlocks(ctx context.Context, accountName, containerName, blobName string) error {
	b := getBlockBlobURL(ctx, accountName, containerName, blobName)
	list, err := GetUncommitedBlocks(ctx, accountName, containerName, blobName)
	if err != nil {
		return err
	}

	IDs := []string{}
	for _, u := range list.UncommittedBlocks {
		IDs = append(IDs, u.Name)
	}

	_, err = b.PutBlockList(ctx, IDs, blob.BlobHTTPHeaders{}, blob.Metadata{}, blob.BlobAccessConditions{})
	return err
}
