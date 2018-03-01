// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package compute

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/network"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/subosito/gotenv"
)

var (
	vmName           = "az-samples-go-" + helpers.GetRandomLetterSequence(10)
	diskName         = "az-samples-go-" + helpers.GetRandomLetterSequence(10)
	nicName          = "nic" + helpers.GetRandomLetterSequence(10)
	username         = "az-samples-go-user"
	password         = "NoSoupForYou1!"
	sshPublicKeyPath = os.Getenv("HOME") + "/.ssh/id_rsa.pub"

	virtualNetworkName = "vnet1"
	subnet1Name        = "subnet1"
	subnet2Name        = "subnet2"
	nsgName            = "nsg1"
	ipName             = "ip1"
	lbName             = "lb"
)

func TestMain(m *testing.M) {
	err := parseArgs()
	if err != nil {
		log.Fatalln("failed to parse args")
	}

	os.Exit(m.Run())
}

func parseArgs() error {
	gotenv.Load()

	virtualNetworkName = os.Getenv("AZ_VNET_NAME")
	flag.StringVar(&virtualNetworkName, "vnetName", virtualNetworkName, "Specify a name for the vnet.")

	err := helpers.ParseArgs()
	if err != nil {
		return fmt.Errorf("cannot parse args: %v", err)
	}

	if !(len(virtualNetworkName) > 0) {
		virtualNetworkName = "vnet1"
	}

	return nil
}

func ExampleCreateVM() {
	helpers.SetResourceGroupName("CreateVM")
	ctx := context.Background()
	defer resources.Cleanup(ctx)
	_, err := resources.CreateGroup(ctx, helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}

	_, err = network.CreateVirtualNetworkAndSubnets(ctx, virtualNetworkName, subnet1Name, subnet2Name)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created vnet and 2 subnets")

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

	_, err = CreateVM(ctx, vmName, nicName, username, password, sshPublicKeyPath)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created VM")

	// Now that the Vm has been created, we can do some simple operations on the VM

	_, err = UpdateVM(ctx, vmName, map[string]*string{
		"who rocks": to.StringPtr("golang"),
		"where":     to.StringPtr("on azure"),
	})
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("updated VM")

	_, err = AttachDataDisks(ctx, vmName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("attached data disks")

	_, err = DetachDataDisks(ctx, vmName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("detached data disks")

	_, err = UpdateOSDiskSize(ctx, vmName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("updated OS disk size")

	_, err = StartVM(ctx, vmName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("started VM")

	_, err = RestartVM(ctx, vmName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("restarted VM")

	_, err = PowerOffVM(ctx, vmName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("stopped VM")

	// Output:
	// created vnet and 2 subnets
	// created network security group
	// created public IP
	// created nic
	// created VM
	// updated VM
	// attached data disks
	// detached data disks
	// updated OS disk size
	// started VM
	// restarted VM
	// stopped VM
}
