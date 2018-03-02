// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package storage

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
	"github.com/subosito/gotenv"
)

var (
	accountName   string
	containerName = "container1"
	blobName      = "blob1"
)

func TestMain(m *testing.M) {
	gotenv.Load()
	name := os.Getenv("AZ_STORAGE_ACCOUNT_NAME")
	if len(name) > 0 {
		accountName = name
	}

	flag.StringVar(&accountName, "storageAccoutName", getAccountName(), "Provide a name for the storage account to be created")
	flag.StringVar(&containerName, "containerName", containerName, "Provide a name for the container.")
	flag.StringVar(&blobName, "blobName", blobName, "Provide a name for the blob.")

	err := helpers.ParseArgs()
	if err != nil {
		log.Fatalln("failed to parse args")
	}

	os.Exit(m.Run())
}

// Example creates a resource group and a storage account. Then it adds a container and a blob in that account.
// Finally it removes the blob, container, account, and group.
// more examples available at https://github.com/Azure/azure-storage-blob-go/2016-05-31/azblob/zt_examples_test.go
func ExampleUploadBlockBlob() {
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

	_, err = CreateBlockBlob(ctx, accountName, containerName, blobName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created blob")

	// Output:
	// created storage account
	// created container
	// created blob
}
