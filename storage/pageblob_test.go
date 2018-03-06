// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package storage

import (
	"context"
	"strings"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

func ExamplePageBlobOperations() {
	accountName = getAccountName()
	containerName = strings.ToLower(containerName)

	helpers.SetResourceGroupName("PageBlob")
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

	_, err = CreatePageBlob(ctx, accountName, containerName, blobName, len(messages))
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created page blob")

	for i, m := range messages {
		err = PutPage(ctx, accountName, containerName, blobName, m, i)
		if err != nil {
			helpers.PrintAndLog(err.Error())
		}
		helpers.PrintAndLog("put page")
	}

	err = ClearPage(ctx, accountName, containerName, blobName, 2)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("cleared page")

	_, err = GetPageRanges(ctx, accountName, containerName, blobName, len(messages))
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("got page ranges")

	message, err := GetBlob(ctx, accountName, containerName, blobName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("downloaded blob")
	var empty byte
	helpers.PrintAndLog(strings.Replace(message, string([]byte{empty}), "", -1))

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
