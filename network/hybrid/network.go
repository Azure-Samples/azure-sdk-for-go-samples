// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package hybridnetwork

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
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
	vnetClient := network.NewVirtualNetworksClientWithBaseURI(helpers.ArmEndpoint(), helpers.SubscriptionID())
	vnetClient.Authorizer = autorest.NewBearerAuthorizer(token)
	vnetClient.AddToUserAgent(helpers.UserAgent())
	return vnetClient
}

func getNsgClient(activeDirectoryEndpoint, tokenAudience string) network.SecurityGroupsClient {
	token, err := iam.GetResourceManagementTokenHybrid(activeDirectoryEndpoint, tokenAudience)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, "security group", fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}
	nsgClient := network.NewSecurityGroupsClientWithBaseURI(helpers.ArmEndpoint(), helpers.SubscriptionID())
	nsgClient.Authorizer = autorest.NewBearerAuthorizer(token)
	nsgClient.AddToUserAgent(helpers.UserAgent())
	return nsgClient
}

func getIPClient(activeDirectoryEndpoint, tokenAudience string) network.PublicIPAddressesClient {
	token, err := iam.GetResourceManagementTokenHybrid(activeDirectoryEndpoint, tokenAudience)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, "public IP address", fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}
	ipClient := network.NewPublicIPAddressesClientWithBaseURI(helpers.ArmEndpoint(), helpers.SubscriptionID())
	ipClient.Authorizer = autorest.NewBearerAuthorizer(token)
	ipClient.AddToUserAgent(helpers.UserAgent())
	return ipClient
}

func getNicClient(activeDirectoryEndpoint, tokenAudience string) network.InterfacesClient {
	token, err := iam.GetResourceManagementTokenHybrid(activeDirectoryEndpoint, tokenAudience)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, "network interface", fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}
	nicClient := network.NewInterfacesClientWithBaseURI(helpers.ArmEndpoint(), helpers.SubscriptionID())
	nicClient.Authorizer = autorest.NewBearerAuthorizer(token)
	nicClient.AddToUserAgent(helpers.UserAgent())
	return nicClient
}

func getSubnetsClient(activeDirectoryEndpoint, tokenAudience string) network.SubnetsClient {
	token, err := iam.GetResourceManagementTokenHybrid(activeDirectoryEndpoint, tokenAudience)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, "subnet", fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}
	subnetsClient := network.NewSubnetsClientWithBaseURI(helpers.ArmEndpoint(), helpers.SubscriptionID())
	subnetsClient.Authorizer = autorest.NewBearerAuthorizer(token)
	subnetsClient.AddToUserAgent(helpers.UserAgent())
	return subnetsClient
}

// CreateVirtualNetworkAndSubnets creates a virtual network with one subnet
func CreateVirtualNetworkAndSubnets(cntx context.Context, vnetName, subnetName string) (vnet network.VirtualNetwork, err error) {
	resourceName := "virtual network and subnet"
	environment, _ := azure.EnvironmentFromURL(helpers.ArmEndpoint())
	vnetClient := getVnetClient(environment.ActiveDirectoryEndpoint, environment.TokenAudience)
	future, err := vnetClient.CreateOrUpdate(
		cntx,
		helpers.ResourceGroupName(),
		vnetName,
		network.VirtualNetwork{
			Location: to.StringPtr(helpers.Location()),
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

	err = future.WaitForCompletion(cntx, vnetClient.Client)
	if err != nil {
		return vnet, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, fmt.Sprintf("cannot get the vnet create or update future response: %v", err)))
	}

	return future.Result(vnetClient)
}

// CreateNetworkSecurityGroup creates a new network security group
func CreateNetworkSecurityGroup(cntx context.Context, nsgName string) (nsg network.SecurityGroup, err error) {
	resourceName := "security group"
	environment, _ := azure.EnvironmentFromURL(helpers.ArmEndpoint())
	nsgClient := getNsgClient(environment.ActiveDirectoryEndpoint, environment.TokenAudience)
	future, err := nsgClient.CreateOrUpdate(
		cntx,
		helpers.ResourceGroupName(),
		nsgName,
		network.SecurityGroup{
			Location: to.StringPtr(helpers.Location()),
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

	err = future.WaitForCompletion(cntx, nsgClient.Client)
	if err != nil {
		return nsg, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, fmt.Sprintf("cannot get nsg create or update future response: %v", err)))
	}

	return future.Result(nsgClient)
}

// CreatePublicIP creates a new public IP
func CreatePublicIP(cntx context.Context, ipName string) (ip network.PublicIPAddress, err error) {
	resourceName := "public IP"
	environment, _ := azure.EnvironmentFromURL(helpers.ArmEndpoint())
	ipClient := getIPClient(environment.ActiveDirectoryEndpoint, environment.TokenAudience)
	future, err := ipClient.CreateOrUpdate(
		cntx,
		helpers.ResourceGroupName(),
		ipName,
		network.PublicIPAddress{
			Name:     to.StringPtr(ipName),
			Location: to.StringPtr(helpers.Location()),
			PublicIPAddressPropertiesFormat: &network.PublicIPAddressPropertiesFormat{
				PublicIPAllocationMethod: network.Static,
			},
		},
	)

	if err != nil {
		return ip, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, err))
	}

	err = future.WaitForCompletion(cntx, ipClient.Client)
	if err != nil {
		return ip, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, fmt.Sprintf("cannot get public ip address create or update future response: %v", err)))
	}
	return future.Result(ipClient)
}

// CreateNetworkInterface creates a new network interface
func CreateNetworkInterface(cntx context.Context, netInterfaceName, nsgName, vnetName, subnetName, ipName string) (nic network.Interface, err error) {
	resourceName := "network interface"
	nsg, err := GetNetworkSecurityGroup(cntx, nsgName)
	if err != nil {
		return nic, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, fmt.Sprintf("failed to get netwrok security group: %v", err)))
	}
	subnet, err := GetVirtualNetworkSubnet(cntx, vnetName, subnetName)
	if err != nil {
		return nic, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, fmt.Sprintf("failed to get subnet: %v", err)))
	}
	ip, err := GetPublicIP(cntx, ipName)
	if err != nil {
		return nic, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, fmt.Sprintf("failed to get ip address: %v", err)))
	}
	environment, _ := azure.EnvironmentFromURL(helpers.ArmEndpoint())
	nicClient := getNicClient(environment.ActiveDirectoryEndpoint, environment.TokenAudience)
	future, err := nicClient.CreateOrUpdate(
		cntx,
		helpers.ResourceGroupName(),
		netInterfaceName,
		network.Interface{
			Name:     to.StringPtr(netInterfaceName),
			Location: to.StringPtr(helpers.Location()),
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
	err = future.WaitForCompletion(cntx, nicClient.Client)
	if err != nil {
		return nic, fmt.Errorf(fmt.Sprintf(errorPrefix, resourceName, fmt.Sprintf("cannot get nic create or update future response: %v", err)))
	}
	return future.Result(nicClient)
}

// GetNetworkSecurityGroup retrieves a netwrok resource group by its name
func GetNetworkSecurityGroup(cntx context.Context, nsgName string) (network.SecurityGroup, error) {
	environment, _ := azure.EnvironmentFromURL(helpers.ArmEndpoint())
	nsgClient := getNsgClient(environment.ActiveDirectoryEndpoint, environment.TokenAudience)
	return nsgClient.Get(cntx, helpers.ResourceGroupName(), nsgName, "")
}

// GetVirtualNetworkSubnet retrieves a virtual netwrok subnet by its name
func GetVirtualNetworkSubnet(cntx context.Context, vnetName string, subnetName string) (network.Subnet, error) {
	environment, _ := azure.EnvironmentFromURL(helpers.ArmEndpoint())
	subnetsClient := getSubnetsClient(environment.ActiveDirectoryEndpoint, environment.TokenAudience)
	return subnetsClient.Get(cntx, helpers.ResourceGroupName(), vnetName, subnetName, "")
}

// GetPublicIP retrieves a public IP by its name
func GetPublicIP(cntx context.Context, ipName string) (network.PublicIPAddress, error) {
	environment, _ := azure.EnvironmentFromURL(helpers.ArmEndpoint())
	ipClient := getIPClient(environment.ActiveDirectoryEndpoint, environment.TokenAudience)
	return ipClient.Get(cntx, helpers.ResourceGroupName(), ipName, "")
}

// GetNic retrieves a network interface by its name
func GetNic(cntx context.Context, nicName string) (network.Interface, error) {
	environment, _ := azure.EnvironmentFromURL(helpers.ArmEndpoint())
	nicClient := getNicClient(environment.ActiveDirectoryEndpoint, environment.TokenAudience)
	return nicClient.Get(cntx, helpers.ResourceGroupName(), nicName, "")
}
