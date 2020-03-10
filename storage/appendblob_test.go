// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package storage

import (
	"context"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
)

func Example_appendBlobOperations() {
	var accountName = testAccountName
	var accountGroupName = testAccountGroupName
	var containerName = generateName("test-appendblobc")
	var blobName = generateName("test-appendblob")
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()

	_, err = CreateContainer(ctx, accountName, accountGroupName, containerName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created container")

	_, err = CreateAppendBlob(ctx, accountName, accountGroupName, containerName, blobName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created append blob")

	blocks := []string{"Hello", "World!", "Hello", "Galaxy!"}
	for _, block := range blocks {
		err = AppendToBlob(ctx, accountName, accountGroupName, containerName, blobName, block)
		if err != nil {
			util.LogAndPanic(err)
		}
		util.PrintAndLog("appended data to blob")
	}

	blob, err := GetBlob(ctx, accountName, accountGroupName, containerName, blobName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("got blob")
	util.PrintAndLog(blob)

	// Output:
	// created container
	// created append blob
	// appended data to blob
	// appended data to blob
	// appended data to blob
	// appended data to blob
	// got blob
	// HelloWorld!HelloGalaxy!
}
