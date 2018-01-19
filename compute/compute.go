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

func getVMClient() (compute.VirtualMachinesClient, error) {
	token, _ := iam.GetResourceManagementToken(iam.AuthGrantType())
	vmClient := compute.NewVirtualMachinesClient(helpers.SubscriptionID())
	vmClient.Authorizer = autorest.NewBearerAuthorizer(token)
	vmClient.AddToUserAgent(helpers.UserAgent())
	return vmClient, nil
}

func getExtensionClient() (compute.VirtualMachineExtensionsClient, error) {
	token, _ := iam.GetResourceManagementToken(iam.AuthGrantType())
	extClient := compute.NewVirtualMachineExtensionsClient(helpers.SubscriptionID())
	extClient.Authorizer = autorest.NewBearerAuthorizer(token)
	extClient.AddToUserAgent(helpers.UserAgent())
	return extClient, nil
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

	vmClient, _ := getVMClient()
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

// CreateVMForMSI creates a virtual machine with a systems assigned identity type
func CreateVMForMSI(ctx context.Context, vmName, nicName, username, password string) (vm compute.VirtualMachine, err error) {
	nic, _ := network.GetNic(ctx, nicName)

	vmClient, _ := getVMClient()
	future, err := vmClient.CreateOrUpdate(
		ctx,
		helpers.ResourceGroupName(),
		vmName,
		compute.VirtualMachine{
			Location: to.StringPtr(helpers.Location()),
			Identity: &compute.VirtualMachineIdentity{
				Type: compute.SystemAssigned,
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
	extClient, _ := getExtensionClient()

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

// GetVM gets the specified VM info
func GetVM(ctx context.Context, vmName string) (compute.VirtualMachine, error) {
	vmClient, _ := getVMClient()
	return vmClient.Get(ctx, helpers.ResourceGroupName(), vmName, compute.InstanceView)
}
