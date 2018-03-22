// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package network

import (
	"context"
	"fmt"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2017-09-01/network"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

// VNet Subnets

func getSubnetsClient() network.SubnetsClient {
	token, _ := iam.GetResourceManagementToken(iam.AuthGrantType())
	subnetsClient := network.NewSubnetsClient(helpers.SubscriptionID())
	subnetsClient.Authorizer = autorest.NewBearerAuthorizer(token)
	subnetsClient.AddToUserAgent(helpers.UserAgent())
	return subnetsClient
}

// CreateVirtualNetworkSubnet creates a subnet in an existing vnet
func CreateVirtualNetworkSubnet(ctx context.Context, vnetName, subnetName string) (subnet network.Subnet, err error) {
	subnetsClient := getSubnetsClient()

	future, err := subnetsClient.CreateOrUpdate(
		ctx,
		helpers.ResourceGroupName(),
		vnetName,
		subnetName,
		network.Subnet{
			SubnetPropertiesFormat: &network.SubnetPropertiesFormat{
				AddressPrefix: to.StringPtr("10.0.0.0/16"),
			},
		})
	if err != nil {
		return subnet, fmt.Errorf("cannot create subnet: %v", err)
	}

	err = future.WaitForCompletion(ctx, subnetsClient.Client)
	if err != nil {
		return subnet, fmt.Errorf("cannot get the subnet create or update future response: %v", err)
	}

	return future.Result(subnetsClient)
}

// CreateSubnetWithNetworkSecurityGroup create a subnet referencing a network security group
func CreateSubnetWithNetworkSecurityGroup(ctx context.Context, vnetName, subnetName, addressPrefix, nsgName string) (subnet network.Subnet, err error) {
	nsg, err := GetNetworkSecurityGroup(ctx, nsgName)
	if err != nil {
		return subnet, fmt.Errorf("cannot get nsg: %v", err)
	}

	subnetsClient := getSubnetsClient()
	future, err := subnetsClient.CreateOrUpdate(
		ctx,
		helpers.ResourceGroupName(),
		vnetName,
		subnetName,
		network.Subnet{
			SubnetPropertiesFormat: &network.SubnetPropertiesFormat{
				AddressPrefix:        to.StringPtr(addressPrefix),
				NetworkSecurityGroup: &nsg,
			},
		})
	if err != nil {
		return subnet, fmt.Errorf("cannot create subnet: %v", err)
	}

	err = future.WaitForCompletion(ctx, subnetsClient.Client)
	if err != nil {
		return subnet, fmt.Errorf("cannot get the subnet create or update future response: %v", err)
	}

	return future.Result(subnetsClient)
}

// DeleteVirtualNetworkSubnet deletes a subnet
func DeleteVirtualNetworkSubnet() {}

// GetVirtualNetworkSubnet returns an existing subnet from a virtual network
func GetVirtualNetworkSubnet(ctx context.Context, vnetName string, subnetName string) (network.Subnet, error) {
	subnetsClient := getSubnetsClient()
	return subnetsClient.Get(ctx, helpers.ResourceGroupName(), vnetName, subnetName, "")
}
