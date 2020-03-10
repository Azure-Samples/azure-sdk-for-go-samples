// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package compute

import (
	"context"
	"fmt"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/network"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/msi/mgmt/2018-11-30/msi"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/pkg/errors"
)

// CreateVMWithMSI creates a virtual machine with a system-assigned managed identity.
func CreateVMWithMSI(ctx context.Context, vmName, nicName, username, password string) (vm compute.VirtualMachine, err error) {
	nic, _ := network.GetNic(ctx, nicName)

	vmClient := getVMClient()
	future, err := vmClient.CreateOrUpdate(
		ctx,
		config.GroupName(),
		vmName,
		compute.VirtualMachine{
			Location: to.StringPtr(config.Location()),
			Identity: &compute.VirtualMachineIdentity{
				Type: compute.ResourceIdentityTypeSystemAssigned,
			},
			VirtualMachineProperties: &compute.VirtualMachineProperties{
				HardwareProfile: &compute.HardwareProfile{
					VMSize: compute.VirtualMachineSizeTypesBasicA0,
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

	err = future.WaitForCompletionRef(ctx, vmClient.Client)
	if err != nil {
		return vm, fmt.Errorf("cannot get the vm create or update future response: %v", err)
	}

	return future.Result(vmClient)
}

// AddIdentityToVM adds a managed identity to an existing VM by activating the
// corresponding VM extension.
func AddIdentityToVM(ctx context.Context, vmName string) (ext compute.VirtualMachineExtension, err error) {
	extensionsClient := getVMExtensionsClient()

	future, err := extensionsClient.CreateOrUpdate(
		ctx,
		config.GroupName(),
		vmName,
		"msiextension",
		compute.VirtualMachineExtension{
			Location: to.StringPtr(config.Location()),
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
		return ext, fmt.Errorf("failed to add MSI extension: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, extensionsClient.Client)
	if err != nil {
		return ext, fmt.Errorf("cannot get the extension create or update future response: %v", err)
	}

	return future.Result(extensionsClient)
}

// CreateVMWithUserAssignedID creates a virtual machine with a user-assigned identity.
func CreateVMWithUserAssignedID(ctx context.Context, vmName, nicName, username, password string, id msi.Identity) (vm compute.VirtualMachine, err error) {
	nic, _ := network.GetNic(ctx, nicName)
	vmClient := getVMClient()
	future, err := vmClient.CreateOrUpdate(
		ctx,
		config.GroupName(),
		vmName,
		compute.VirtualMachine{
			Location: to.StringPtr(config.Location()),
			Identity: &compute.VirtualMachineIdentity{
				Type: compute.ResourceIdentityTypeUserAssigned,
				UserAssignedIdentities: map[string]*compute.VirtualMachineIdentityUserAssignedIdentitiesValue{
					*id.ID: &compute.VirtualMachineIdentityUserAssignedIdentitiesValue{},
				},
			},
			VirtualMachineProperties: &compute.VirtualMachineProperties{
				HardwareProfile: &compute.HardwareProfile{
					VMSize: compute.VirtualMachineSizeTypesBasicA0,
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
		return vm, errors.Wrap(err, "failed to create VM")
	}
	err = future.WaitForCompletionRef(ctx, vmClient.Client)
	if err != nil {
		return vm, errors.Wrap(err, "failed waiting for async operation to complete")
	}
	return future.Result(vmClient)
}

// AddUserAssignedIDToVM adds the specified user-assigned identity to the specified pre-existing VM.
func AddUserAssignedIDToVM(ctx context.Context, vmName string, id msi.Identity) (*compute.VirtualMachine, error) {
	vmClient := getVMClient()
	future, err := vmClient.Update(
		ctx,
		config.GroupName(),
		vmName,
		compute.VirtualMachineUpdate{
			Identity: &compute.VirtualMachineIdentity{
				Type: compute.ResourceIdentityTypeUserAssigned,
				UserAssignedIdentities: map[string]*compute.VirtualMachineIdentityUserAssignedIdentitiesValue{
					*id.ID: &compute.VirtualMachineIdentityUserAssignedIdentitiesValue{},
				},
			},
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update VM")
	}
	err = future.WaitForCompletionRef(ctx, vmClient.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed waiting for async operation to complete")
	}
	vm, err := future.Result(vmClient)
	return &vm, err
}

// RemoveUserAssignedIDFromVM removes the specified user-assigned identity from the specified pre-existing VM.
func RemoveUserAssignedIDFromVM(ctx context.Context, vmName string, id msi.Identity) (*compute.VirtualMachine, error) {
	vmClient := getVMClient()
	future, err := vmClient.Update(
		ctx,
		config.GroupName(),
		vmName,
		compute.VirtualMachineUpdate{
			Identity: &compute.VirtualMachineIdentity{
				Type: compute.ResourceIdentityTypeUserAssigned,
				UserAssignedIdentities: map[string]*compute.VirtualMachineIdentityUserAssignedIdentitiesValue{
					*id.ID: nil,
				},
			},
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update VM")
	}
	err = future.WaitForCompletionRef(ctx, vmClient.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed waiting for async operation to complete")
	}
	vm, err := future.Result(vmClient)
	return &vm, err
}
