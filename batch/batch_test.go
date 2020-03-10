// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.
package batch

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
	"github.com/marstr/randname"
)

var (
	accountName = strings.ToLower(randname.GenerateWithPrefix("gosdkbatch", 5))
	jobID       = randname.GenerateWithPrefix("gosdk-batch-j-", 5)
	poolID      = randname.GenerateWithPrefix("gosdk-batch-p-", 5)
)

// TestMain sets up the environment and initiates tests.
func TestMain(m *testing.M) {
	var err error
	err = config.ParseEnvironment()
	if err != nil {
		log.Fatalf("failed to parse env: %+v", err)
	}
	err = config.AddFlags()
	if err != nil {
		log.Fatalf("failed to parse env: %+v", err)
	}
	flag.Parse()

	code := m.Run()
	os.Exit(code)
}

func ExampleCreateAzureBatchAccount() {
	var groupName = config.GenerateGroupName("Batch")
	config.SetGroupName(groupName)
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*30))
	defer cancel()
	defer resources.Cleanup(ctx)

	_, err := resources.CreateGroup(ctx, config.GroupName())
	if err != nil {
		util.LogAndPanic(err)
	}

	_, err = CreateAzureBatchAccount(ctx, accountName, config.Location(), config.GroupName())
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created batch account")

	err = CreateBatchPool(ctx, accountName, config.Location(), poolID)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created batch pool")

	err = CreateBatchJob(ctx, accountName, config.Location(), poolID, jobID)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created batch job")

	taskID, err := CreateBatchTask(ctx, accountName, config.Location(), jobID)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created batch task")

	taskOutput, err := WaitForTaskResult(ctx, accountName, config.Location(), jobID, taskID)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("output from task:")
	util.PrintAndLog(taskOutput)

	// Output:
	// created batch account
	// created batch pool
	// created batch job
	// created batch task
	// output from task:
	// Hello world from the Batch Hello world sample!
}
