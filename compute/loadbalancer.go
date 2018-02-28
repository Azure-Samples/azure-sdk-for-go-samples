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
		})
}

func CreateVMWithLoadBalancer(ctx context.Context, vmName, lbName, vnetName, subnetName, pipName string, natRule int) (vm compute.VirtualMachine, err error) {
	nicName := fmt.Sprintf("nic-%s", vmName)

	nic, err := network.CreateNICWithLoadBalancer(ctx, lbName, vnetName, subnetName, nicName, natRule)
	if err != nil {
		return vm, err
	}
}
