// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package compute

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/authorization"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/network"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

func ExampleCreateVMForMSI() {
	internal.SetResourceGroupName("CreateVMForMSI")
	ctx := context.Background()
	defer resources.Cleanup(ctx)
	_, err := resources.CreateGroup(ctx, internal.ResourceGroupName())
	if err != nil {
		internal.PrintAndLog(err.Error())
	}

	_, err = network.CreateVirtualNetworkAndSubnets(ctx, virtualNetworkName, subnet1Name, subnet2Name)
	if err != nil {
		internal.PrintAndLog(err.Error())
	}
	internal.PrintAndLog("created vnet and 2 subnets")

	_, err = network.CreateNetworkSecurityGroup(ctx, nsgName)
	if err != nil {
		internal.PrintAndLog(err.Error())
	}
	internal.PrintAndLog("created network security group")

	_, err = network.CreatePublicIP(ctx, ipName)
	if err != nil {
		internal.PrintAndLog(err.Error())
	}
	internal.PrintAndLog("created public IP")

	_, err = network.CreateNIC(ctx, virtualNetworkName, subnet1Name, nsgName, ipName, nicName)
	if err != nil {
		internal.PrintAndLog(err.Error())
	}
	internal.PrintAndLog("created nic")

	_, err = CreateVMForMSI(ctx, vmName, nicName, username, password)
	if err != nil {
		internal.PrintAndLog(err.Error())
	}
	internal.PrintAndLog("created VM")

	_, err = AddMSIExtension(ctx, vmName)
	if err != nil {
		internal.PrintAndLog(err.Error())
	}
	internal.PrintAndLog("added MSI extension")

	vm, err := GetVM(ctx, vmName)
	if err != nil {
		internal.PrintAndLog(err.Error())
	}
	internal.PrintAndLog("got VM")

	list, err := authorization.ListRoles(ctx, "roleName eq 'Contributor'")
	if err != nil {
		internal.PrintAndLog(err.Error())
	}
	internal.PrintAndLog("got role definitions list")

	_, err = authorization.AssignRole(ctx, *vm.Identity.PrincipalID, *list.Values()[0].ID)
	if err != nil {
		internal.PrintAndLog(err.Error())
	}
	internal.PrintAndLog("role assigned")

	// Output:
	// created vnet and 2 subnets
	// created network security group
	// created public IP
	// created nic
	// created VM
	// added MSI extension
	// got VM
	// got role definitions list
	// role assigned
}
