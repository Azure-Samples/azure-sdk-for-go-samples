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

func Example_containerAndBlobs() {
	var accountName = testAccountName
	var accountGroupName = testAccountGroupName
	var containerName = generateName("test-blobc")
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()

	_, err = CreateContainer(ctx, accountName, accountGroupName, containerName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created container")

	for i := 0; i < 3; i++ {
		blobName := fmt.Sprintf("test-blob%d", i)
		_, err = CreateBlockBlob(ctx, accountName, accountGroupName, containerName, blobName)
		if err != nil {
			util.LogAndPanic(err)
		}
		util.PrintAndLog(fmt.Sprintf("created test-blob%d", i))
	}

	list, err := ListBlobs(ctx, accountName, accountGroupName, containerName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog(fmt.Sprintf("listed %d blobs", len(list.Segment.BlobItems)))

	// Output:
	// created container
	// created test-blob0
	// created test-blob1
	// created test-blob2
	// listed 3 blobs
}
