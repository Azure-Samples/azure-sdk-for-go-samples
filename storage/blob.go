// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package storage

import (
	"context"
	"io/ioutil"

	blob "github.com/Azure/azure-storage-blob-go/2016-05-31/azblob"
)

func getBlobURL(ctx context.Context, accountName, containerName, blobName string) blob.BlobURL {
	container := getContainerURL(ctx, accountName, containerName)
	blob := container.NewBlobURL(blobName)
	return blob
}

func GetBlob(ctx context.Context, accountName, containerName, blobName string) (string, error) {
	b := getBlobURL(ctx, accountName, containerName, blobName)

	resp, err := b.GetBlob(ctx, blob.BlobRange{}, blob.BlobAccessConditions{}, false)
	if err != nil {
		return "", err
	}
	defer resp.Body().Close()
	body, err := ioutil.ReadAll(resp.Body())
	return string(body), err
}
