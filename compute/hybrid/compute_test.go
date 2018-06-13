// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package compute

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	hybridnetwork "github.com/Azure-Samples/azure-sdk-for-go-samples/network/hybrid"
	hybridresources "github.com/Azure-Samples/azure-sdk-for-go-samples/resources/hybrid"
	hybridstorage "github.com/Azure-Samples/azure-sdk-for-go-samples/storage/hybrid"

	"github.com/marstr/randname"
)

var (
	vmName           = randname.GenerateWithPrefix("az-samples-go-", 10)
	nicName          = "nic1"
	username         = "az-samples-go-user"
	password         = "NoSoupForYou1!"
	sshPublicKeyPath = os.Getenv("HOME") + "/.ssh/id_rsa.pub"

	virtualNetworkName = "vnet1"
	subnetName         = "subnet1"
	nsgName            = "nsg1"
	ipName             = "ip1"
	storageAccountName = randname.Prefixed{Prefix: "storageaccount", Acceptable: randname.LowercaseAlphabet, Len: 10}.Generate()
)

func TestMain(m *testing.M) {
	err := iam.ParseArgs()
	if err != nil {
		log.Fatalln("failed to parse IAM args")
	}

	os.Exit(m.Run())
}

func ExampleCreateVM() {
	ctx := context.Background()
	defer hybridresources.Cleanup(ctx)
	_, err := hybridresources.CreateGroup(ctx)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	_, err = hybridnetwork.CreateVirtualNetworkAndSubnets(ctx, virtualNetworkName, subnetName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created vnet and a subnet")

	_, err = hybridnetwork.CreateNetworkSecurityGroup(ctx, nsgName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created network security group")

	_, err = hybridnetwork.CreatePublicIP(ctx, ipName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created public IP")

	_, err = hybridnetwork.CreateNetworkInterface(ctx, nicName, nsgName, virtualNetworkName, subnetName, ipName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created nic")

	_, err = hybridstorage.CreateStorageAccount(ctx, storageAccountName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created storage account")

	_, err = CreateVM(ctx, vmName, nicName, username, password, storageAccountName, sshPublicKeyPath)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created VM")

	// Output:
	// created vnet and a subnet
	// created network security group
	// created public IP
	// created nic
	// created storage account
	// created VM
}
