// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package compute

import (
	"context"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/graphrbac"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/keyvault"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/network"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

func ExampleCreateVMWithEncryptedManagedDisks() {
	helpers.SetResourceGroupName("CreateVMEncryptedDisks")
	vaultName := "az-samples-go-" + helpers.GetRandomLetterSequence(10)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*20))
	defer cancel()
	defer resources.Cleanup(ctx)
	_, err := resources.CreateGroup(ctx, helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}

	_, err = network.CreateVirtualNetworkAndSubnets(ctx, virtualNetworkName, subnet1Name, subnet2Name)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created vnet and subnets")

	// If authenticating as a user, also add the user to the keyvault access policies
	userID := ""
	if iam.AuthGrantType() == iam.OAuthGrantTypeDeviceFlow {
		cu, err := graphrbac.GetCurrentUser(ctx)
		if err != nil {
			helpers.PrintAndLog(err.Error())
		}
		userID = *cu.ObjectID
	}

	_, err = keyvault.CreateComplexKeyVault(ctx, vaultName, userID)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created keyvault")

	_, err = CreateManagedDisk(ctx, diskName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created disk")

	_, err = network.CreatePublicIP(ctx, ipName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created public IP")

	_, err = network.CreateNIC(ctx, virtualNetworkName, subnet1Name, "", ipName, nicName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created nic")

	_, err = CreateVMWithManagedDisk(ctx, nicName, diskName, vmName, username, password)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created virtual machine")

	key, err := keyvault.CreateKeyBundle(ctx, vaultName)
	if err != nil {
		helpers.PrintAndLog(err.Error())

	}
	helpers.PrintAndLog("created key bundle")

	_, err = AddEncyptionExtension(ctx, vmName, vaultName, *key.Key.Kid)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("added vm encryption extension")

	// Output:
	// created vnet and subnets
	// created keyvault
	// created disk
	// created public IP
	// created nic
	// created virtual machine
	// created key bundle
	// added vm encryption extension
}
