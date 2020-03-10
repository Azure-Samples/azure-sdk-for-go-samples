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
)

func getAvailabilitySetsClient() compute.AvailabilitySetsClient {
	asClient := compute.NewAvailabilitySetsClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	asClient.Authorizer = a
	asClient.AddToUserAgent(config.UserAgent())
	return asClient
}

// CreateAvailabilitySet creates an availability set
func CreateAvailabilitySet(ctx context.Context, asName string) (compute.AvailabilitySet, error) {
	asClient := getAvailabilitySetsClient()
	return asClient.CreateOrUpdate(ctx,
		config.GroupName(),
		asName,
		compute.AvailabilitySet{
			Location: to.StringPtr(config.Location()),
			AvailabilitySetProperties: &compute.AvailabilitySetProperties{
				PlatformFaultDomainCount:  to.Int32Ptr(1),
				PlatformUpdateDomainCount: to.Int32Ptr(1),
			},
			Sku: &compute.Sku{
				Name: to.StringPtr("Aligned"),
			},
		})
}

// GetAvailabilitySet gets info on an availability set
func GetAvailabilitySet(ctx context.Context, asName string) (compute.AvailabilitySet, error) {
	asClient := getAvailabilitySetsClient()
	return asClient.Get(ctx, config.GroupName(), asName)
}

// CreateVMWithLoadBalancer creates a new VM in an availability set. It also
// creates and configures a load balancer and associates that with the VM's
// NIC.
func CreateVMWithLoadBalancer(ctx context.Context, vmName, lbName, vnetName, subnetName, publicipName, availabilitySetName string, natRule int) (vm compute.VirtualMachine, err error) {
	nicName := fmt.Sprintf("nic-%s", vmName)

	_, err = network.CreateNICWithLoadBalancer(ctx, lbName, vnetName, subnetName, nicName, natRule)
	if err != nil {
		return
	}
	nic, err := network.GetNic(ctx, nicName)
	if err != nil {
		return
	}

	as, err := GetAvailabilitySet(ctx, availabilitySetName)
	if err != nil {
		return
	}

	vmClient := getVMClient()
	future, err := vmClient.CreateOrUpdate(
		ctx,
		config.GroupName(),
		vmName,
		compute.VirtualMachine{
			Location: to.StringPtr(config.Location()),
			VirtualMachineProperties: &compute.VirtualMachineProperties{
				HardwareProfile: &compute.HardwareProfile{
					VMSize: compute.VirtualMachineSizeTypesStandardA0,
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
					AdminUsername: to.StringPtr("azureuser"),
					AdminPassword: to.StringPtr("password!1delete"),
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
				AvailabilitySet: &compute.SubResource{
					ID: as.ID,
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
