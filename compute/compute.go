// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package compute

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/network"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/storage"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2017-03-30/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/satori/uuid"
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

// CreateManagedDisk creates an empty 64 GB managed disk
func CreateManagedDisk(ctx context.Context, diskName string) (disk compute.Disk, err error) {
	disksClient := getDisksClient()
	future, err := disksClient.CreateOrUpdate(
		ctx,
		helpers.ResourceGroupName(),
		diskName,
		compute.Disk{
			Location: to.StringPtr(helpers.Location()),
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

func GetDisk(ctx context.Context, diskName string) (disk compute.Disk, err error) {
	disksClient := getDisksClient()
	return disksClient.Get(ctx, helpers.ResourceGroupName(), diskName)
}

func CreateVMWithManagedDisk(ctx context.Context, nicName, diskName, storageAccountName, vmName string) (vm compute.VirtualMachine, err error) {
	nic, _ := network.GetNic(ctx, nicName)
	storageAccount, _ := storage.GetStorageAccount(ctx, storageAccountName)
	disk, _ := GetDisk(ctx, diskName)

	var storageURI *string
	if storageAccount.PrimaryEndpoints != nil {
		storageURI = (*storageAccount.PrimaryEndpoints).Blob
	} else {
		err = errors.New("No storage endpoint found")
		return
	}

	vmClient := getVMClient()
	future, err := vmClient.CreateOrUpdate(
		ctx,
		helpers.ResourceGroupName(),
		vmName, compute.VirtualMachine{
			Location: to.StringPtr(helpers.Location()),
			VirtualMachineProperties: &compute.VirtualMachineProperties{
				DiagnosticsProfile: &compute.DiagnosticsProfile{
					BootDiagnostics: &compute.BootDiagnostics{
						Enabled:    to.BoolPtr(true),
						StorageURI: storageURI,
					},
				},
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
					AdminUsername: to.StringPtr("sampleuser"),
					AdminPassword: to.StringPtr("azureRocksWithGo!"),
					LinuxConfiguration: &compute.LinuxConfiguration{
						DisablePasswordAuthentication: to.BoolPtr(false),
					},
				},
				StorageProfile: &compute.StorageProfile{
					ImageReference: &compute.ImageReference{
						Publisher: to.StringPtr("Canonical"),
						Offer:     to.StringPtr("UbuntuServer"),
						Sku:       to.StringPtr("14.04.5-LTS"),
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
								ID:                 disk.ID,
								StorageAccountType: compute.StorageAccountTypes(storageAccount.Sku.Name),
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
		helpers.ResourceGroupName(),
		vmName,
		"AzureDiskEncryptionForLinux",
		compute.VirtualMachineExtension{
			Location: to.StringPtr(helpers.Location()),
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
