// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package compute

import (
	"context"
	"fmt"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/network"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/satori/go.uuid"
)

func getDisksClient() compute.DisksClient {
	disksClient := compute.NewDisksClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	disksClient.Authorizer = a
	disksClient.AddToUserAgent(config.UserAgent())
	return disksClient
}

func getDisk(ctx context.Context, diskName string) (disk compute.Disk, err error) {
	disksClient := getDisksClient()
	return disksClient.Get(ctx, config.GroupName(), diskName)
}

// AttachDataDisk attaches a 1GB data disk to the specified VM.
func AttachDataDisk(ctx context.Context, vmName string) (vm compute.VirtualMachine, err error) {
	// first GET the VM object
	vm, err = GetVM(ctx, vmName)
	if err != nil {
		return vm, fmt.Errorf("cannot get vm: %v", err)
	}

	// then update it
	vm.StorageProfile.DataDisks = &[]compute.DataDisk{{
		Lun:          to.Int32Ptr(0),
		Name:         to.StringPtr("gosdksamples-datadisk"),
		CreateOption: compute.DiskCreateOptionTypesEmpty,
		DiskSizeGB:   to.Int32Ptr(1),
	}}

	// then PUT it back
	vmClient := getVMClient()
	future, err := vmClient.CreateOrUpdate(ctx, config.GroupName(), vmName, vm)
	if err != nil {
		return vm, fmt.Errorf("cannot update vm: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, vmClient.Client)
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
	future, err := vmClient.CreateOrUpdate(ctx, config.GroupName(), vmName, vm)
	if err != nil {
		return vm, fmt.Errorf("cannot update vm: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, vmClient.Client)
	if err != nil {
		return vm, fmt.Errorf("cannot get the vm create or update future response: %v", err)
	}

	return future.Result(vmClient)
}

// UpdateOSDiskSize increases the selected VM's OS disk size by 10GB.
func UpdateOSDiskSize(ctx context.Context, vmName string) (d compute.Disk, err error) {
	vm, err := GetVM(ctx, vmName)
	if err != nil {
		return d, fmt.Errorf("cannot get vm: %v", err)
	}

	sizeGB := vm.StorageProfile.OsDisk.DiskSizeGB
	if sizeGB == nil {
		sizeGB = to.Int32Ptr(0)
	}
	if *sizeGB <= 0 {
		*sizeGB = 256
	}
	*sizeGB += 10

	_, err = DeallocateVM(ctx, vmName)
	if err != nil {
		return d, fmt.Errorf("cannot deallocate vm: %v", err)
	}

	disksClient := getDisksClient()
	future, err := disksClient.Update(ctx,
		config.GroupName(),
		*vm.StorageProfile.OsDisk.Name,
		compute.DiskUpdate{
			DiskUpdateProperties: &compute.DiskUpdateProperties{
				DiskSizeGB: sizeGB,
			},
		})
	if err != nil {
		return d, fmt.Errorf("cannot update disk: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, disksClient.Client)
	if err != nil {
		return d, fmt.Errorf("cannot get the disk update future response: %v", err)
	}

	return future.Result(disksClient)
}

// CreateDisk creates an empty 64GB disk which can be attached to a VM.
func CreateDisk(ctx context.Context, diskName string) (disk compute.Disk, err error) {
	disksClient := getDisksClient()
	future, err := disksClient.CreateOrUpdate(
		ctx,
		config.GroupName(),
		diskName,
		compute.Disk{
			Location: to.StringPtr(config.Location()),
			DiskProperties: &compute.DiskProperties{
				CreationData: &compute.CreationData{
					CreateOption: compute.Empty,
				},
				DiskSizeGB: to.Int32Ptr(64),
			},
		})
	if err != nil {
		return disk, fmt.Errorf("cannot create disk: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, disksClient.Client)
	if err != nil {
		return disk, fmt.Errorf("cannot get the disk create or update future response: %v", err)
	}

	return future.Result(disksClient)
}

// CreateVMWithDisk creates a VM, attaching an already existing data disk
func CreateVMWithDisk(ctx context.Context, nicName, diskName, vmName, username, password string) (vm compute.VirtualMachine, err error) {

	nic, _ := network.GetNic(ctx, nicName)
	disk, _ := getDisk(ctx, diskName)

	vmClient := getVMClient()
	future, err := vmClient.CreateOrUpdate(
		ctx,
		config.GroupName(),
		vmName, compute.VirtualMachine{
			Location: to.StringPtr(config.Location()),
			VirtualMachineProperties: &compute.VirtualMachineProperties{
				HardwareProfile: &compute.HardwareProfile{
					VMSize: compute.VirtualMachineSizeTypesBasicA0,
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
				OsProfile: &compute.OSProfile{
					ComputerName:  to.StringPtr(vmName),
					AdminUsername: to.StringPtr(username),
					AdminPassword: to.StringPtr(password),
					LinuxConfiguration: &compute.LinuxConfiguration{
						DisablePasswordAuthentication: to.BoolPtr(false),
					},
				},
				StorageProfile: &compute.StorageProfile{
					ImageReference: &compute.ImageReference{
						Publisher: to.StringPtr(publisher),
						Offer:     to.StringPtr(offer),
						Sku:       to.StringPtr(sku),
						Version:   to.StringPtr("latest"),
					},
					OsDisk: &compute.OSDisk{
						CreateOption: compute.DiskCreateOptionTypesFromImage,
						DiskSizeGB:   to.Int32Ptr(64),
					},
					DataDisks: &[]compute.DataDisk{
						{
							CreateOption: compute.DiskCreateOptionTypesAttach,
							Lun:          to.Int32Ptr(0),
							ManagedDisk: &compute.ManagedDiskParameters{
								ID: disk.ID,
							},
						},
					},
				},
			},
		})
	if err != nil {
		return vm, fmt.Errorf("cannot create vm: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, vmClient.Client)
	if err != nil {
		return vm, fmt.Errorf("cannot get the vm create or update future response: %v", err)
	}

	return future.Result(vmClient)
}

// AddDiskEncryptionToVM adds an extension to a VM to enable use of encryption
// keys from Key Vault to decrypt disks.
func AddDiskEncryptionToVM(ctx context.Context, vmName, vaultName, keyID string) (ext compute.VirtualMachineExtension, err error) {
	extensionsClient := getVMExtensionsClient()
	future, err := extensionsClient.CreateOrUpdate(
		ctx,
		config.GroupName(),
		vmName,
		"AzureDiskEncryptionForLinux",
		compute.VirtualMachineExtension{
			Location: to.StringPtr(config.Location()),
			VirtualMachineExtensionProperties: &compute.VirtualMachineExtensionProperties{
				AutoUpgradeMinorVersion: to.BoolPtr(true),
				ProtectedSettings: &map[string]interface{}{
					"AADClientSecret": config.ClientSecret(), // replace with your own
					"Passphrase":      "yourPassPhrase",
				},
				Publisher: to.StringPtr("Microsoft.Azure.Security"),
				Settings: &map[string]interface{}{
					"AADClientID":               config.ClientID(), // replace with your own
					"EncryptionOperation":       "EnableEncryption",
					"KeyEncryptionAlgorithm":    "RSA-OAEP",
					"KeyEncryptionKeyAlgorithm": keyID,
					"KeyVaultURL":               fmt.Sprintf("https://%s.%s/", vaultName, config.Environment().KeyVaultDNSSuffix),
					"SequenceVersion":           uuid.NewV4().String(),
					"VolumeType":                "ALL",
				},
				Type:               to.StringPtr("AzureDiskEncryptionForLinux"),
				TypeHandlerVersion: to.StringPtr("0.1"),
			},
		})
	if err != nil {
		return ext, fmt.Errorf("cannot create vm extension: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, extensionsClient.Client)
	if err != nil {
		return ext, fmt.Errorf("cannot get the extension create or update future response: %v", err)
	}

	return future.Result(extensionsClient)
}
