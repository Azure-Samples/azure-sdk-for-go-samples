// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package compute

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/graphrbac"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/keyvault"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/network"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/storage"
	"github.com/subosito/gotenv"
)

var (
	vmName           = "az-samples-go-" + helpers.GetRandomLetterSequence(10)
	accountName      = strings.ToLower("azuresamplesgo" + helpers.GetRandomLetterSequence(10))
	vaultName        = "az-samples-go-" + helpers.GetRandomLetterSequence(10)
	diskName         = "az-samples-go-" + helpers.GetRandomLetterSequence(10)
	nicName          = "nic1"
	username         = "az-samples-go-user"
	password         = "NoSoupForYou1!"
	sshPublicKeyPath = os.Getenv("HOME") + "/.ssh/id_rsa.pub"

	virtualNetworkName = "vnet1"
	subnet1Name        = "subnet1"
	subnet2Name        = "subnet2"
	nsgName            = "nsg1"
	ipName             = "ip1"
)

func TestMain(m *testing.M) {
	err := parseArgs()
	if err != nil {
		log.Fatalln("failed to parse args")
	}

	ctx := context.Background()
	defer resources.Cleanup(ctx)

	_, err = resources.CreateGroup(ctx, helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog(fmt.Sprintf("resource group created on location: %s", helpers.Location()))

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
	ctx := context.Background()

	_, err := network.CreateVirtualNetworkAndSubnets(ctx, virtualNetworkName, subnet1Name, subnet2Name)
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

	// Output:
	// created vnet and 2 subnets
	// created network security group
	// created public IP
	// created nic
	// created VM
}

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
