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

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/network"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

func GetVMSSClient() compute.VirtualMachineScaleSetsClient {
	vmssClient := compute.NewVirtualMachineScaleSetsClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	vmssClient.Authorizer = a
	vmssClient.AddToUserAgent(config.UserAgent())
	return vmssClient
}

func GetVMSSExtensionsClient() compute.VirtualMachineScaleSetExtensionsClient {
	extClient := compute.NewVirtualMachineScaleSetExtensionsClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	extClient.Authorizer = a
	extClient.AddToUserAgent(config.UserAgent())
	return extClient
}

// CreateVMSS creates a new virtual machine scale set with the specified name using the specified vnet and subnet.
// Username, password, and sshPublicKeyPath determine logon credentials.
func CreateVMSS(ctx context.Context, vmssName, vnetName, subnetName, username, password, sshPublicKeyPath string) (vmss compute.VirtualMachineScaleSet, err error) {
	// see the network samples for how to create and get a subnet resource
	subnet, _ := network.GetVirtualNetworkSubnet(ctx, vnetName, subnetName)

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

	vmssClient := GetVMSSClient()
	future, err := vmssClient.CreateOrUpdate(
		ctx,
		config.GroupName(),
		vmssName,
		compute.VirtualMachineScaleSet{
			Location: to.StringPtr(config.DefaultLocation()),
			Sku: &compute.Sku{
				Name:     to.StringPtr(string(compute.VirtualMachineSizeTypesBasicA0)),
				Capacity: to.Int64Ptr(1),
			},
			VirtualMachineScaleSetProperties: &compute.VirtualMachineScaleSetProperties{
				Overprovision: to.BoolPtr(false),
				UpgradePolicy: &compute.UpgradePolicy{
					Mode: compute.Manual,
					AutomaticOSUpgradePolicy: &compute.AutomaticOSUpgradePolicy{
						EnableAutomaticOSUpgrade: to.BoolPtr(false),
						DisableAutomaticRollback: to.BoolPtr(false),
					},
				},
				VirtualMachineProfile: &compute.VirtualMachineScaleSetVMProfile{
					OsProfile: &compute.VirtualMachineScaleSetOSProfile{
						ComputerNamePrefix: to.StringPtr(vmssName),
						AdminUsername:      to.StringPtr(username),
						AdminPassword:      to.StringPtr(password),
						LinuxConfiguration: &compute.LinuxConfiguration{
							SSH: &compute.SSHConfiguration{
								PublicKeys: &[]compute.SSHPublicKey{
									{
										Path: to.StringPtr(
											fmt.Sprintf("/home/%s/.ssh/authorized_keys",
												username)),
										KeyData: to.StringPtr(sshKeyData),
									},
								},
							},
						},
					},
					StorageProfile: &compute.VirtualMachineScaleSetStorageProfile{
						ImageReference: &compute.ImageReference{
							Offer:     to.StringPtr(offer),
							Publisher: to.StringPtr(publisher),
							Sku:       to.StringPtr(sku),
							Version:   to.StringPtr("latest"),
						},
					},
					NetworkProfile: &compute.VirtualMachineScaleSetNetworkProfile{
						NetworkInterfaceConfigurations: &[]compute.VirtualMachineScaleSetNetworkConfiguration{
							{
								Name: to.StringPtr(vmssName),
								VirtualMachineScaleSetNetworkConfigurationProperties: &compute.VirtualMachineScaleSetNetworkConfigurationProperties{
									Primary:            to.BoolPtr(true),
									EnableIPForwarding: to.BoolPtr(true),
									IPConfigurations: &[]compute.VirtualMachineScaleSetIPConfiguration{
										{
											Name: to.StringPtr(vmssName),
											VirtualMachineScaleSetIPConfigurationProperties: &compute.VirtualMachineScaleSetIPConfigurationProperties{
												Subnet: &compute.APIEntityReference{
													ID: subnet.ID,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	)
	if err != nil {
		return vmss, fmt.Errorf("cannot create vmss: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, vmssClient.Client)
	if err != nil {
		return vmss, fmt.Errorf("cannot get the vmss create or update future response: %v", err)
	}

	return future.Result(vmssClient)
}

// GetVMSS gets the specified VMSS info
func GetVMSS(ctx context.Context, vmssName string) (compute.VirtualMachineScaleSet, error) {
	vmssClient := GetVMSSClient()
	return vmssClient.Get(ctx, config.GroupName(), vmssName)
}

// UpdateVMSS modifies the VMSS resource by getting it, updating it locally, and
// putting it back to the server.
func UpdateVMSS(ctx context.Context, vmssName string, tags map[string]*string) (vmss compute.VirtualMachineScaleSet, err error) {
	// get the VMSS resource
	vmss, err = GetVMSS(ctx, vmssName)
	if err != nil {
		return
	}

	// update it
	vmss.Tags = tags

	// PUT it back
	vmssClient := GetVMSSClient()
	future, err := vmssClient.CreateOrUpdate(ctx, config.GroupName(), vmssName, vmss)
	if err != nil {
		return vmss, fmt.Errorf("cannot update vmss: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, vmssClient.Client)
	if err != nil {
		return vmss, fmt.Errorf("cannot get the vmss create or update future response: %v", err)
	}

	return future.Result(vmssClient)
}

// DeallocateVMSS deallocates the selected VMSS
func DeallocateVMSS(ctx context.Context, vmssName string) (osr autorest.Response, err error) {
	vmssClient := GetVMSSClient()
	// passing nil instance ids will deallocate all VMs in the VMSS
	future, err := vmssClient.Deallocate(ctx, config.GroupName(), vmssName, nil)
	if err != nil {
		return osr, fmt.Errorf("cannot deallocate vmss: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, vmssClient.Client)
	if err != nil {
		return osr, fmt.Errorf("cannot get the vmss deallocate future response: %v", err)
	}

	return future.Result(vmssClient)
}

// StartVMSS starts the selected VMSS
func StartVMSS(ctx context.Context, vmssName string) (osr autorest.Response, err error) {
	vmssClient := GetVMSSClient()
	// passing nil instance ids will start all VMs in the VMSS
	future, err := vmssClient.Start(ctx, config.GroupName(), vmssName, nil)
	if err != nil {
		return osr, fmt.Errorf("cannot start vmss: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, vmssClient.Client)
	if err != nil {
		return osr, fmt.Errorf("cannot get the vmss start future response: %v", err)
	}

	return future.Result(vmssClient)
}

// RestartVMSS restarts the selected VMSS
func RestartVMSS(ctx context.Context, vmssName string) (osr autorest.Response, err error) {
	vmssClient := GetVMSSClient()
	// passing nil instance ids will restart all VMs in the VMSS
	future, err := vmssClient.Restart(ctx, config.GroupName(), vmssName, nil)
	if err != nil {
		return osr, fmt.Errorf("cannot restart vm: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, vmssClient.Client)
	if err != nil {
		return osr, fmt.Errorf("cannot get the vm restart future response: %v", err)
	}

	return future.Result(vmssClient)
}

// StopVMSS stops the selected VMSS
func StopVMSS(ctx context.Context, vmssName string) (osr autorest.Response, err error) {
	vmssClient := GetVMSSClient()
	// passing nil instance ids will stop all VMs in the VMSS
	future, err := vmssClient.PowerOff(ctx, config.GroupName(), vmssName, nil, nil)
	if err != nil {
		return osr, fmt.Errorf("cannot power off vmss: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, vmssClient.Client)
	if err != nil {
		return osr, fmt.Errorf("cannot get the vmss power off future response: %v", err)
	}

	return future.Result(vmssClient)
}
