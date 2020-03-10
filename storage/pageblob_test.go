// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
)

func Example_pageBlobOperations() {
	var accountName = testAccountName
	var accountGroupName = testAccountGroupName
	var containerName = generateName("test-pageblobc")
	var blobName = generateName("test-pageblob")
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()

	_, err = CreateContainer(ctx, accountName, accountGroupName, containerName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created container")

	pages := []string{"Hello", "World!", "Hello", "Galaxy!"}
	_, err = CreatePageBlob(ctx, accountName, accountGroupName, containerName, blobName, len(pages))
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created page blob")

	for i, page := range pages {
		err = PutPage(ctx, accountName, accountGroupName, containerName, blobName, page, i)
		if err != nil {
			util.LogAndPanic(err)
		}
		util.PrintAndLog(fmt.Sprintf("put page %d", i))
	}

	_, err = GetBlob(ctx, accountName, accountGroupName, containerName, blobName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("downloaded blob")
	// empty bytes are in fact mixed in between the strings
	// so although this appears to emit `HelloWorld!HelloGalaxy!`
	// it doesn't match the expected output
	// TODO: find a better way to test
	// util.PrintAndLog(string(blob))

	var pageToClear int = 2
	err = ClearPage(ctx, accountName, accountGroupName, containerName, blobName, pageToClear)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog(fmt.Sprintf("cleared page %d", pageToClear))

	_, err = GetPageRanges(ctx, accountName, accountGroupName, containerName, blobName, len(pages))
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("got page ranges")

	// Output:
	// created container
	// created page blob
	// put page 0
	// put page 1
	// put page 2
	// put page 3
	// downloaded blob
	// cleared page 2
	// got page ranges
}
