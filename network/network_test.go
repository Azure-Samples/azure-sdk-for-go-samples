// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package network

import (
	"context"
	"flag"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

var (
	virtualNetworkName = "vnet1"
	subnet1Name        = "subnet1"
	subnet2Name        = "subnet2"
	nsgName            = "nsg1"
	nicName            = "nic1"
	ipName             = "ip1"
)

func TestMain(m *testing.M) {
	err := setupEnvironment()
	if err != nil {
		log.Fatalf("could not set up environment: %v\n", err)
	}

	os.Exit(m.Run())
}

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
	groupName := config.GenerateGroupName("network")
	config.SetGroupName(groupName) // TODO: don't use globals

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()
	defer resources.Cleanup(ctx)

	_, err := resources.CreateGroup(ctx, groupName)
	if err != nil {
		t.Fatalf("failed to create group: %+v", err)
	}
	t.Logf("created group %s\n", groupName)

	_, err = CreateVirtualNetworkAndSubnets(ctx, virtualNetworkName, subnet1Name, subnet2Name)
	if err != nil {
		t.Fatalf("failed to create vnet: %+v", err)
	}
	t.Logf("created vnet with 2 subnets")

	_, err = CreateNetworkSecurityGroup(ctx, nsgName)
	if err != nil {
		t.Fatalf("failed to create NSG: %+v", err)
	}
	t.Logf("created network security group")

	_, err = CreatePublicIP(ctx, ipName)
	if err != nil {
		t.Fatalf("failed to create public IP: %+v", err)
	}
	t.Logf("created public IP")

	_, err = CreateNIC(ctx, virtualNetworkName, subnet1Name, nsgName, ipName, nicName)
	if err != nil {
		t.Fatalf("failed to create NIC: %+v", err)
	}
	t.Logf("created nic")
}

func ExampleCreateNIC() {
	groupName := config.GenerateGroupName("CreateNIC")
	config.SetGroupName(groupName) // TODO: don't use globals
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()
	defer resources.Cleanup(ctx)

	_, err := resources.CreateGroup(ctx, config.GroupName())
	if err != nil {
		util.LogAndPanic(err)
	}

	_, err = CreateVirtualNetworkAndSubnets(ctx, virtualNetworkName, subnet1Name, subnet2Name)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created vnet and 2 subnets")

	_, err = CreateNetworkSecurityGroup(ctx, nsgName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created network security group")

	_, err = CreatePublicIP(ctx, ipName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created public IP")

	_, err = CreateNIC(ctx, virtualNetworkName, subnet1Name, nsgName, ipName, nicName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created nic")

	// Output:
	// created vnet and 2 subnets
	// created network security group
	// created public IP
	// created nic
}

func ExampleCreateNetworkSecurityGroup() {
	groupName := config.GenerateGroupName("CreateNSG")
	config.SetGroupName(groupName) // TODO: don't use globals
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()
	defer resources.Cleanup(ctx)

	_, err := resources.CreateGroup(ctx, config.GroupName())
	if err != nil {
		util.LogAndPanic(err)
	}

	_, err = CreateVirtualNetwork(ctx, virtualNetworkName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created vnet")

	frontNSGName := "frontend"
	backNSGName := "backend"

	_, err = CreateNetworkSecurityGroup(ctx, frontNSGName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created frontend network security group")

	_, err = CreateNetworkSecurityGroup(ctx, backNSGName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created backend network security group")

	frontEndAddressPrefix := "10.0.0.0/16"
	_, err = CreateSubnetWithNetworkSecurityGroup(ctx, virtualNetworkName, "frontend", frontEndAddressPrefix, frontNSGName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created subnet with frontend network security group")

	_, err = CreateSubnetWithNetworkSecurityGroup(ctx, virtualNetworkName, "backend", "10.1.0.0/16", backNSGName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created subnet with backend network security group")

	_, err = CreateSSHRule(ctx, frontNSGName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created frontend SSH security rule")

	_, err = CreateHTTPRule(ctx, frontNSGName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created frontend HTTP security rule")

	_, err = CreateSQLRule(ctx, frontNSGName, frontEndAddressPrefix)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created frontend SQL security rule")

	_, err = CreateDenyOutRule(ctx, backNSGName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created backend deny out security rule")

	// Output:
	// created vnet
	// created frontend network security group
	// created backend network security group
	// created subnet with frontend network security group
	// created subnet with backend network security group
	// created frontend SSH security rule
	// created frontend HTTP security rule
	// created frontend SQL security rule
	// created backend deny out security rule
}
