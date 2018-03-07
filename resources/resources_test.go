// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package resources

import (
	"context"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal"
)

func init() {
	err := internal.ParseArgs()
	if err != nil {
		log.Fatalf("cannot parse arguments: %v", err)
	}
}

func ExampleCreateGroup() {
	internal.SetResourceGroupName("CreateGroup")
	defer Cleanup(context.Background())

	_, err := CreateGroup(context.Background(), internal.ResourceGroupName())
	if err != nil {
		internal.PrintAndLog(err.Error())
	}
	internal.PrintAndLog("resource group created")

	// Output:
	// resource group created
}

func ExampleCreateGroupWithAuthFile() {
	internal.SetResourceGroupName("CreateGroupWithAuthFile")
	defer Cleanup(context.Background())

	_, err := CreateGroupWithAuthFile(context.Background(), internal.ResourceGroupName())
	if err != nil {
		internal.PrintAndLog(err.Error())
	}
	internal.PrintAndLog("resource group created, authentication was set up with an Azure CLI auth file")

	// Output:
	// resource group created, authentication was set up with an Azure CLI auth file
}
