// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cosmosdb

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

func ExampleGetCosmosDbAccountKeys() {
	context := context.Background()

	helpers.SetResourceGroupName(resourceGroupNameSuffix)

	defer resources.Cleanup(context)

	_, err := resources.CreateGroup(context, helpers.ResourceGroupName())

	if err != nil {
		helpers.PrintAndLog("Failed to create resource group.")
		helpers.PrintAndLog(err.Error())

		return
	}

	_, err = CreateCosmosDbAccount(context, cosmosDbAccountName)

	if err != nil {
		helpers.PrintAndLog("Failed to create CosmosDB Account.")
		helpers.PrintAndLog(err.Error())

		return
	}

	_, err = GetCosmosDbAccountKeys(context, cosmosDbAccountName)

	if err != nil {
		helpers.PrintAndLog("Failed to get CosmosDB Account keys.")
		helpers.PrintAndLog(err.Error())

		return
	}

	helpers.PrintAndLog("Got CosmosDB Account keys.")

	// Output:
	// Got CosmosDB Account keys.
}
