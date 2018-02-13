// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package compute

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/network"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2017-03-30/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

const (
	publisher = "Canonical"
	offer     = "UbuntuServer"
	sku       = "16.04.0-LTS"
)

// fakepubkey is used if a key isn't available at the specified path in CreateVM(...)
var fakepubkey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC7laRyN4B3YZmVrDEZLZoIuUA72pQ0DpGuZBZWykCofIfCPrFZAJgFvonKGgKJl6FGKIunkZL9Us/mV4ZPkZhBlE7uX83AAf5i9Q8FmKpotzmaxN10/1mcnEE7pFvLoSkwqrQSkrrgSm8zaJ3g91giXSbtqvSIj/vk2f05stYmLfhAwNo3Oh27ugCakCoVeuCrZkvHMaJgcYrIGCuFo6q0Pfk9rsZyriIqEa9AtiUOtViInVYdby7y71wcbl0AbbCZsTSqnSoVxm2tRkOsXV6+8X4SnwcmZbao3H+zfO1GBhQOLxJ4NQbzAa8IJh810rYARNLptgmsd4cYXVOSosTX azureuser"

func getVMClient() compute.VirtualMachinesClient {
	token, _ := iam.GetResourceManagementToken(iam.AuthGrantType())
	vmClient := compute.NewVirtualMachinesClient(helpers.SubscriptionID())
	vmClient.Authorizer = autorest.NewBearerAuthorizer(token)
	vmClient.AddToUserAgent(helpers.UserAgent())
	return vmClient
}

func getExtensionClient() compute.VirtualMachineExtensionsClient {
	token, _ := iam.GetResourceManagementToken(iam.AuthGrantType())
	extClient := compute.NewVirtualMachineExtensionsClient(helpers.SubscriptionID())
	extClient.Authorizer = autorest.NewBearerAuthorizer(token)
	extClient.AddToUserAgent(helpers.UserAgent())
	return extClient
}

func getDisksClient() compute.DisksClient {
	token, _ := iam.GetResourceManagementToken(iam.AuthGrantType())
	disksClient := compute.NewDisksClient(helpers.SubscriptionID())
	disksClient.Authorizer = autorest.NewBearerAuthorizer(token)
	disksClient.AddToUserAgent(helpers.UserAgent())
	return disksClient
}

// CreateVM creates a new virtual machine with the specified name using the specified NIC.
// Username, password, and sshPublicKeyPath determine logon credentials.
func CreateVM(ctx context.Context, vmName, nicName, username, password, sshPublicKeyPath string) (vm compute.VirtualMachine, err error) {
	nic, _ := network.GetNic(ctx, nicName)

	var sshKeyData string
	if _, err = os.Stat(sshPublicKeyPath); err == nil {
		sshBytes, err := ioutil.ReadFile(sshPublicKeyPath)
		if err != nil {
			log.Fatalf("failed to read SSH key data: %v", err)
		}
		sshKeyData = string(sshBytes)
	} else {
		sshKeyData = fakepubkey
	}

	vmClient := getVMClient()
	future, err := vmClient.CreateOrUpdate(
		ctx,
		helpers.ResourceGroupName(),
		vmName,
		compute.VirtualMachine{
			Location: to.StringPtr(helpers.Location()),
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
					LinuxConfiguration: &compute.LinuxConfiguration{
						SSH: &compute.SSHConfiguration{
							PublicKeys: &[]compute.SSHPublicKey{
								{
									Path:    to.StringPtr(fmt.Sprintf("/home/%s/.ssh/authorized_keys", username)),
									KeyData: to.StringPtr(sshKeyData),
								},
							},
						},
					},
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

// GetVM gets the specified VM info
func GetVM(ctx context.Context, vmName string) (compute.VirtualMachine, error) {
	vmClient := getVMClient()
	return vmClient.Get(ctx, helpers.ResourceGroupName(), vmName, compute.InstanceView)
}

// UpdateVM adds tags to the VM
func UpdateVM(ctx context.Context, vmName string, tags map[string]*string) (vm compute.VirtualMachine, err error) {
	vm, err = GetVM(ctx, vmName)
	if err != nil {
		return
	}

	vm.Tags = tags

	vmClient := getVMClient()
	future, err := vmClient.CreateOrUpdate(ctx, helpers.ResourceGroupName(), vmName, vm)
	if err != nil {
		return vm, fmt.Errorf("cannot update vm: %v", err)
	}

	err = future.WaitForCompletion(ctx, vmClient.Client)
	if err != nil {
		return vm, fmt.Errorf("cannot get the vm create or update future response: %v", err)
	}

	return future.Result(vmClient)
}

// AttachDataDisks attaches a 1 GB data disk to the selected VM
func AttachDataDisks(ctx context.Context, vmName string) (vm compute.VirtualMachine, err error) {
	vm, err = GetVM(ctx, vmName)
	if err != nil {
		return vm, fmt.Errorf("cannot get vm: %v", err)
	}

	vm.StorageProfile.DataDisks = &[]compute.DataDisk{
		{
			Lun:          to.Int32Ptr(0),
			Name:         to.StringPtr("dataDisk"),
			CreateOption: compute.DiskCreateOptionTypesEmpty,
			DiskSizeGB:   to.Int32Ptr(1),
		},
	}

	vmClient := getVMClient()
	future, err := vmClient.CreateOrUpdate(ctx, helpers.ResourceGroupName(), vmName, vm)
	if err != nil {
		return vm, fmt.Errorf("cannot update vm: %v", err)
	}

	err = future.WaitForCompletion(ctx, vmClient.Client)
	if err != nil {
		return vm, fmt.Errorf("cannot get the vm create or update future response: %v", err)
	}

	return future.Result(vmClient)
}

// DetachDataDisks detaches all data disks from the selected VM
func DetachDataDisks(ctx context.Context, vmName string) (vm compute.VirtualMachine, err error) {
	vm, err = GetVM(ctx, vmName)
	if err != nil {
		return vm, fmt.Errorf("cannot get vm: %v", err)
	}

	vm.StorageProfile.DataDisks = &[]compute.DataDisk{}

	vmClient := getVMClient()
	future, err := vmClient.CreateOrUpdate(ctx, helpers.ResourceGroupName(), vmName, vm)
	if err != nil {
		return vm, fmt.Errorf("cannot update vm: %v", err)
	}

	err = future.WaitForCompletion(ctx, vmClient.Client)
	if err != nil {
		return vm, fmt.Errorf("cannot get the vm create or update future response: %v", err)
	}

	return future.Result(vmClient)
}

// Deallocate deallocates the selected VM
func Deallocate(ctx context.Context, vmName string) (osr compute.OperationStatusResponse, err error) {
	vmClient := getVMClient()
	future, err := vmClient.Deallocate(ctx, helpers.ResourceGroupName(), vmName)
	if err != nil {
		return osr, fmt.Errorf("cannot deallocate vm: %v", err)
	}

	err = future.WaitForCompletion(ctx, vmClient.Client)
	if err != nil {
		return osr, fmt.Errorf("cannot get the vm deallocate future response: %v", err)
	}

	return future.Result(vmClient)
}

func getDisk(ctx context.Context, diskName string) (disk compute.Disk, err error) {
	disksClient := getDisksClient()
	return disksClient.Get(ctx, helpers.ResourceGroupName(), diskName)
}

// UpdateOSDiskSize increases the selected VM's OS disk size
func UpdateOSDiskSize(ctx context.Context, vmName string) (d compute.Disk, err error) {
	vm, err := GetVM(ctx, vmName)
	if err != nil {
		return d, fmt.Errorf("cannot get vm: %v", err)
	}

	size := vm.StorageProfile.OsDisk.DiskSizeGB
	if size == nil {
		size = to.Int32Ptr(0)
	}
	if *size <= 0 {
		*size = 256
	}
	*size += 10

	_, err = Deallocate(ctx, vmName)
	if err != nil {
		return d, fmt.Errorf("cannot deallocate vm: %v", err)
	}

	disksClient := getDisksClient()
	future, err := disksClient.Update(ctx, helpers.ResourceGroupName(), *vm.StorageProfile.OsDisk.Name, compute.DiskUpdate{
		DiskUpdateProperties: &compute.DiskUpdateProperties{
			DiskSizeGB: size,
		},
	})

	err = future.WaitForCompletion(ctx, disksClient.Client)
	if err != nil {
		return d, fmt.Errorf("cannot get the disk update future response: %v", err)
	}

	return future.Result(disksClient)
}

// StartVM starts the selected VM
func StartVM(ctx context.Context, vmName string) (osr compute.OperationStatusResponse, err error) {
	vmClient := getVMClient()
	future, err := vmClient.Start(ctx, helpers.ResourceGroupName(), vmName)
	if err != nil {
		return osr, fmt.Errorf("cannot start vm: %v", err)
	}

	err = future.WaitForCompletion(ctx, vmClient.Client)
	if err != nil {
		return osr, fmt.Errorf("cannot get the vm start future response: %v", err)
	}

	return future.Result(vmClient)
}

// RestartVM restarts the selected VM
func RestartVM(ctx context.Context, vmName string) (osr compute.OperationStatusResponse, err error) {
	vmClient := getVMClient()
	future, err := vmClient.Restart(ctx, helpers.ResourceGroupName(), vmName)
	if err != nil {
		return osr, fmt.Errorf("cannot restart vm: %v", err)
	}

	err = future.WaitForCompletion(ctx, vmClient.Client)
	if err != nil {
		return osr, fmt.Errorf("cannot get the vm restart future response: %v", err)
	}

	return future.Result(vmClient)
}

// PowerOffVM stops the selected VM
func PowerOffVM(ctx context.Context, vmName string) (osr compute.OperationStatusResponse, err error) {
	vmClient := getVMClient()
	future, err := vmClient.PowerOff(ctx, helpers.ResourceGroupName(), vmName)
	if err != nil {
		return osr, fmt.Errorf("cannot power off vm: %v", err)
	}

	err = future.WaitForCompletion(ctx, vmClient.Client)
	if err != nil {
		return osr, fmt.Errorf("cannot get the vm power off future response: %v", err)
	}

	return future.Result(vmClient)
}
