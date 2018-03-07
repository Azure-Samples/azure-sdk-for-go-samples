// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package storage

import (
	"context"
	"strings"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

func ExamplePageBlobOperations() {
	accountName = getAccountName()
	containerName = strings.ToLower(containerName)

	internal.SetResourceGroupName("PageBlob")
	ctx := context.Background()
	defer resources.Cleanup(ctx)
	_, err := resources.CreateGroup(ctx, internal.ResourceGroupName())
	if err != nil {
		internal.PrintAndLog(err.Error())
	}

	_, err = CreateStorageAccount(ctx, accountName)
	if err != nil {
		internal.PrintAndLog(err.Error())
	}
	internal.PrintAndLog("created storage account")

	_, err = CreateContainer(ctx, accountName, containerName)
	if err != nil {
		internal.PrintAndLog(err.Error())
	}
	internal.PrintAndLog("created container")

	_, err = CreatePageBlob(ctx, accountName, containerName, blobName, len(messages))
	if err != nil {
		internal.PrintAndLog(err.Error())
	}
	internal.PrintAndLog("created page blob")

	for i, m := range messages {
		err = PutPage(ctx, accountName, containerName, blobName, m, i)
		if err != nil {
			internal.PrintAndLog(err.Error())
		}
		internal.PrintAndLog("put page")
	}

	err = ClearPage(ctx, accountName, containerName, blobName, 2)
	if err != nil {
		internal.PrintAndLog(err.Error())
	}
	internal.PrintAndLog("cleared page")

	_, err = GetPageRanges(ctx, accountName, containerName, blobName, len(messages))
	if err != nil {
		internal.PrintAndLog(err.Error())
	}
	internal.PrintAndLog("got page ranges")

	message, err := GetBlob(ctx, accountName, containerName, blobName)
	if err != nil {
		internal.PrintAndLog(err.Error())
	}
	internal.PrintAndLog("downloaded blob")
	var empty byte
	internal.PrintAndLog(strings.Replace(message, string([]byte{empty}), "", -1))

	// Output:
	// created storage account
	// created container
	// created page blob
	// put page
	// put page
	// put page
	// put page
	// cleared page
	// got page ranges
	// downloaded blob
	// HelloWorld!Galaxy!
}
