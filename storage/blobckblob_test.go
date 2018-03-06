// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

// Example creates a resource group and a storage account. Then it adds a container and a blob in that account.
// Finally it removes the blob, container, account, and group.
// more examples available at https://github.com/Azure/azure-storage-blob-go/2016-05-31/azblob/zt_examples_test.go
func ExampleBlockBlobOperations() {
	accountName = getAccountName()
	containerName = strings.ToLower(containerName)

	helpers.SetResourceGroupName("BlockBlob")
	ctx := context.Background()
	defer resources.Cleanup(ctx)
	_, err := resources.CreateGroup(ctx, helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}

	_, err = CreateStorageAccount(ctx, accountName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created storage account")

	_, err = CreateContainer(ctx, accountName, containerName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created container")

	_, err = CreateBlockBlob(ctx, accountName, containerName, blobName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created blob")

	for i, m := range messages {
		err = PutBlockOnBlob(ctx, accountName, containerName, blobName, m, i)
		if err != nil {
			helpers.PrintAndLog(err.Error())
		}
		helpers.PrintAndLog("put block")
	}

	list, err := GetUncommitedBlocks(ctx, accountName, containerName, blobName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog(fmt.Sprintf("list of uncommitted blocks has %d elements", len(list.UncommittedBlocks)))

	err = CommitBlocks(ctx, accountName, containerName, blobName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("committed blocks")

	message, err := GetBlob(ctx, accountName, containerName, blobName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("downloaded blob")
	helpers.PrintAndLog(message)

	// Output:
	// created storage account
	// created container
	// created blob
	// put block
	// put block
	// put block
	// put block
	// list of uncommitted blocks has 4 elements
	// committed blocks
	// downloaded blob
	// HelloWorld!HelloGalaxy!
}
