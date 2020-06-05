// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package eventhubs

import (
	"context"
	"flag"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

const (
	nsName  = "goehtestns"
	hubName = "goehtesthub"

	// for storage.LeaserCheckpointer
	storageAccountName   = "goehteststorage"
	storageContainerName = "goeventhubsleasercheckpointer"
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

func Example_eventHubs() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	// defer goes in LIFO order
	defer cancel()
	defer resources.Cleanup(context.Background()) // cleanup can take a long time

	// create group
	var err error
	var groupName = config.GenerateGroupName("EventHubs")
	config.SetGroupName(groupName)

	_, err = resources.CreateGroup(ctx, config.GroupName())
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created group")

	// create Event Hubs namespace
	_, err = CreateNamespace(ctx, nsName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created namespace")

	// create Event Hubs hub
	_, err = CreateHub(ctx, nsName, hubName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created hub")

	// send and receive messages
	log.Printf("Send(ctx)\n")
	Send(ctx, nsName, hubName)
	log.Printf("Receive(ctx)\n")
	Receive(ctx, nsName, hubName)
	log.Printf("ReceiveViaEPH(ctx)\n")
	ReceiveViaEPH(ctx, nsName, hubName, storageAccountName, storageContainerName)

	// Output:
	// created group
	// created namespace
	// created hub
	// received: test-message
	// received: test-message
}
