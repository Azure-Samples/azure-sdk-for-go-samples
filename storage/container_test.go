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

func ExampleListBlobs() {
	accountName = getAccountName()
	containerName = strings.ToLower(containerName)

	helpers.SetResourceGroupName("ListBlobs")
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

	for i := 0; i < 3; i++ {
		name := fmt.Sprintf("blob%d", i)
		_, err = CreateBlockBlob(ctx, accountName, containerName, name)
		if err != nil {
			helpers.PrintAndLog(err.Error())
		}
		helpers.PrintAndLog("created blob")
	}

	list, err := ListBlobs(ctx, accountName, containerName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog(fmt.Sprintf("listed blobs: %d", len(list.Blobs.Blob)))

	// Output:
	// created storage account
	// created container
	// created blob
	// created blob
	// created blob
	// listed blobs: 3
}
