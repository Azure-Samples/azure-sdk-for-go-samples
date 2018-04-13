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

	"github.com/subosito/gotenv"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
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
	err := parseArgs()
	if err != nil {
		log.Fatalln("failed to parse args")
	}

	err = iam.ParseArgs()
	if err != nil {
		log.Fatalln("failed to parse IAM args")
	}

	os.Exit(m.Run())
}

func parseArgs() error {
	gotenv.Load()

	virtualNetworkName = os.Getenv("AZURE_VNET_NAME")
	flag.StringVar(&virtualNetworkName, "vnetName", virtualNetworkName, "Specify a name for the vnet.")

	if !(len(virtualNetworkName) > 0) {
		virtualNetworkName = "vnet1"
	}

	return nil
}

func ExampleCreateNIC() {
	helpers.SetResourceGroupName("CreateNIC")
	ctx := context.Background()
	defer resources.Cleanup(ctx)
	_, err := resources.CreateGroup(ctx, helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}

	_, err = CreateVirtualNetworkAndSubnets(ctx, virtualNetworkName, subnet1Name, subnet2Name)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created vnet and 2 subnets")

	_, err = CreateNetworkSecurityGroup(ctx, nsgName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created network security group")

	_, err = CreatePublicIP(ctx, ipName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created public IP")

	_, err = CreateNIC(ctx, virtualNetworkName, subnet1Name, nsgName, ipName, nicName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created nic")

	// Output:
	// created vnet and 2 subnets
	// created network security group
	// created public IP
	// created nic
}

func ExampleCreateNetworkSecurityGroup() {
	helpers.SetResourceGroupName("CreateNSG")
	ctx := context.Background()
	defer resources.Cleanup(ctx)
	_, err := resources.CreateGroup(ctx, helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}

	_, err = CreateVirtualNetwork(ctx, virtualNetworkName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created vnet")

	frontNSGName := "frontend"
	backNSGName := "backend"

	_, err = CreateNetworkSecurityGroup(ctx, frontNSGName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created frontend network security group")

	_, err = CreateNetworkSecurityGroup(ctx, backNSGName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created backend network security group")

	frontEndAddressPrefix := "10.0.0.0/16"
	_, err = CreateSubnetWithNetowrkSecurityGroup(ctx, virtualNetworkName, "frontend", frontEndAddressPrefix, frontNSGName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created subnet with frontend network security group")

	_, err = CreateSubnetWithNetowrkSecurityGroup(ctx, virtualNetworkName, "backend", "10.1.0.0/16", backNSGName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created subnet with backend network security group")

	_, err = CreateSSHRule(ctx, frontNSGName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created frontend SSH security rule")

	_, err = CreateHTTPRule(ctx, frontNSGName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created frontend HTTP security rule")

	_, err = CreateSQLRule(ctx, frontNSGName, frontEndAddressPrefix)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created frontend SQL security rule")

	_, err = CreateDenyOutRule(ctx, backNSGName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created backend deny out security rule")

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
