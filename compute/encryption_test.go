// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package compute

import (
	"context"
	"io/ioutil"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/graphrbac"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/keyvault"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/network"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/storage"
)

func ExampleCreateVMWithEncryptedManagedDisks() {
	ctx := context.Background()

	_, err := storage.CreateStorageAccount(ctx, accountName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created storage account")

	_, err = network.CreateVirtualNetwork(ctx, virtualNetworkName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created vnet")

	_, err = network.CreateVirtualNetworkSubnet(ctx, virtualNetworkName, subnet1Name)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created subnet")

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

	_, err = network.CreateNetworkSecurityGroup(ctx, nsgName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created network security group")

	_, err = network.CreatePublicIP(ctx, ipName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created public IP")

	_, err = network.CreateNIC(ctx, virtualNetworkName, subnet1Name, nsgName, ipName, nicName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created nic")

	_, err = CreateVMWithManagedDisk(ctx, nicName, diskName, accountName, vmName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created virtual machine")

	key, err := keyvault.CreateKeyBundle(ctx, vaultName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
		b, err := ioutil.ReadAll(key.Body)
		if err != nil {
			helpers.PrintAndLog(err.Error())
		}
		helpers.PrintAndLog(string(b))

	}
	helpers.PrintAndLog("created key bundle")

	_, err = AddEncyptionExtension(ctx, vmName, vaultName, *key.Key.Kid)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("added vm encryption extension")

	// Output:
	// created storage account
	// created vnet
	// created subnet
	// created keyvault
	// created disk
	// created network security group
	// created public IP
	// created nic
	// created virtual machine
	// created key bundle
	// added vm encryption extension
}
