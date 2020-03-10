// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package network

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
	"github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/network/mgmt/network"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
)

const (
	errorPrefix = "Cannot create %v, reason: %v"
)

func getVnetClient(activeDirectoryEndpoint, tokenAudience string) network.VirtualNetworksClient {
	token, err := iam.GetResourceManagementTokenHybrid(activeDirectoryEndpoint, tokenAudience)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, "virtual network", fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}
	vnetClient := network.NewVirtualNetworksClientWithBaseURI(config.Environment().ResourceManagerEndpoint, config.SubscriptionID())
	vnetClient.Authorizer = autorest.NewBearerAuthorizer(token)
	vnetClient.AddToUserAgent(config.UserAgent())
	return vnetClient
}

func getNsgClient(activeDirectoryEndpoint, tokenAudience string) network.SecurityGroupsClient {
	token, err := iam.GetResourceManagementTokenHybrid(activeDirectoryEndpoint, tokenAudience)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, "security group", fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}
	nsgClient := network.NewSecurityGroupsClientWithBaseURI(config.Environment().ResourceManagerEndpoint, config.SubscriptionID())
	nsgClient.Authorizer = autorest.NewBearerAuthorizer(token)
	nsgClient.AddToUserAgent(config.UserAgent())
	return nsgClient
}

func getIPClient(activeDirectoryEndpoint, tokenAudience string) network.PublicIPAddressesClient {
	token, err := iam.GetResourceManagementTokenHybrid(activeDirectoryEndpoint, tokenAudience)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, "public IP address", fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}
	ipClient := network.NewPublicIPAddressesClientWithBaseURI(config.Environment().ResourceManagerEndpoint, config.SubscriptionID())
	ipClient.Authorizer = autorest.NewBearerAuthorizer(token)
	ipClient.AddToUserAgent(config.UserAgent())
	return ipClient
}

func getNicClient(activeDirectoryEndpoint, tokenAudience string) network.InterfacesClient {
	token, err := iam.GetResourceManagementTokenHybrid(activeDirectoryEndpoint, tokenAudience)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, "network interface", fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}
	nicClient := network.NewInterfacesClientWithBaseURI(config.Environment().ResourceManagerEndpoint, config.SubscriptionID())
	nicClient.Authorizer = autorest.NewBearerAuthorizer(token)
	nicClient.AddToUserAgent(config.UserAgent())
	return nicClient
}

func getSubnetClient(activeDirectoryEndpoint, tokenAudience string) network.SubnetsClient {
	token, err := iam.GetResourceManagementTokenHybrid(activeDirectoryEndpoint, tokenAudience)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, "subnet", fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}
	subnetsClient := network.NewSubnetsClientWithBaseURI(config.Environment().ResourceManagerEndpoint, config.SubscriptionID())
	subnetsClient.Authorizer = autorest.NewBearerAuthorizer(token)
	subnetsClient.AddToUserAgent(config.UserAgent())
	return subnetsClient
}

// CreateVirtualNetworkAndSubnets creates a virtual network with one subnet
func CreateVirtualNetworkAndSubnets(ctx context.Context, vnetName, subnetName string) (vnet network.VirtualNetwork, err error) {
	resourceName := "virtual network and subnet"
	environment := config.Environment()
	vnetClient := getVnetClient(environment.ActiveDirectoryEndpoint, environment.TokenAudience)
	future, err := vnetClient.CreateOrUpdate(
		ctx,
		config.GroupName(),
		vnetName,
		network.VirtualNetwork{
			Location: to.StringPtr(config.Location()),
			VirtualNetworkPropertiesFormat: &network.VirtualNetworkPropertiesFormat{
				AddressSpace: &network.AddressSpace{
					AddressPrefixes: &[]string{"10.0.0.0/8"},
				},
				Subnets: &[]network.Subnet{
					{
						Name: to.StringPtr(subnetName),
						SubnetPropertiesFormat: &network.SubnetPropertiesFormat{
							AddressPrefix: to.StringPtr("10.0.0.0/16"),
						},
					},
				},
			},
		})

	if err != nil {
		return vnet, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, err))
	}

	err = future.WaitForCompletionRef(ctx, vnetClient.Client)
	if err != nil {
		return vnet, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, fmt.Sprintf("cannot get the vnet create or update future response: %v", err)))
	}

	return future.Result(vnetClient)
}

// CreateNetworkSecurityGroup creates a new network security group
func CreateNetworkSecurityGroup(ctx context.Context, nsgName string) (nsg network.SecurityGroup, err error) {
	resourceName := "security group"
	environment := config.Environment()
	nsgClient := getNsgClient(environment.ActiveDirectoryEndpoint, environment.TokenAudience)
	future, err := nsgClient.CreateOrUpdate(
		ctx,
		config.GroupName(),
		nsgName,
		network.SecurityGroup{
			Location: to.StringPtr(config.Location()),
			SecurityGroupPropertiesFormat: &network.SecurityGroupPropertiesFormat{
				SecurityRules: &[]network.SecurityRule{
					{
						Name: to.StringPtr("allow_ssh"),
						SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
							Protocol:                 network.TCP,
							SourceAddressPrefix:      to.StringPtr("0.0.0.0/0"),
							SourcePortRange:          to.StringPtr("1-65535"),
							DestinationAddressPrefix: to.StringPtr("0.0.0.0/0"),
							DestinationPortRange:     to.StringPtr("22"),
							Access:                   network.Allow,
							Direction:                network.Inbound,
							Priority:                 to.Int32Ptr(100),
						},
					},
					{
						Name: to.StringPtr("allow_https"),
						SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
							Protocol:                 network.TCP,
							SourceAddressPrefix:      to.StringPtr("0.0.0.0/0"),
							SourcePortRange:          to.StringPtr("1-65535"),
							DestinationAddressPrefix: to.StringPtr("0.0.0.0/0"),
							DestinationPortRange:     to.StringPtr("443"),
							Access:                   network.Allow,
							Direction:                network.Inbound,
							Priority:                 to.Int32Ptr(200),
						},
					},
				},
			},
		},
	)

	if err != nil {
		return nsg, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, err))
	}

	err = future.WaitForCompletionRef(ctx, nsgClient.Client)
	if err != nil {
		return nsg, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, fmt.Sprintf("cannot get nsg create or update future response: %v", err)))
	}

	return future.Result(nsgClient)
}

// CreatePublicIP creates a new public IP
func CreatePublicIP(ctx context.Context, ipName string) (ip network.PublicIPAddress, err error) {
	resourceName := "public IP"
	environment, _ := azure.EnvironmentFromURL(config.Environment().ResourceManagerEndpoint)
	ipClient := getIPClient(environment.ActiveDirectoryEndpoint, environment.TokenAudience)
	future, err := ipClient.CreateOrUpdate(
		ctx,
		config.GroupName(),
		ipName,
		network.PublicIPAddress{
			Name:     to.StringPtr(ipName),
			Location: to.StringPtr(config.Location()),
			PublicIPAddressPropertiesFormat: &network.PublicIPAddressPropertiesFormat{
				PublicIPAllocationMethod: network.Static,
			},
		},
	)

	if err != nil {
		return ip, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, err))
	}

	err = future.WaitForCompletionRef(ctx, ipClient.Client)
	if err != nil {
		return ip, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, fmt.Sprintf("cannot get public ip address create or update future response: %v", err)))
	}
	return future.Result(ipClient)
}

// CreateNetworkInterface creates a new network interface
func CreateNetworkInterface(ctx context.Context, netInterfaceName, nsgName, vnetName, subnetName, ipName string) (nic network.Interface, err error) {
	resourceName := "network interface"
	nsg, err := GetNetworkSecurityGroup(ctx, nsgName)
	if err != nil {
		return nic, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, fmt.Sprintf("failed to get netwrok security group: %v", err)))
	}
	subnet, err := GetVirtualNetworkSubnet(ctx, vnetName, subnetName)
	if err != nil {
		return nic, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, fmt.Sprintf("failed to get subnet: %v", err)))
	}
	ip, err := GetPublicIP(ctx, ipName)
	if err != nil {
		return nic, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, fmt.Sprintf("failed to get ip address: %v", err)))
	}
	environment := config.Environment()
	nicClient := getNicClient(environment.ActiveDirectoryEndpoint, environment.TokenAudience)
	future, err := nicClient.CreateOrUpdate(
		ctx,
		config.GroupName(),
		netInterfaceName,
		network.Interface{
			Name:     to.StringPtr(netInterfaceName),
			Location: to.StringPtr(config.Location()),
			InterfacePropertiesFormat: &network.InterfacePropertiesFormat{
				NetworkSecurityGroup: &nsg,
				IPConfigurations: &[]network.InterfaceIPConfiguration{
					{
						Name: to.StringPtr("ipConfig1"),
						InterfaceIPConfigurationPropertiesFormat: &network.InterfaceIPConfigurationPropertiesFormat{
							Subnet: &subnet,
							PrivateIPAllocationMethod: network.Dynamic,
							PublicIPAddress:           &ip,
						},
					},
				},
			},
		},
	)
	if err != nil {
		return nic, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, err))
	}
	err = future.WaitForCompletionRef(ctx, nicClient.Client)
	if err != nil {
		return nic, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, fmt.Sprintf("cannot get nic create or update future response: %v", err)))
	}
	return future.Result(nicClient)
}

// GetNetworkSecurityGroup retrieves a netwrok resource group by its name
func GetNetworkSecurityGroup(ctx context.Context, nsgName string) (network.SecurityGroup, error) {
	environment := config.Environment()
	nsgClient := getNsgClient(environment.ActiveDirectoryEndpoint, environment.TokenAudience)
	return nsgClient.Get(ctx, config.GroupName(), nsgName, "")
}

// GetVirtualNetworkSubnet retrieves a virtual netwrok subnet by its name
func GetVirtualNetworkSubnet(ctx context.Context, vnetName string, subnetName string) (network.Subnet, error) {
	environment := config.Environment()
	subnetsClient := getSubnetClient(environment.ActiveDirectoryEndpoint, environment.TokenAudience)
	return subnetsClient.Get(ctx, config.GroupName(), vnetName, subnetName, "")
}

// GetPublicIP retrieves a public IP by its name
func GetPublicIP(ctx context.Context, ipName string) (network.PublicIPAddress, error) {
	environment := config.Environment()
	ipClient := getIPClient(environment.ActiveDirectoryEndpoint, environment.TokenAudience)
	return ipClient.Get(ctx, config.GroupName(), ipName, "")
}

// GetNic retrieves a network interface by its name
func GetNic(ctx context.Context, nicName string) (network.Interface, error) {
	environment := config.Environment()
	nicClient := getNicClient(environment.ActiveDirectoryEndpoint, environment.TokenAudience)
	return nicClient.Get(ctx, config.GroupName(), nicName, "")
}
