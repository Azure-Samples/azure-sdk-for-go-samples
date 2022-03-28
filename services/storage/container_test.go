// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/internal/util"
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
