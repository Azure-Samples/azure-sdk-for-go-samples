// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package hybridnetwork

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/subosito/gotenv"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	hybridresources "github.com/Azure-Samples/azure-sdk-for-go-samples/resources/hybrid"
)

var (
	virtualNetworkName   = "vnet1"
	subnetName           = "subnet1"
	nsgName              = "nsg1"
	nicName              = "nic1"
	ipName               = "ip1"
	networkInterfaceName = "netinterface1"
)

func parseArgs() error {
	gotenv.Load()

	virtualNetworkName = os.Getenv("AZ_VNET_NAME")
	flag.StringVar(&virtualNetworkName, "vnetName", virtualNetworkName, "Specify a name for the vnet.")

	err := helpers.ParseArgs()
	if err != nil {
		log.Fatalln("failed to parse args")
	}

	if !(len(virtualNetworkName) > 0) {
		virtualNetworkName = "vnet1"
	}

	return nil
}

func TestMain(m *testing.M) {
	err := helpers.ParseArgs()
	if err != nil {
		log.Fatalln("failed to parse args")
	}

	ctx := context.Background()
	defer hybridresources.Cleanup(ctx)

	_, err = hybridresources.CreateGroup(ctx)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog(fmt.Sprintf("resource group created on location: %s", helpers.Location()))

	os.Exit(m.Run())
}

func ExampleCreateVirtualNetworkAndSubnets() {
	_, err := CreateVirtualNetworkAndSubnets(context.Background(), virtualNetworkName, subnetName)
	if err != nil {
		log.Fatal(fmt.Sprintf("Cannot create virtual network. Error details: %s", err.Error()))
	}
	fmt.Println("VNET created")

	// Output:
	// VNET created
}

func ExampleCreateNetworkSecurityGroup() {
	_, err := CreateNetworkSecurityGroup(context.Background(), nsgName)
	if err != nil {
		log.Fatal(fmt.Sprintf("Cannot create network security group. Error details: %s", err.Error()))
	}
	fmt.Println("VNET security group created")

	// Output:
	// VNET security group created
}

func ExampleCreatePublicIP() {
	_, err := CreatePublicIP(context.Background(), ipName)
	if err != nil {
		log.Fatal(fmt.Sprintf("Cannot create public IP. Error details: %s", err.Error()))
	}
	fmt.Println("Public IP created")

	// Output:
	// Public IP created
}

func ExampleCreateNetworkInterface() {
	_, err := CreateNetworkInterface(context.Background(), networkInterfaceName, nsgName, virtualNetworkName, subnetName, ipName)
	if err != nil {
		log.Fatal(fmt.Sprintf("Cannot create network interface. Error details: %s", err.Error()))
	}
	fmt.Println("Network interface created")

	// Output:
	// Network interface created
}
