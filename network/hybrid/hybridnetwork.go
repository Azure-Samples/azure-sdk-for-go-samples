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
	"github.com/Azure/go-autorest/autorest/to"
)

var (
	location = "local"
)

func getVnetClient() network.VirtualNetworksClient {
	token, err := iam.GetResourceManagementTokenHybrid(helpers.ActiveDirectoryEndpoint(), helpers.TenantID(), helpers.ClientID(), helpers.ClientSecret(), helpers.ActiveDirectoryResourceID())
	if err != nil {
		log.Fatal(fmt.Sprintf("Cannot generate token. Error details: %s.", err.Error()))
	}
	vnetClient := network.NewVirtualNetworksClientWithBaseURI(helpers.ArmEndpoint(), helpers.SubscriptionID())
	vnetClient.Authorizer = autorest.NewBearerAuthorizer(token)

	return vnetClient
}

func getNsgClient() network.SecurityGroupsClient {
	token, err := iam.GetResourceManagementTokenHybrid(helpers.ActiveDirectoryEndpoint(), helpers.TenantID(), helpers.ClientID(), helpers.ClientSecret(), helpers.ActiveDirectoryResourceID())
	if err != nil {
		log.Fatal(fmt.Sprintf("Cannot generate token. Error details: %s.", err.Error()))
	}
	nsgClient := network.NewSecurityGroupsClientWithBaseURI(helpers.ArmEndpoint(), helpers.SubscriptionID())
	nsgClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return nsgClient
}

func getIPClient() network.PublicIPAddressesClient {
	token, err := iam.GetResourceManagementTokenHybrid(helpers.ActiveDirectoryEndpoint(), helpers.TenantID(), helpers.ClientID(), helpers.ClientSecret(), helpers.ActiveDirectoryResourceID())
	if err != nil {
		log.Fatal(fmt.Sprintf("Cannot generate token. Error details: %s.", err.Error()))
	}
	ipClient := network.NewPublicIPAddressesClientWithBaseURI(helpers.ArmEndpoint(), helpers.SubscriptionID())
	ipClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return ipClient
}

func getNicClient() network.InterfacesClient {
	token, err := iam.GetResourceManagementTokenHybrid(helpers.ActiveDirectoryEndpoint(), helpers.TenantID(), helpers.ClientID(), helpers.ClientSecret(), helpers.ActiveDirectoryResourceID())
	if err != nil {
		log.Fatal(fmt.Sprintf("Cannot generate token. Error details: %s.", err.Error()))
	}
	nicClient := network.NewInterfacesClientWithBaseURI(helpers.ArmEndpoint(), helpers.SubscriptionID())
	nicClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return nicClient
}

func getSubnetsClient() network.SubnetsClient {
	token, err := iam.GetResourceManagementTokenHybrid(helpers.ActiveDirectoryEndpoint(), helpers.TenantID(), helpers.ClientID(), helpers.ClientSecret(), helpers.ActiveDirectoryResourceID())
	if err != nil {
		log.Fatal(fmt.Sprintf("Cannot generate token. Error details: %s.", err.Error()))
	}
	subnetsClient := network.NewSubnetsClientWithBaseURI(helpers.ArmEndpoint(), helpers.SubscriptionID())
	subnetsClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return subnetsClient
}

// CreateVirtualNetworkAndSubnets creates a virtual network with two subnets
func CreateVirtualNetworkAndSubnets(cntx context.Context, vnetName, subnetName string) (vnet network.VirtualNetwork, err error) {
	vnetClient := getVnetClient()
	future, err := vnetClient.CreateOrUpdate(
		cntx,
		helpers.ResourceGroupName(),
		vnetName,
		network.VirtualNetwork{
			Location: to.StringPtr(location),
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
		return vnet, fmt.Errorf("cannot create virtual network: %v", err)
	}

	err = future.WaitForCompletion(cntx, vnetClient.Client)
	if err != nil {
		return vnet, fmt.Errorf("cannot get the vnet create or update future response: %v", err)
	}

	return future.Result(vnetClient)
}

// CreateNetworkSecurityGroup creates a new network security group
func CreateNetworkSecurityGroup(cntx context.Context, nsgName string) (nsg network.SecurityGroup, err error) {
	nsgClient := getNsgClient()
	future, err := nsgClient.CreateOrUpdate(
		cntx,
		helpers.ResourceGroupName(),
		nsgName,
		network.SecurityGroup{
			Location: to.StringPtr(location),
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
		return nsg, fmt.Errorf("cannot create nsg: %v", err)
	}

	err = future.WaitForCompletion(cntx, nsgClient.Client)
	if err != nil {
		return nsg, fmt.Errorf("cannot get nsg create or update future response: %v", err)
	}

	return future.Result(nsgClient)
}

// CreatePublicIP creates a new public IP
func CreatePublicIP(cntx context.Context, ipName string) (ip network.PublicIPAddress, err error) {
	ipClient := getIPClient()
	future, err := ipClient.CreateOrUpdate(
		cntx,
		helpers.ResourceGroupName(),
		ipName,
		network.PublicIPAddress{
			Name:     to.StringPtr(ipName),
			Location: to.StringPtr(location),
			PublicIPAddressPropertiesFormat: &network.PublicIPAddressPropertiesFormat{
				PublicIPAllocationMethod: network.Static,
			},
		},
	)

	if err != nil {
		return ip, fmt.Errorf("cannot create public ip address: %v", err)
	}

	err = future.WaitForCompletion(cntx, ipClient.Client)
	if err != nil {
		return ip, fmt.Errorf("cannot get public ip address create or update future response: %v", err)
	}
	return future.Result(ipClient)
}

// CreateNetworkInterface creates a new network interface
func CreateNetworkInterface(cntx context.Context, netInterfaceName, nsgName, vnetName, subnetName, ipName string) (nic network.Interface, err error) {
	nsg, err := GetNetworkSecurityGroup(cntx, nsgName)
	if err != nil {
		log.Fatalf("failed to get netwrok security group: %v", err)
	}
	subnet, err := GetVirtualNetworkSubnet(cntx, vnetName, subnetName)
	if err != nil {
		log.Fatalf("failed to get subnet: %v", err)
	}
	ip, err := GetPublicIP(cntx, ipName)
	if err != nil {
		log.Fatalf("failed to get ip address: %v", err)
	}
	nicClient := getNicClient()
	future, err := nicClient.CreateOrUpdate(
		cntx,
		helpers.ResourceGroupName(),
		netInterfaceName,
		network.Interface{
			Name:     to.StringPtr(netInterfaceName),
			Location: to.StringPtr(location),
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
		return nic, fmt.Errorf("cannot create nic: %v", err)
	}
	err = future.WaitForCompletion(cntx, nicClient.Client)
	if err != nil {
		return nic, fmt.Errorf("cannot get nic create or update future response: %v", err)
	}
	return future.Result(nicClient)
}

func GetNetworkSecurityGroup(cntx context.Context, nsgName string) (network.SecurityGroup, error) {
	nsgClient := getNsgClient()
	return nsgClient.Get(cntx, helpers.ResourceGroupName(), nsgName, "")
}

func GetVirtualNetworkSubnet(cntx context.Context, vnetName string, subnetName string) (network.Subnet, error) {
	subnetsClient := getSubnetsClient()
	return subnetsClient.Get(cntx, helpers.ResourceGroupName(), vnetName, subnetName, "")
}

func GetPublicIP(cntx context.Context, ipName string) (network.PublicIPAddress, error) {
	ipClient := getIPClient()
	return ipClient.Get(cntx, helpers.ResourceGroupName(), ipName, "")
}

func GetNic(cntx context.Context, nicName string) (network.Interface, error) {
	nicClient := getNicClient()
	return nicClient.Get(cntx, helpers.ResourceGroupName(), nicName, "")
}
