// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package hdinsight

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/storage"
)

// TestMain sets up the environment and initiates tests.
func TestMain(m *testing.M) {
	var err error
	err = config.ParseEnvironment()
	if err != nil {
		log.Fatalf("failed to parse env: %v\n", err)
	}
	err = config.AddFlags()
	if err != nil {
		log.Fatalf("failed to parse env: %v\n", err)
	}
	flag.Parse()

	code := m.Run()
	os.Exit(code)
}

func ExampleCreateHadoopCluster() {
	rgName := config.GenerateGroupName("HadoopClusterExample")
	config.SetGroupName(rgName)

	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()
	defer resources.Cleanup(ctx)

	_, err := resources.CreateGroup(ctx, rgName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created resource group")

	storageAccountName := strings.ToLower(config.AppendRandomSuffix("exampleforhadoop"))
	sa, err := storage.CreateStorageAccount(context.Background(), storageAccountName, rgName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created storage account")

	containerName := strings.ToLower(config.AppendRandomSuffix("hadoopfilesystem"))
	_, err = storage.CreateContainer(context.Background(), storageAccountName, rgName, containerName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created container")

	keys, err := storage.GetAccountKeys(context.Background(), storageAccountName, rgName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved storage account keys")

	clusterName := strings.ToLower(config.AppendRandomSuffix("exhadoop36cluster"))
	_, err = CreateHadoopCluster(rgName, clusterName, StorageAccountInfo{
		Name:      fmt.Sprintf("%s.blob.core.windows.net", *sa.Name), // TODO: can we get the full URL from the service?
		Container: containerName,
		Key:       *(*keys.Keys)[0].Value,
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created cluster")

	// Output:
	// created resource group
	// created storage account
	// created container
	// retrieved storage account keys
	// creating hadoop cluster
	// waiting for hadoop cluster to finish deploying, this will take a while...
	// created cluster
}
