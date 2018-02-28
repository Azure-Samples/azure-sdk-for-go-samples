// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package compute

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/network"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

func ExampleCreateVMsWithLoadBalancer() {
	helpers.SetResourceGroupName("CreateVMsWithLoadBalancer")
	ctx := context.Background()
	defer resources.Cleanup(ctx)
	_, err := resources.CreateGroup(ctx, helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}

	avaSetName := "availSet"

	_, err = network.CreatePublicIP(ctx, ipName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created public IP")

	_, err = network.CreateLoadBalancer(ctx, lbName, ipName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created load balancer")

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

	_, err = CreateAvailabilitySet(ctx, avaSetName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created availability set")

	_, err = CreateVMWithLoadBalancer(ctx, "vm1", lbName, virtualNetworkName, subnet1Name, ipName, avaSetName, 0)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created virtual machine on load balance, with NAT rule 1")

	_, err = CreateVMWithLoadBalancer(ctx, "vm2", lbName, virtualNetworkName, subnet1Name, ipName, avaSetName, 1)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created virtual machine on load balance, with NAT rule 2")

	// Output:
	// created public IP
	// created load balancer
	// created vnet
	// created subnet
	// created availability set
	// created virtual machine on load balance, with NAT rule 1
	// created virtual machine on load balance, with NAT rule 2
}
