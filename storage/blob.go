// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package storage

import (
	"context"
	"io/ioutil"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-06-01/storage"
	"github.com/Azure/azure-storage-blob-go/azblob"
)

func getBlobClient() storage.BlobServicesClient {
	blobClient := storage.NewBlobServicesClient(config.SubscriptionID())
	auth, _ := iam.GetResourceManagementAuthorizer()
	blobClient.Authorizer = auth
	blobClient.AddToUserAgent(config.UserAgent())
	return blobClient
}

func getObjRepClient() storage.ObjectReplicationPoliciesClient {
	objRepClient := storage.NewObjectReplicationPoliciesClient(config.SubscriptionID())
	auth, _ := iam.GetResourceManagementAuthorizer()
	objRepClient.Authorizer = auth
	objRepClient.AddToUserAgent(config.UserAgent())
	return objRepClient
}

func getBlobURL(ctx context.Context, accountName, accountGroupName, containerName, blobName string) azblob.BlobURL {
	container := getContainerURL(ctx, accountName, accountGroupName, containerName)
	blob := container.NewBlobURL(blobName)
	return blob
}

// GetBlob downloads the specified blob contents
func GetBlob(ctx context.Context, accountName, accountGroupName, containerName, blobName string) (string, error) {
	b := getBlobURL(ctx, accountName, accountGroupName, containerName, blobName)

	resp, err := b.Download(ctx, 0, azblob.CountToEnd, azblob.BlobAccessConditions{}, false)

	if err != nil {
		return "", err
	}
	defer resp.Response().Body.Close()
	body, err := ioutil.ReadAll(resp.Body(azblob.RetryReaderOptions{}))
	return string(body), err
}
