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

// Public IP Addresses

func getIPClient() network.PublicIPAddressesClient {
	token, _ := iam.GetResourceManagementToken(iam.AuthGrantType())
	ipClient := network.NewPublicIPAddressesClient(helpers.SubscriptionID())
	ipClient.Authorizer = autorest.NewBearerAuthorizer(token)
	ipClient.AddToUserAgent(helpers.UserAgent())
	return ipClient
}

// CreatePublicIP creates a new public IP
func CreatePublicIP(ctx context.Context, ipName string) (ip network.PublicIPAddress, err error) {
	ipClient := getIPClient()
	future, err := ipClient.CreateOrUpdate(
		ctx,
		helpers.ResourceGroupName(),
		ipName,
		network.PublicIPAddress{
			Name:     to.StringPtr(ipName),
			Location: to.StringPtr(helpers.Location()),
			PublicIPAddressPropertiesFormat: &network.PublicIPAddressPropertiesFormat{
				PublicIPAddressVersion:   network.IPv4,
				PublicIPAllocationMethod: network.Static,
			},
		},
	)

	if err != nil {
		return ip, fmt.Errorf("cannot create public ip address: %v", err)
	}

	err = future.WaitForCompletion(ctx, ipClient.Client)
	if err != nil {
		return ip, fmt.Errorf("cannot get public ip address create or update future response: %v", err)
	}

	return future.Result(ipClient)
}

// GetPublicIP returns an existing public IP
func GetPublicIP(ctx context.Context, ipName string) (network.PublicIPAddress, error) {
	ipClient := getIPClient()
	return ipClient.Get(ctx, helpers.ResourceGroupName(), ipName, "")
}

// DeletePublicIP deletes an existing public IP
func DeletePublicIP(ctx context.Context, ipName string) (result network.PublicIPAddressesDeleteFuture, err error) {
	ipClient := getIPClient()
	return ipClient.Delete(ctx, helpers.ResourceGroupName(), ipName)
}
