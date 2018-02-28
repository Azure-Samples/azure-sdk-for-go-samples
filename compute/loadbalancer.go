// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package compute

import (
	"context"
	"fmt"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/network"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2017-03-30/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

func getAvailabilitySetsClient() compute.AvailabilitySetsClient {
	token, _ := iam.GetResourceManagementToken(iam.AuthGrantType())
	avaSetClient := compute.NewAvailabilitySetsClient(helpers.SubscriptionID())
	avaSetClient.Authorizer = autorest.NewBearerAuthorizer(token)
	avaSetClient.AddToUserAgent(helpers.UserAgent())
	return avaSetClient
}

func CreateAvailabilitySet(ctx context.Context, avaSetName string) (compute.AvailabilitySet, error) {
	avaSetClient := getAvailabilitySetsClient()
	return avaSetClient.CreateOrUpdate(ctx,
		helpers.ResourceGroupName(),
		avaSetName,
		compute.AvailabilitySet{
			Location: to.StringPtr(helpers.Location()),
			AvailabilitySetProperties: &compute.AvailabilitySetProperties{
				PlatformFaultDomainCount:  to.Int32Ptr(2),
				PlatformUpdateDomainCount: to.Int32Ptr(2),
			},
			Sku: &compute.Sku{
				Name: to.StringPtr("Aligned"),
			},
		})
}

func GetAvailabilitySet(ctx context.Context, avaSetName string) (compute.AvailabilitySet, error) {
	avaSetClient := getAvailabilitySetsClient()
	return avaSetClient.Get(ctx, helpers.ResourceGroupName(), avaSetName)
}

func CreateVMWithLoadBalancer(ctx context.Context, vmName, lbName, vnetName, subnetName, pipName, availabilySetName string, natRule int) (vm compute.VirtualMachine, err error) {
	nicName := fmt.Sprintf("nic-%s", vmName)

	_, err = network.CreateNICWithLoadBalancer(ctx, lbName, vnetName, subnetName, nicName, natRule)
	if err != nil {
		return
	}
	nic, err := network.GetNic(ctx, nicName)
	if err != nil {
		return
	}

	avaSet, err := GetAvailabilitySet(ctx, availabilySetName)
	if err != nil {
		return
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
					ID: avaSet.ID,
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
