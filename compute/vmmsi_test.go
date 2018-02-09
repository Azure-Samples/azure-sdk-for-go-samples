// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package compute

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/authorization"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/network"
)

func ExampleCreateVMForMSI() {
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

	_, err = CreateVMForMSI(ctx, vmName, nicName, username, password)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created VM")

	_, err = AddMSIExtension(ctx, vmName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("added MSI extension")

	vm, err := GetVM(ctx, vmName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("got VM")

	list, err := authorization.ListRoles(ctx, "roleName eq 'Contributor'")
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("got role definitions list")

	_, err = authorization.AssignRole(ctx, *vm.Identity.PrincipalID, *list.Values()[0].ID)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("role assigned")

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
