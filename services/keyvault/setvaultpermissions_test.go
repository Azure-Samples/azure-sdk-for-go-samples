// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package keyvault

import (
	"context"

	"github.com/marstr/randname"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/internal/util"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/resources"
)

var (
	kvName  = randname.GenerateWithPrefix("vault-sample-go-", 5)
	keyName = randname.GenerateWithPrefix("key-sample-go-", 5)
)

func ExampleSetVaultPermissions() {
	var groupName = config.GenerateGroupName("KeyVault")
	config.SetGroupName(groupName)

	ctx := context.Background()
	defer resources.Cleanup(ctx)

	_, err := resources.CreateGroup(ctx, config.GroupName())
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("resource group created")

	_, err = CreateVault(ctx, kvName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("vault created")

	_, err = SetVaultPermissions(ctx, kvName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("set vault permissions")

	_, err = CreateKey(ctx, kvName, keyName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created key")

	// Output:
	// resource group created
	// vault created
	// set vault permissions
	// created key
}
