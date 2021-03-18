// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package compute

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/authorization"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/graphrbac"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/keyvault"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/msi"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/network"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/marstr/randname"
)

// Example_createVM creates a group and network artifacts needed for a VM, then
// creates a VM and tests operations on it.
func Example_createVM() {
	var groupName = config.GenerateGroupName("VM")
	// TODO: remove and use local `groupName` only
	config.SetGroupName(groupName)

	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()
	defer resources.Cleanup(ctx)

	_, err := resources.CreateGroup(ctx, groupName)
	if err != nil {
		util.LogAndPanic(err)
	}

	_, err = network.CreateVirtualNetworkAndSubnets(ctx, virtualNetworkName, subnet1Name, subnet2Name)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created vnet and 2 subnets")

	_, err = network.CreateNetworkSecurityGroup(ctx, nsgName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created network security group")

	_, err = network.CreatePublicIP(ctx, ipName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created public IP")

	_, err = network.CreateNIC(ctx, virtualNetworkName, subnet1Name, nsgName, ipName, nicName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created nic")

	_, err = CreateVM(ctx, vmName, nicName, username, password, sshPublicKeyPath)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created VM")

	// set or change VM metadata
	_, err = UpdateVM(ctx, vmName, map[string]*string{
		"runtime": to.StringPtr("go"),
		"cloud":   to.StringPtr("azure"),
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("updated VM")

	// set or change system state
	_, err = StartVM(ctx, vmName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("started VM")

	_, err = RestartVM(ctx, vmName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("restarted VM")

	_, err = StopVM(ctx, vmName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("stopped VM")

	// Output:
	// created vnet and 2 subnets
	// created network security group
	// created public IP
	// created nic
	// created VM
	// updated VM
	// started VM
	// restarted VM
	// stopped VM
}

func Example_createVMWithMSI() {
	var groupName = config.GenerateGroupName("VMWithMSI")
	// TODO: remove and use local `groupName` only
	config.SetGroupName(groupName)

	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()
	defer resources.Cleanup(ctx)

	_, err := resources.CreateGroup(ctx, groupName)
	if err != nil {
		util.LogAndPanic(err)
	}

	_, err = network.CreateVirtualNetworkAndSubnets(ctx, virtualNetworkName, subnet1Name, subnet2Name)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created vnet and 2 subnets")

	_, err = network.CreateNetworkSecurityGroup(ctx, nsgName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created network security group")

	_, err = network.CreatePublicIP(ctx, ipName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created public IP")

	_, err = network.CreateNIC(ctx, virtualNetworkName, subnet1Name, nsgName, ipName, nicName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created nic")

	_, err = CreateVMWithMSI(ctx, vmName, nicName, username, password)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created VM")

	_, err = AddIdentityToVM(ctx, vmName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("added MSI extension")

	vm, err := GetVM(ctx, vmName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("got VM")

	list, err := authorization.ListRoleDefinitions(ctx, "roleName eq 'Contributor'")
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("got role definitions list")

	_, err = authorization.AssignRole(ctx, *vm.Identity.PrincipalID, *list.Values()[0].ID)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("role assigned")

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

func Example_createVMWithEncryptedDisks() {
	vaultName := randname.GenerateWithPrefix("gosdk-vault", 10)

	var groupName = config.GenerateGroupName("VMWithEncryptedDisks")
	config.SetGroupName(groupName)

	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()
	defer resources.Cleanup(ctx)

	_, err := resources.CreateGroup(ctx, config.GroupName())
	if err != nil {
		util.LogAndPanic(err)
	}

	_, err = network.CreateVirtualNetworkAndSubnets(ctx, virtualNetworkName, subnet1Name, subnet2Name)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created vnet and subnets")

	_, err = network.CreatePublicIP(ctx, ipName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created public IP")

	_, err = network.CreateNIC(ctx, virtualNetworkName, subnet1Name, "", ipName, nicName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created nic")

	_, err = CreateDisk(ctx, diskName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created disk")

	_, err = CreateVMWithDisk(ctx, nicName, diskName, vmName, username, password)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created virtual machine")

	// create a KV Vault and grant rights to current user
	var userID string
	currentUser, err := graphrbac.GetCurrentUser(ctx)
	if err != nil {
		util.LogAndPanic(err)
	}
	userID = *currentUser.ObjectID
	_, err = keyvault.CreateVaultWithPolicies(ctx, vaultName, userID)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created keyvault")

	key, err := keyvault.CreateKey(ctx, vaultName, "keyName")
	if err != nil {
		util.LogAndPanic(err)

	}
	util.PrintAndLog("created key bundle")

	_, err = AddDiskEncryptionToVM(ctx, vmName, vaultName, *key.Key.Kid)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("added vm encryption extension")

	// Output:
	// created vnet and subnets
	// created keyvault
	// created disk
	// created public IP
	// created nic
	// created virtual machine
	// created key bundle
	// added vm encryption extension
}

func Example_createVMsWithLoadBalancer() {
	var groupName = config.GenerateGroupName("VMsWithLoadBalancer")
	config.SetGroupName(groupName)

	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()
	defer resources.Cleanup(ctx)

	_, err := resources.CreateGroup(ctx, config.GroupName())
	if err != nil {
		util.LogAndPanic(err)
	}

	asName := "as1"

	_, err = network.CreatePublicIP(ctx, ipName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created public IP")

	_, err = network.CreateLoadBalancer(ctx, lbName, ipName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created load balancer")

	_, err = network.CreateVirtualNetwork(ctx, virtualNetworkName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created vnet")

	_, err = network.CreateVirtualNetworkSubnet(ctx, virtualNetworkName, subnet1Name)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created subnet")

	_, err = CreateAvailabilitySet(ctx, asName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created availability set")

	_, err = CreateVMWithLoadBalancer(ctx, "vm1", lbName, virtualNetworkName, subnet1Name, ipName, asName, 0)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created virtual machine on load balance, with NAT rule 1")

	_, err = CreateVMWithLoadBalancer(ctx, "vm2", lbName, virtualNetworkName, subnet1Name, ipName, asName, 1)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created virtual machine on load balance, with NAT rule 2")

	// Output:
	// created public IP
	// created load balancer
	// created vnet
	// created subnet
	// created availability set
	// created virtual machine on load balance, with NAT rule 1
	// created virtual machine on load balance, with NAT rule 2
}

func Example_createVMWithDisks() {
	var groupName = config.GenerateGroupName("VMWithDisks")
	// TODO: remove and use local `groupName` only
	config.SetGroupName(groupName)

	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()
	defer resources.Cleanup(ctx)

	_, err := resources.CreateGroup(ctx, groupName)
	if err != nil {
		util.LogAndPanic(err)
	}

	_, err = network.CreateVirtualNetworkAndSubnets(ctx, virtualNetworkName, subnet1Name, subnet2Name)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created vnet and 2 subnets")

	_, err = network.CreateNetworkSecurityGroup(ctx, nsgName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created network security group")

	_, err = network.CreatePublicIP(ctx, ipName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created public IP")

	_, err = network.CreateNIC(ctx, virtualNetworkName, subnet1Name, nsgName, ipName, nicName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created nic")

	// Disks
	_, err = AttachDataDisk(ctx, vmName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("attached data disks")

	_, err = DetachDataDisks(ctx, vmName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("detached data disks")

	_, err = UpdateOSDiskSize(ctx, vmName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("updated OS disk size")

	// Output:
	// attached data disks
	// detached data disks
	// updated OS disk size
}

func Example_createVMWithUserAssignedIdentity() {
	var groupName = config.GenerateGroupName("VMWithUserAssignedID")
	// TODO: remove and use local `groupName` only
	config.SetGroupName(groupName)

	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()
	defer resources.Cleanup(ctx)

	_, err := resources.CreateGroup(ctx, groupName)
	if err != nil {
		util.LogAndPanic(err)
		return
	}

	_, err = network.CreateVirtualNetworkAndSubnets(ctx, virtualNetworkName, subnet1Name, subnet2Name)
	if err != nil {
		util.LogAndPanic(err)
		return
	}
	util.PrintAndLog("created vnet and 2 subnets")

	_, err = network.CreateNetworkSecurityGroup(ctx, nsgName)
	if err != nil {
		util.LogAndPanic(err)
		return
	}
	util.PrintAndLog("created network security group")

	_, err = network.CreatePublicIP(ctx, ipName)
	if err != nil {
		util.LogAndPanic(err)
		return
	}
	util.PrintAndLog("created public IP")

	_, err = network.CreateNIC(ctx, virtualNetworkName, subnet1Name, nsgName, ipName, nicName)
	if err != nil {
		util.LogAndPanic(err)
		return
	}
	util.PrintAndLog("created nic")

	id1, err := msi.CreateUserAssignedIdentity(groupName, "useridentity1")
	if err != nil {
		util.LogAndPanic(err)
		return
	}
	util.PrintAndLog("created first user-assigned identity")

	_, err = CreateVMWithUserAssignedID(ctx, vmName, nicName, username, password, *id1)
	if err != nil {
		util.LogAndPanic(err)
		return
	}
	util.PrintAndLog("created VM")

	id2, err := msi.CreateUserAssignedIdentity(groupName, "useridentity2")
	if err != nil {
		util.LogAndPanic(err)
		return
	}
	util.PrintAndLog("created second user-assigned identity")

	_, err = AddUserAssignedIDToVM(ctx, vmName, *id2)
	if err != nil {
		util.LogAndPanic(err)
		return
	}
	util.PrintAndLog("added second user-assigned identity to VM")

	_, err = RemoveUserAssignedIDFromVM(ctx, vmName, *id1)
	if err != nil {
		util.LogAndPanic(err)
		return
	}
	util.PrintAndLog("removed first user-assigned identity from VM")

	// Output:
	// created vnet and 2 subnets
	// created network security group
	// created public IP
	// created nic
	// created first user-assigned identity
	// created VM
	// created second user-assigned identity
	// added second user-assigned identity to VM
	// removed first user-assigned identity from VM
}

func Example_list() {
	// list the VMs we've created in our resource group by page.
	// uses the default page size returned by the service.
	vmClient := getVMClient()
	for page, err := vmClient.List(context.Background(), config.GroupName()); page.NotDone(); err = page.Next() {
		if err != nil {
			util.LogAndPanic(err)
		}
		// print out the list of VMs per page
		for _, vm := range page.Values() {
			util.PrintAndLog(fmt.Sprintf("found VM with name %s", *vm.Name))
		}
	}
}

func Example_listComplete() {
	// list the VMs we've created in our resource group using an iterator.
	vmClient := getVMClient()
	for iter, err := vmClient.ListComplete(context.Background(), config.GroupName()); iter.NotDone(); err = iter.Next() {
		if err != nil {
			util.LogAndPanic(err)
		}
		util.PrintAndLog(fmt.Sprintf("found VM with name %s", *iter.Value().Name))
	}
}

func Example_get() {
	// retrieve information about a specific VM
	vmClient := getVMClient()
	vm, err := vmClient.Get(context.Background(), config.GroupName(), vmName, compute.InstanceView)
	if err != nil {
		util.LogAndPanic(err)
	}
	fmt.Printf("Name: %s\n", *vm.Name)
	fmt.Printf("ID: %s\n", *vm.ID)
	fmt.Printf("VM size: %s\n", vm.HardwareProfile.VMSize)
}
