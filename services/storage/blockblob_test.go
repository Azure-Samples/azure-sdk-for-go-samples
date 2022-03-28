// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/internal/util"
)

func Example_blockBlobOperations() {
	var accountName = testAccountName
	var accountGroupName = testAccountGroupName
	var containerName = generateName("test-blockblobc")
	var blobName = generateName("test-blockblob")
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()

	_, err = CreateContainer(ctx, accountName, accountGroupName, containerName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created container")

	_, err = CreateBlockBlob(ctx, accountName, accountGroupName, containerName, blobName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created blob")

	blocks := []string{"Hello", "World!", "Hello", "Galaxy!"}
	for i, block := range blocks {
		err = PutBlockOnBlob(ctx, accountName, accountGroupName, containerName, blobName, block, i)
		if err != nil {
			util.LogAndPanic(err)
		}
		util.PrintAndLog(fmt.Sprintf("put block %d", i))
	}

	list, err := GetUncommitedBlocks(ctx, accountName, accountGroupName, containerName, blobName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog(fmt.Sprintf(
		"list of uncommitted blocks has %d elements",
		len(list.UncommittedBlocks)))

	err = CommitBlocks(ctx, accountName, accountGroupName, containerName, blobName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("committed blocks")

	blob, err := GetBlob(ctx, accountName, accountGroupName, containerName, blobName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("downloaded blob")
	util.PrintAndLog(blob)

	// Output:
	// created container
	// created blob
	// put block 0
	// put block 1
	// put block 2
	// put block 3
	// list of uncommitted blocks has 4 elements
	// committed blocks
	// downloaded blob
	// HelloWorld!HelloGalaxy!
}
