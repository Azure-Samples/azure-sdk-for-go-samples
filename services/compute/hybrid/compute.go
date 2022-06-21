// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package compute

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/internal/iam"
	hybridnetwork "github.com/Azure-Samples/azure-sdk-for-go-samples/services/network/hybrid"
	hybridcompute "github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/compute/mgmt/compute"
	"github.com/Azure/go-autorest/autorest"
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
	vmClient := hybridcompute.NewVirtualMachinesClientWithBaseURI(
		config.Environment().ResourceManagerEndpoint, config.SubscriptionID())
	vmClient.Authorizer = autorest.NewBearerAuthorizer(token)
	vmClient.AddToUserAgent(config.UserAgent())
	return vmClient
}

// CreateVM creates a new virtual machine with the specified name using the specified network interface and storage account.
// Username, password, and sshPublicKeyPath determine logon credentials.
func CreateVM(ctx context.Context, vmName, nicName, username, password, storageAccountName, sshPublicKeyPath string) (vm hybridcompute.VirtualMachine, err error) {
	nic, _ := hybridnetwork.GetNic(ctx, nicName)
	environment := config.Environment()
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
		Location: to.StringPtr(config.Location()),
		VirtualMachineProperties: &hybridcompute.VirtualMachineProperties{
			HardwareProfile: hardwareProfile,
			StorageProfile:  storageProfile,
			OsProfile:       osProfile,
			NetworkProfile:  networkProfile,
		},
	}
	future, err := vmClient.CreateOrUpdate(
		ctx,
		config.GroupName(),
		vmName,
		virtualMachine,
	)
	if err != nil {
		return vm, fmt.Errorf(fmt.Sprintf(errorPrefix, err))
	}
	err = future.WaitForCompletionRef(ctx, vmClient.Client)
	if err != nil {
		return vm, fmt.Errorf(fmt.Sprintf(errorPrefix, err))
	}
	return future.Result(vmClient)
}
