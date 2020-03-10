// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package compute

import (
	"context"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

func ExampleCreateContainerGroup() {
	var groupName = config.GenerateGroupName("CreateContainerGroup")
	config.SetGroupName(groupName)

	ctx := context.Background()
	defer resources.Cleanup(ctx)

	_, err := resources.CreateGroup(ctx, groupName)
	if err != nil {
		util.LogAndPanic(err)
	}

	_, err = CreateContainerGroup(ctx, containerGroupName, config.Location(), groupName)
	if err != nil {
		log.Fatalf("cannot create container group: %v", err)
	}
	util.PrintAndLog("created container group")

	c, err := GetContainerGroup(ctx, groupName, containerGroupName)
	if err != nil {
		log.Fatalf("cannot get container group %v from resource group %v", containerGroupName, groupName)
	}

	if *c.Name != containerGroupName {
		log.Fatalf("incorrect name of container group: expected %v, got %v", containerGroupName, *c.Name)
	}
	util.PrintAndLog("retrieved container group")

	_, err = UpdateContainerGroup(ctx, groupName, containerGroupName)
	if err != nil {
		log.Fatalf("cannot upate container group: %v", err)
	}
	util.PrintAndLog("updated container group")

	_, err = DeleteContainerGroup(ctx, groupName, containerGroupName)
	if err != nil {
		log.Fatalf("cannot delete container group %v from resource group %v: %v", containerGroupName, groupName, err)
	}
	util.PrintAndLog("deleted container group")

	// Output:
	// created container group
	// retrieved container group
	// updated container group
	// deleted container group
}
