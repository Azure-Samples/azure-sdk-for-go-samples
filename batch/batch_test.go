// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.
package batch

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
	"github.com/marstr/randname"
)

var (
	accountName string
	jobID       string
	poolID      string
)

func TestMain(m *testing.M) {
	err := parseArgs()
	if err != nil {
		log.Fatalf("failed to parse batch args: %v\n", err)
	}

	err = iam.ParseArgs()
	if err != nil {
		log.Fatalln("failed to parse IAM args")
	}

	os.Exit(m.Run())
}

func parseArgs() error {
	namer := randname.Prefixed{
		Prefix:     "b",
		Len:        10,
		Acceptable: randname.LowercaseAlphabet,
	}
	accountName = os.Getenv("AZURE_BATCH_NAME")
	if !(len(accountName) > 0) {
		accountName = namer.Generate()
	}

	namer.Prefix = "j"
	jobID = namer.Generate()
	namer.Prefix = "p"
	poolID = namer.Generate()

	return nil
}

func ExampleCreateAzureBatchAccount() {
	helpers.SetResourceGroupName("CreateBatch")
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*30))
	defer cancel()
	defer resources.Cleanup(ctx)
	_, err := resources.CreateGroup(ctx, helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}

	_, err = CreateAzureBatchAccount(ctx, accountName, helpers.Location(), helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
		return
	}

	helpers.PrintAndLog("created batch account")

	err = CreateBatchPool(ctx, accountName, helpers.Location(), poolID)
	if err != nil {
		helpers.PrintAndLog(err.Error())
		return
	}

	helpers.PrintAndLog("created batch pool")

	err = CreateBatchJob(ctx, accountName, helpers.Location(), poolID, jobID)
	if err != nil {
		helpers.PrintAndLog(err.Error())
		return
	}

	helpers.PrintAndLog("created batch job")

	taskID, err := CreateBatchTask(ctx, accountName, helpers.Location(), jobID)
	if err != nil {
		helpers.PrintAndLog(err.Error())
		return
	}

	helpers.PrintAndLog("created batch task")

	taskOutput, err := WaitForTaskResult(ctx, accountName, helpers.Location(), jobID, taskID)
	if err != nil {
		helpers.PrintAndLog(err.Error())
		return
	}

	helpers.PrintAndLog("output from task:")
	helpers.PrintAndLog(taskOutput)

	// Output:
	// created batch account
	// created batch pool
	// created batch job
	// created batch task
	// output from task:
	// Hello world from the Batch Hello world sample!
}
