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

func ExampleAppendBlobOperations() {
	accountName = getAccountName()
	containerName = strings.ToLower(containerName)

	helpers.SetResourceGroupName("UploadBlockBlob")
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

	_, err = CreateAppendBlob(ctx, accountName, containerName, blobName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created append blob")

	messages := []string{"Hello", "World!", "Hello", "Galaxy!"}
	for _, m := range messages {
		err = AppendToBlob(ctx, accountName, containerName, blobName, m)
		if err != nil {
			helpers.PrintAndLog(err.Error())
		}
		helpers.PrintAndLog("appended data to blob")
	}

	message, err := GetBlob(ctx, accountName, containerName, blobName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("downloaded blob")
	helpers.PrintAndLog(message)

	// Output:
	// created storage account
	// created container
	// created append blob
	// appended data to blob
	// appended data to blob
	// appended data to blob
	// appended data to blob
	// downloaded blob
	// HelloWorld!HelloGalaxy!
}
