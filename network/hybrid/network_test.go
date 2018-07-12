// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package network

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	hybridresources "github.com/Azure-Samples/azure-sdk-for-go-samples/resources/hybrid"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
)

var (
	virtualNetworkName   = "vnet1"
	subnetName           = "subnet1"
	nsgName              = "nsg1"
	nicName              = "nic1"
	ipName               = "ip1"
	networkInterfaceName = "netinterface1"
)

func setupEnvironment() error {
	err1 := config.ParseEnvironment()
	err2 := config.AddFlags()
	err3 := addLocalConfig()

	for _, err := range []error{err1, err2, err3} {
		if err != nil {
			return err
		}
	}

	flag.Parse()
	return nil
}

func addLocalConfig() error {
	vnetNameFromEnv := os.Getenv("AZURE_VNET_NAME")
	if len(vnetNameFromEnv) > 0 {
		virtualNetworkName = vnetNameFromEnv
	}
	flag.StringVar(&virtualNetworkName, "vnetName", virtualNetworkName, "Name for the VNET.")
	return nil
}

func TestNetwork(t *testing.T) {
	err := setupEnvironment()
	if err != nil {
		t.Fatalf("could not set up environment: %v\n", err)
	}

	groupName := config.GenerateGroupName("network-test")
	config.SetGroupName(groupName) // TODO: don't use globals
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()
	defer hybridresources.Cleanup(ctx)

	_, err = hybridresources.CreateGroup(ctx)
	if err != nil {
		t.Fatalf("could not create group %v\n", err.Error())
	}
	_, err = CreateVirtualNetworkAndSubnets(context.Background(), virtualNetworkName, subnetName)
	if err != nil {
		t.Fatalf("could not create vnet: %v\n", err.Error())
	}
	t.Logf("created vnet")
}

func ExampleCreateNetworkSecurityGroup() {
	groupName := config.GenerateGroupName("CreateNetworkSecurityGroup")
	config.SetGroupName(groupName) // TODO: don't use globals
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()
	defer hybridresources.Cleanup(ctx)

	_, err := hybridresources.CreateGroup(ctx)
	if err != nil {
		util.PrintAndLog(err.Error())
	}
	_, err = CreateNetworkSecurityGroup(ctx, nsgName)
	if err != nil {
		log.Fatal(fmt.Sprintf("Cannot create network security group. Error details: %s", err.Error()))
	}
	fmt.Println("VNET security group created")

	// Output:
	// VNET security group created
}

func ExampleCreatePublicIP() {
	groupName := config.GenerateGroupName("CreatePublicIP")
	config.SetGroupName(groupName) // TODO: don't use globals
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()
	defer hybridresources.Cleanup(ctx)

	_, err := hybridresources.CreateGroup(ctx)
	if err != nil {
		util.PrintAndLog(err.Error())
	}
	_, err = CreatePublicIP(ctx, ipName)
	if err != nil {
		log.Fatal(fmt.Sprintf("Cannot create public IP. Error details: %s", err.Error()))
	}
	fmt.Println("Public IP created")

	// Output:
	// Public IP created
}

func ExampleCreateNetworkInterface() {
	groupName := config.GenerateGroupName("CreateNIC")
	config.SetGroupName(groupName) // TODO: don't use globals
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()
	defer hybridresources.Cleanup(ctx)

	_, err := hybridresources.CreateGroup(ctx)
	if err != nil {
		util.PrintAndLog(err.Error())
	}

	_, err = CreateNetworkInterface(ctx, networkInterfaceName, nsgName, virtualNetworkName, subnetName, ipName)
	if err != nil {
		log.Fatal(fmt.Sprintf("Cannot create network interface. Error details: %s", err.Error()))
	}
	fmt.Println("Network interface created")

	// Output:
	// Network interface created
}
