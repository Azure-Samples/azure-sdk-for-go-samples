// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package resources

import (
	"context"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
)

func init() {
	err := iam.ParseArgs()
	if err != nil {
		log.Fatalln("failed to parse IAM args")
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

func ExampleCreateGroupWithAuthFile() {
	helpers.SetResourceGroupName("CreateGroupWithAuthFile")
	defer Cleanup(context.Background())

	_, err := CreateGroupWithAuthFile(context.Background(), helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("resource group created, authentication was set up with an Azure CLI auth file")

	// Output:
	// resource group created, authentication was set up with an Azure CLI auth file
}
