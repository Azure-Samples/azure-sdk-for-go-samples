// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package storage

import (
	"context"
	"log"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

func Example_storageAccountOperations() {
	var groupName = testAccountGroupName
	var accountName = testAccountName

	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()
	// don't cleanup yet so dataplane tests can use account
	// defer resources.Cleanup(ctx)

	_, err := resources.CreateGroup(ctx, groupName)
	if err != nil {
		util.LogAndPanic(err)
	}

	_, err = resources.RegisterProvider(ctx, "Microsoft.Storage")
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("registered resource provider")

	result, err := CheckAccountNameAvailability(ctx, accountName)
	log.Printf("[%T]: %+v\n", result, result)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("checked for account availability")

	var errOuter, errInner error // see next comment
	_, errOuter = CreateStorageAccount(ctx, accountName, groupName)
	if errOuter != nil {
		// this could be because we've already created it, and a way to check
		// that is to try to get it, so that's what we do here. if we can get
		// it we pretend we created it so tests can proceed. if we can't get
		// it, we don't want to confuse the tester and return the error for the
		// Get, so we return the orignial error from the Create.
		_, errInner = GetStorageAccount(ctx, accountName, groupName)
		if errInner != nil {
			util.PrintAndLog(errOuter.Error())
		}
	}
	util.PrintAndLog("created storage account")

	_, err = GetStorageAccount(ctx, accountName, groupName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("got storage account details")

	_, err = UpdateAccount(ctx, accountName, groupName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("updated storage account")

	_, err = ListAccountsByResourceGroup(ctx, groupName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("listed storage accounts in resource group")

	_, err = ListAccountsBySubscription(ctx)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("listed storage accounts in subscription")

	_, err = GetAccountKeys(ctx, accountName, groupName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("get storage account keys")

	_, err = RegenerateAccountKey(ctx, accountName, groupName, 1)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("regenerated second storage account key")

	_, err = ListUsage(ctx, config.Location())
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("listed usage")

	// Output:
	// registered resource provider
	// checked for account availability
	// created storage account
	// got storage account details
	// updated storage account
	// listed storage accounts in resource group
	// listed storage accounts in subscription
	// get storage account keys
	// regenerated second storage account key
	// listed usage
}
