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
	publisher = "Canonical"
	offer     = "UbuntuServer"
	sku       = "16.04-LTS"
)

var (
	location = "local"
)

// fakepubkey is used if a key isn't available at the specified path in CreateVM(...)
var fakepubkey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC7laRyN4B3YZmVrDEZLZoIuUA72pQ0DpGuZBZWykCofIfCPrFZAJgFvonKGgKJl6FGKIunkZL9Us/mV4ZPkZhBlE7uX83AAf5i9Q8FmKpotzmaxN10/1mcnEE7pFvLoSkwqrQSkrrgSm8zaJ3g91giXSbtqvSIj/vk2f05stYmLfhAwNo3Oh27ugCakCoVeuCrZkvHMaJgcYrIGCuFo6q0Pfk9rsZyriIqEa9AtiUOtViInVYdby7y71wcbl0AbbCZsTSqnSoVxm2tRkOsXV6+8X4SnwcmZbao3H+zfO1GBhQOLxJ4NQbzAa8IJh810rYARNLptgmsd4cYXVOSosTX azureuser"

func getVMClient() (compute.VirtualMachinesClient, error) {
	token, err := iam.GetResourceManagementTokenHybrid(helpers.ActiveDirectoryEndpoint(), helpers.TenantID(), helpers.ClientID(), helpers.ClientSecret(), helpers.ActiveDirectoryResourceID())
	if err != nil {
		log.Fatal(fmt.Sprintf("Cannot generate token. Error details: %s.", err.Error()))
	}
	vmClient := compute.NewVirtualMachinesClientWithBaseURI(helpers.ArmEndpoint(), helpers.SubscriptionID())
	vmClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return vmClient, nil
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
			log.Fatalf("failed to read SSH key data: %v", err)
		}
		sshKeyData = string(sshBytes)
	} else {
		sshKeyData = fakepubkey
	}
	vhdURItemplate := "https://%s.blob." + helpers.StorageEndpointSuffix() + "/vhds/%s.vhd"
	vmClient, _ := getVMClient()
	future, err := vmClient.CreateOrUpdate(
		cntx,
		helpers.ResourceGroupName(),
		vmName,
		compute.VirtualMachine{
			Location: to.StringPtr(location),
			VirtualMachineProperties: &compute.VirtualMachineProperties{
				HardwareProfile: &compute.HardwareProfile{
					VMSize: compute.StandardA1,
				},
				StorageProfile: &compute.StorageProfile{
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
	err = future.WaitForCompletion(cntx, vmClient.Client)
	if err != nil {
		return vm, fmt.Errorf("cannot get the vm create or update future response: %v", err)
	}
	return future.Result(vmClient)
}
