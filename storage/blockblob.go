// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package storage

import (
	"context"
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

func getBlockBlobURL(ctx context.Context, accountName, accountGroupName, containerName, blobName string) azblob.BlockBlobURL {
	container := getContainerURL(ctx, accountName, accountGroupName, containerName)
	blob := container.NewBlockBlobURL(blobName)
	return blob
}

// CreateBlockBlob creates a new block blob
func CreateBlockBlob(ctx context.Context, accountName, accountGroupName, containerName, blobName string) (azblob.BlockBlobURL, error) {
	b := getBlockBlobURL(ctx, accountName, accountGroupName, containerName, blobName)
	data := "blob created by Azure-Samples, okay to delete!"

	_, err := b.Upload(
		ctx,
		strings.NewReader(data),
		azblob.BlobHTTPHeaders{
			ContentType: "text/plain",
		},
		azblob.Metadata{},
		azblob.BlobAccessConditions{},
	)

	return b, err
}

// PutBlockOnBlob adds a block to a block blob. It does not commit the block.
func PutBlockOnBlob(ctx context.Context, accountName, accountGroupName, containerName, blobName, message string, blockNum int) error {
	b := getBlockBlobURL(ctx, accountName, accountGroupName, containerName, blobName)
	id := base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(blockNum)))
	_, err := b.StageBlock(ctx, id, strings.NewReader(message), azblob.LeaseAccessConditions{}, nil)
	return err
}

// GetUncommitedBlocks gets a list of uncommited blobs
func GetUncommitedBlocks(ctx context.Context, accountName, accountGroupName, containerName, blobName string) (*azblob.BlockList, error) {
	b := getBlockBlobURL(ctx, accountName, accountGroupName, containerName, blobName)
	return b.GetBlockList(ctx, azblob.BlockListUncommitted, azblob.LeaseAccessConditions{})
}

// CommitBlocks commits the uncommitted blocks to the blob
func CommitBlocks(ctx context.Context, accountName, accountGroupName, containerName, blobName string) error {
	b := getBlockBlobURL(ctx, accountName, accountGroupName, containerName, blobName)
	list, err := GetUncommitedBlocks(ctx, accountName, accountGroupName, containerName, blobName)
	if err != nil {
		return err
	}

	IDs := []string{}
	for _, u := range list.UncommittedBlocks {
		IDs = append(IDs, u.Name)
	}

	_, err = b.CommitBlockList(ctx, IDs, azblob.BlobHTTPHeaders{}, azblob.Metadata{}, azblob.BlobAccessConditions{})
	return err
}
