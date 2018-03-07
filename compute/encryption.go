// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package compute

import (
	"context"
	"fmt"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/network"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2017-03-30/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/satori/go.uuid"
)

// CreateManagedDisk creates an empty, 64GB disk. The disk can be attached to a VM later.
func CreateManagedDisk(ctx context.Context, diskName string) (disk compute.Disk, err error) {
	disksClient := getDisksClient()
	future, err := disksClient.CreateOrUpdate(
		ctx,
		internal.ResourceGroupName(),
		diskName,
		compute.Disk{
			Location: to.StringPtr(internal.Location()),
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

	err = future.WaitForCompletion(ctx, disksClient.Client)
	if err != nil {
		return disk, fmt.Errorf("cannot get the disk create or update future response: %v", err)
	}

	return future.Result(disksClient)
}

// CreateVMWithManagedDisk creates a VM, attaching an already existing data disk
func CreateVMWithManagedDisk(ctx context.Context, nicName, diskName, vmName string, username string, password string) (vm compute.VirtualMachine, err error) {
	nic, _ := network.GetNic(ctx, nicName)
	disk, _ := getDisk(ctx, diskName)

	vmClient := getVMClient()
	future, err := vmClient.CreateOrUpdate(
		ctx,
		internal.ResourceGroupName(),
		vmName, compute.VirtualMachine{
			Location: to.StringPtr(internal.Location()),
			VirtualMachineProperties: &compute.VirtualMachineProperties{
				HardwareProfile: &compute.HardwareProfile{
					VMSize: compute.StandardDS2V2,
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

	err = future.WaitForCompletion(ctx, vmClient.Client)
	if err != nil {
		return vm, fmt.Errorf("cannot get the vm create or update future response: %v", err)
	}

	return future.Result(vmClient)
}

// AddEncyptionExtension add the disk encryption extension to the selected VM
func AddEncyptionExtension(ctx context.Context, vmName, vaultName, keyID string) (ext compute.VirtualMachineExtension, err error) {
	extClient := getExtensionClient()
	future, err := extClient.CreateOrUpdate(
		ctx,
		internal.ResourceGroupName(),
		vmName,
		"AzureDiskEncryptionForLinux",
		compute.VirtualMachineExtension{
			Location: to.StringPtr(internal.Location()),
			VirtualMachineExtensionProperties: &compute.VirtualMachineExtensionProperties{
				AutoUpgradeMinorVersion: to.BoolPtr(true),
				ProtectedSettings: &map[string]interface{}{
					"AADClientSecret": iam.ClientSecret(), // The Secret that was created for the service principal secret.
					"Passphrase":      "yourPassPhrase",   // This sample uses a simple passphrase, but you should absolutely use something more sophisticated.
				},
				Publisher: to.StringPtr("Microsoft.Azure.Security"),
				Settings: &map[string]interface{}{
					"AADClientID":               iam.ClientID(),
					"EncryptionOperation":       "EnableEncryption",
					"KeyEncryptionAlgorithm":    "RSA-OAEP",
					"KeyEncryptionKeyAlgorithm": keyID,
					"KeyVaultURL":               fmt.Sprintf("https://%s.vault.azure.net/", vaultName),
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

	err = future.WaitForCompletion(ctx, extClient.Client)
	if err != nil {
		return ext, fmt.Errorf("cannot get the extension create or update future response: %v", err)
	}

	return future.Result(extClient)
}
