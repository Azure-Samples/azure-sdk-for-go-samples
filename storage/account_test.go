// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package storage

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

func ExampleStorageAccountOperations() {
	accountName = getAccountName()

	helpers.SetResourceGroupName("StorageAccountOperations")
	ctx := context.Background()
	defer resources.Cleanup(ctx)
	_, err := resources.CreateGroup(ctx, helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}

	_, err = resources.RegisterProvider(ctx, "Microsoft.Storage")
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("registered resource provider")

	_, err = CheckAccountAvailability(ctx, accountName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("checked for account availability")

	_, err = CreateStorageAccount(ctx, accountName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created storage account")

	_, err = GetStorageAccount(ctx, accountName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("got storage account details")

	_, err = UpdateAccount(ctx, accountName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("updated storage account")

	_, err = ListAccountsByResourceGroup(ctx)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("listed storage accounts in resource group")

	_, err = ListAccountsBySubscription(ctx)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("listed storage accounts in subscription")

	_, err = GetAccountKeys(ctx, accountName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("get storage account keys")

	_, err = RegenerateAccountKey(ctx, accountName, 0)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("regenerated first storage account key")

	_, err = ListUsage(ctx)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("listed usage")

	// Output:
	// registered resource provider
	// checked for account availability
	// created storage account
	// got storage account details
	// updated storage account
	// listed storage accounts in resource group
	// listed storage accounts in subscription
	// get storage account keys
	// regenerated first storage account key
	// listed usage
}
