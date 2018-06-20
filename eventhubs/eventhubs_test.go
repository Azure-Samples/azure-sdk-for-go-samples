// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package eventhubs

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

const (
	location = "westus2"
	nsName   = "ehtest-03-ns"
	hubName  = "ehtest-03-hub"

	// for storage.LeaserCheckpointer
	storageAccountName   = "ehtest0001storage"
	storageContainerName = "eventhubs0001leasercheckpointer"
)

func TestMain(m *testing.M) {
	err := helpers.ParseArgs()
	if err != nil {
		log.Fatalf("failed to parse args: %v\n", err)
	}

	err = iam.ParseArgs()
	if err != nil {
		log.Fatalln("failed to parse IAM args")
	}

	os.Exit(m.Run())
}

func ExampleEventHubs() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	// defer goes in LIFO order
	defer cancel()
	defer resources.Cleanup(context.Background()) // cleanup can take a long time

	// create group
	var err error
	helpers.SetResourceGroupName("eventhubstest")
	_, err = resources.CreateGroup(ctx, helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created group")

	// create Event Hubs namespace
	_, err = CreateNamespace(ctx, nsName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created namespace")

	// create Event Hubs hub
	_, err = CreateHub(ctx, nsName, hubName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created hub")

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
