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
	hybridnetwork "github.com/Azure-Samples/azure-sdk-for-go-samples/network/hybrid"
	hybridcompute "github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/compute/mgmt/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
)

const (
	publisher   = "Canonical"
	offer       = "UbuntuServer"
	sku         = "16.04-LTS"
	errorPrefix = "Cannot create VM, reason: %v"
)

func getVMClient(activeDirectoryEndpoint, tokenAudience string) hybridcompute.VirtualMachinesClient {
	token, err := iam.GetResourceManagementTokenHybrid(activeDirectoryEndpoint, tokenAudience)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}
	vmClient := hybridcompute.NewVirtualMachinesClientWithBaseURI(helpers.ArmEndpoint(), helpers.SubscriptionID())
	vmClient.Authorizer = autorest.NewBearerAuthorizer(token)
	vmClient.AddToUserAgent(helpers.UserAgent())
	return vmClient
}

// CreateVM creates a new virtual machine with the specified name using the specified network interface and storage account.
// Username, password, and sshPublicKeyPath determine logon credentials.
func CreateVM(ctx context.Context, vmName, nicName, username, password, storageAccountName, sshPublicKeyPath string) (vm hybridcompute.VirtualMachine, err error) {
	cntx := context.Background()
	nic, _ := hybridnetwork.GetNic(cntx, nicName)
	environment, _ := azure.EnvironmentFromURL(helpers.ArmEndpoint())
	vhdURItemplate := "https://%s.blob." + environment.StorageEndpointSuffix + "/vhds/%s.vhd"

	vmClient := getVMClient(environment.ActiveDirectoryEndpoint, environment.TokenAudience)
	hardwareProfile := &hybridcompute.HardwareProfile{
		VMSize: hybridcompute.StandardA1,
	}
	storageProfile := &hybridcompute.StorageProfile{
		ImageReference: &hybridcompute.ImageReference{
			Publisher: to.StringPtr(publisher),
			Offer:     to.StringPtr(offer),
			Sku:       to.StringPtr(sku),
			Version:   to.StringPtr("latest"),
		},
		OsDisk: &hybridcompute.OSDisk{
			Name: to.StringPtr("osDisk"),
			Vhd: &hybridcompute.VirtualHardDisk{
				URI: to.StringPtr(fmt.Sprintf(vhdURItemplate, storageAccountName, vmName)),
			},
			CreateOption: hybridcompute.FromImage,
		},
	}
	osProfile := &hybridcompute.OSProfile{
		ComputerName:  to.StringPtr(vmName),
		AdminUsername: to.StringPtr(username),
		AdminPassword: to.StringPtr(password),
	}

	_, err = os.Stat(sshPublicKeyPath)
	if err == nil {
		sshBytes, err := ioutil.ReadFile(sshPublicKeyPath)
		if err != nil {
			log.Fatalf(fmt.Sprintf(errorPrefix, fmt.Sprintf("failed to read SSH key data: %v", err)))
		}

		// if a key is available at the specified path then populate LinuxConfiguration
		osProfile.LinuxConfiguration = &hybridcompute.LinuxConfiguration{
			SSH: &hybridcompute.SSHConfiguration{
				PublicKeys: &[]hybridcompute.SSHPublicKey{
					{
						Path:    to.StringPtr(fmt.Sprintf("/home/%s/.ssh/authorized_keys", username)),
						KeyData: to.StringPtr(string(sshBytes)),
					},
				},
			},
		}
	}

	networkProfile := &hybridcompute.NetworkProfile{
		NetworkInterfaces: &[]hybridcompute.NetworkInterfaceReference{
			{
				ID: nic.ID,
				NetworkInterfaceReferenceProperties: &hybridcompute.NetworkInterfaceReferenceProperties{
					Primary: to.BoolPtr(true),
				},
			},
		},
	}
	virtualMachine := hybridcompute.VirtualMachine{
		Location: to.StringPtr(helpers.Location()),
		VirtualMachineProperties: &hybridcompute.VirtualMachineProperties{
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
