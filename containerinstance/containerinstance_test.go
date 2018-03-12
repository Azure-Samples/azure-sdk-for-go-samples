// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package containerinstance

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

var (
	containerGroupName string
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

	containerGroupName = os.Getenv("AZ_CONTAINERINSTANCE_CONTAINER_GROUP_NAME")
	if !(len(containerGroupName) > 0) {
		containerGroupName = "az-samples-go-container-group-" + internal.GetRandomLetterSequence(10)
	}

	// Container instance is not yet available in many Azure locations
	internal.OverrideLocation([]string{
		"westus",
		"eastus",
		"westeurope",
	})
	return nil
}

func ExampleCreateContainerGroup() {
	internal.SetResourceGroupName("CreateContainerGroup")
	ctx := context.Background()
	defer resources.Cleanup(ctx)
	_, err := resources.CreateGroup(ctx, internal.ResourceGroupName())
	if err != nil {
		internal.PrintAndLog(err.Error())
	}

	_, err = CreateContainerGroup(ctx, containerGroupName, internal.Location(), internal.ResourceGroupName())
	if err != nil {
		log.Fatalf("cannot create container group: %v", err)
	}

	internal.PrintAndLog("created container group")

	c, err := GetContainerGroup(ctx, internal.ResourceGroupName(), containerGroupName)
	if err != nil {
		log.Fatalf("cannot get container group %v from resource group %v", containerGroupName, internal.ResourceGroupName())
	}

	if *c.Name != containerGroupName {
		log.Fatalf("incorrect name of container group: expected %v, got %v", containerGroupName, *c.Name)
	}

	internal.PrintAndLog("retrieved container group")

	_, err = UpdateContainerGroup(ctx, internal.ResourceGroupName(), containerGroupName)
	if err != nil {
		log.Fatalf("cannot upate container group: %v", err)
	}

	internal.PrintAndLog("updated container group")

	_, err = DeleteContainerGroup(ctx, internal.ResourceGroupName(), containerGroupName)
	if err != nil {
		log.Fatalf("cannot delete container group %v from resource group %v: %v", containerGroupName, internal.ResourceGroupName(), err)
	}

	internal.PrintAndLog("deleted container group")

	// Output:
	// created container group
	// retrieved container group
	// updated container group
	// deleted container group
}
