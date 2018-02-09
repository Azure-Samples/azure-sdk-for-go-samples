// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package compute

import (
	"context"
	"fmt"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/network"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2017-03-30/compute"
	"github.com/Azure/go-autorest/autorest/to"
)

// CreateVMForMSI creates a virtual machine with a systems assigned identity type
func CreateVMForMSI(ctx context.Context, vmName, nicName, username, password string) (vm compute.VirtualMachine, err error) {
	nic, _ := network.GetNic(ctx, nicName)

	vmClient := getVMClient()
	future, err := vmClient.CreateOrUpdate(
		ctx,
		helpers.ResourceGroupName(),
		vmName,
		compute.VirtualMachine{
			Location: to.StringPtr(helpers.Location()),
			Identity: &compute.VirtualMachineIdentity{
				Type: compute.SystemAssigned, // needed to add MSI authentication
			},
			VirtualMachineProperties: &compute.VirtualMachineProperties{
				HardwareProfile: &compute.HardwareProfile{
					VMSize: compute.StandardDS1V2,
				},
				StorageProfile: &compute.StorageProfile{
					ImageReference: &compute.ImageReference{
						Publisher: to.StringPtr(publisher),
						Offer:     to.StringPtr(offer),
						Sku:       to.StringPtr(sku),
						Version:   to.StringPtr("latest"),
					},
				},
				OsProfile: &compute.OSProfile{
					ComputerName:  to.StringPtr(vmName),
					AdminUsername: to.StringPtr(username),
					AdminPassword: to.StringPtr(password),
				},
				NetworkProfile: &compute.NetworkProfile{
					NetworkInterfaces: &[]compute.NetworkInterfaceReference{
						{
							ID: nic.ID,
							NetworkInterfaceReferenceProperties: &compute.NetworkInterfaceReferenceProperties{
								Primary: to.BoolPtr(true),
							},
						},
					},
				},
			},
		},
	)
	if err != nil {
		return vm, fmt.Errorf("cannot create vm: %v", err)
	}

	err = future.WaitForCompletion(ctx, vmClient.Client)
	if err != nil {
		return vm, fmt.Errorf("cannot get the vm create or update future response: %v", err)
	}

	return future.Result(vmClient)
}

// AddMSIExtension adds the MSI (managed service identity) extension to a virtual machine.
func AddMSIExtension(ctx context.Context, vmName string) (ext compute.VirtualMachineExtension, err error) {
	extClient := getExtensionClient()

	future, err := extClient.CreateOrUpdate(
		ctx,
		helpers.ResourceGroupName(),
		vmName,
		"msiextension", compute.VirtualMachineExtension{
			Location: to.StringPtr(helpers.Location()),
			VirtualMachineExtensionProperties: &compute.VirtualMachineExtensionProperties{
				Publisher:               to.StringPtr("Microsoft.ManagedIdentity"),
				Type:                    to.StringPtr("ManagedIdentityExtensionForLinux"),
				TypeHandlerVersion:      to.StringPtr("1.0"),
				AutoUpgradeMinorVersion: to.BoolPtr(true),
				Settings: &map[string]interface{}{
					"port": "50342",
				},
			},
		})
	if err != nil {
		return ext, fmt.Errorf("cannot add MSI extension: %v", err)
	}

	err = future.WaitForCompletion(ctx, extClient.Client)
	if err != nil {
		return ext, fmt.Errorf("cannot get the extension create or update future response: %v", err)
	}

	return future.Result(extClient)
}
