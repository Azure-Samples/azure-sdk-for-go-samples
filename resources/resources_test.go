// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package resources

import (
	"context"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
)

func init() {
	err := helpers.ParseArgs()
	if err != nil {
		log.Fatalf("cannot parse arguments: %v", err)
	}
}

func ExampleCreateGroup() {
	helpers.SetResourceGroupName("CreateGroup")
	defer Cleanup(context.Background())

	_, err := CreateGroup(context.Background(), helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("resource group created")

	// Output:
	// resource group created
}
