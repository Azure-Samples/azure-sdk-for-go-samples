// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package compute

import (
	"context"
	"flag"
	"log"
	"os"
	"testing"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
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
	if err := config.ParseEnvironment(); err != nil {
		log.Fatalf("failed to parse env: %+v", err)
	}
	if err := config.AddFlags(); err != nil {
		log.Fatalf("failed to add flags: %+v", err)
	}
	flag.Parse()
	os.Exit(m.Run())
}

func ExampleCreateVM() {
	var groupName = config.GenerateGroupName("HybridVM")
	config.SetGroupName(groupName)

	ctx := context.Background()
	defer hybridresources.Cleanup(ctx)
	_, err := hybridresources.CreateGroup(ctx)
	if err != nil {
		util.LogAndPanic(err)
	}
	_, err = hybridnetwork.CreateVirtualNetworkAndSubnets(ctx, virtualNetworkName, subnetName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created vnet and a subnet")

	_, err = hybridnetwork.CreateNetworkSecurityGroup(ctx, nsgName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created network security group")

	_, err = hybridnetwork.CreatePublicIP(ctx, ipName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created public IP")

	_, err = hybridnetwork.CreateNetworkInterface(ctx, nicName, nsgName, virtualNetworkName, subnetName, ipName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created nic")

	_, err = hybridstorage.CreateStorageAccount(ctx, storageAccountName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created storage account")

	_, err = CreateVM(ctx, vmName, nicName, username, password, storageAccountName, sshPublicKeyPath)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created VM")

	// Output:
	// created vnet and a subnet
	// created network security group
	// created public IP
	// created nic
	// created storage account
	// created VM
}
