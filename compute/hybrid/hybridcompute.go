// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package hybridcompute

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	hybridnetwork "github.com/Azure-Samples/azure-sdk-for-go-samples/network/hybrid"
	"github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/compute/mgmt/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

const (
	publisher   = "Canonical"
	offer       = "UbuntuServer"
	sku         = "16.04-LTS"
	errorPrefix = "Cannot create VM, reason: %v"
)

// fakepubkey is used if a key isn't available at the specified path in CreateVM(...)
var fakepubkey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC7laRyN4B3YZmVrDEZLZoIuUA72pQ0DpGuZBZWykCofIfCPrFZAJgFvonKGgKJl6FGKIunkZL9Us/mV4ZPkZhBlE7uX83AAf5i9Q8FmKpotzmaxN10/1mcnEE7pFvLoSkwqrQSkrrgSm8zaJ3g91giXSbtqvSIj/vk2f05stYmLfhAwNo3Oh27ugCakCoVeuCrZkvHMaJgcYrIGCuFo6q0Pfk9rsZyriIqEa9AtiUOtViInVYdby7y71wcbl0AbbCZsTSqnSoVxm2tRkOsXV6+8X4SnwcmZbao3H+zfO1GBhQOLxJ4NQbzAa8IJh810rYARNLptgmsd4cYXVOSosTX azureuser"

func getVMClient() compute.VirtualMachinesClient {
	token, err := iam.GetResourceManagementTokenHybrid(helpers.ActiveDirectoryEndpoint(), helpers.TenantID(), helpers.ClientID(), helpers.ClientSecret(), helpers.ActiveDirectoryResourceID())
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}
	vmClient := compute.NewVirtualMachinesClientWithBaseURI(helpers.ArmEndpoint(), helpers.SubscriptionID())
	vmClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return vmClient
}

// CreateVM creates a new virtual machine with the specified name using the specified network interface and storage account.
// Username, password, and sshPublicKeyPath determine logon credentials.
func CreateVM(ctx context.Context, vmName, nicName, username, password, storageAccountName, sshPublicKeyPath string) (vm compute.VirtualMachine, err error) {
	cntx := context.Background()
	nic, _ := hybridnetwork.GetNic(cntx, nicName)

	var sshKeyData string
	_, err = os.Stat(sshPublicKeyPath)
	if err == nil {
		sshBytes, err := ioutil.ReadFile(sshPublicKeyPath)
		if err != nil {
			log.Fatalf(fmt.Sprintf(errorPrefix, fmt.Sprintf("failed to read SSH key data: %v", err)))
		}
		sshKeyData = string(sshBytes)
	} else {
		sshKeyData = fakepubkey
	}
	vhdURItemplate := "https://%s.blob." + helpers.StorageEndpointSuffix() + "/vhds/%s.vhd"
	vmClient := getVMClient()
	hardwareProfile := &compute.HardwareProfile{
		VMSize: compute.StandardA1,
	}
	storageProfile := &compute.StorageProfile{
		ImageReference: &compute.ImageReference{
			Publisher: to.StringPtr(publisher),
			Offer:     to.StringPtr(offer),
			Sku:       to.StringPtr(sku),
			Version:   to.StringPtr("latest"),
		},
		OsDisk: &compute.OSDisk{
			Name: to.StringPtr("osDisk"),
			Vhd: &compute.VirtualHardDisk{
				URI: to.StringPtr(fmt.Sprintf(vhdURItemplate, storageAccountName, vmName)),
			},
			CreateOption: compute.FromImage,
		},
	}
	osProfile := &compute.OSProfile{
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
	}
	networkProfile := &compute.NetworkProfile{
		NetworkInterfaces: &[]compute.NetworkInterfaceReference{
			{
				ID: nic.ID,
				NetworkInterfaceReferenceProperties: &compute.NetworkInterfaceReferenceProperties{
					Primary: to.BoolPtr(true),
				},
			},
		},
	}
	virtualMachine := compute.VirtualMachine{
		Location: to.StringPtr(helpers.Location()),
		VirtualMachineProperties: &compute.VirtualMachineProperties{
			HardwareProfile: hardwareProfile,
			StorageProfile:  storageProfile,
			OsProfile:       osProfile,
			NetworkProfile:  networkProfile,
		},
	}
	future, err := vmClient.CreateOrUpdate(
		cntx,
		helpers.ResourceGroupName(),
		vmName,
		virtualMachine,
	)
	if err != nil {
		return vm, fmt.Errorf(fmt.Sprintf(errorPrefix, err))
	}
	err = future.WaitForCompletion(cntx, vmClient.Client)
	if err != nil {
		return vm, fmt.Errorf(fmt.Sprintf(errorPrefix, err))
	}
	return future.Result(vmClient)
}
