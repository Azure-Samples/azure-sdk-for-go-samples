// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.
package batch

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

var (
	accountName string
	jobID       string
	poolID      string
)

func TestMain(m *testing.M) {
	err := parseArgs()
	if err != nil {
		log.Fatalln("failed to parse args")
	}
	os.Exit(m.Run())
}

func parseArgs() error {
	err := internal.ParseArgs()
	if err != nil {
		return fmt.Errorf("cannot parse args: %v", err)
	}

	accountName = os.Getenv("AZ_BATCH_NAME")
	if !(len(accountName) > 0) {
		accountName = strings.ToLower("b" + internal.GetRandomLetterSequence(10))
	}

	jobID = strings.ToLower("j" + internal.GetRandomLetterSequence(10))
	poolID = strings.ToLower("p" + internal.GetRandomLetterSequence(10))

	return nil
}

func ExampleCreateAzureBatchAccount() {
	internal.SetResourceGroupName("CreateBatch")
	ctx := context.Background()
	defer resources.Cleanup(ctx)
	_, err := resources.CreateGroup(ctx, internal.ResourceGroupName())
	if err != nil {
		internal.PrintAndLog(err.Error())
	}

	_, err = CreateAzureBatchAccount(ctx, accountName, internal.Location(), internal.ResourceGroupName())
	if err != nil {
		internal.PrintAndLog(err.Error())
		return
	}

	internal.PrintAndLog("created batch account")

	err = CreateBatchPool(ctx, accountName, internal.Location(), poolID)
	if err != nil {
		internal.PrintAndLog(err.Error())
		return
	}

	internal.PrintAndLog("created batch pool")

	err = CreateBatchJob(ctx, accountName, internal.Location(), poolID, jobID)
	if err != nil {
		internal.PrintAndLog(err.Error())
		return
	}

	internal.PrintAndLog("created batch job")

	taskID, err := CreateBatchTask(ctx, accountName, internal.Location(), jobID)
	if err != nil {
		internal.PrintAndLog(err.Error())
		return
	}

	internal.PrintAndLog("created batch task")

	taskOutput, err := WaitForTaskResult(ctx, accountName, internal.Location(), jobID, taskID)
	if err != nil {
		internal.PrintAndLog(err.Error())
		return
	}

	internal.PrintAndLog("output from task:")
	internal.PrintAndLog(taskOutput)

	// Output:
	// created batch account
	// created batch pool
	// created batch job
	// created batch task
	// output from task:
	// Hello world from the Batch Hello world sample!
}
