package network

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2017-09-01/network"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

// Vnets

func getVnetClient() network.VirtualNetworksClient {
	token, _ := iam.GetResourceManagementToken(iam.AuthGrantType())
	vnetClient := network.NewVirtualNetworksClient(helpers.SubscriptionID())
	vnetClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return vnetClient
}

// CreateVirtualNetworkAndSubnets creates a virtual network with two subnets
func CreateVirtualNetworkAndSubnets(ctx context.Context, vnetName, subnet1Name, subnet2Name string) (vnet network.VirtualNetwork, err error) {
	vnetClient := getVnetClient()
	future, err := vnetClient.CreateOrUpdate(
		ctx,
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
						Name: to.StringPtr(subnet1Name),
						SubnetPropertiesFormat: &network.SubnetPropertiesFormat{
							AddressPrefix: to.StringPtr("10.0.0.0/16"),
						},
					},
					{
						Name: to.StringPtr(subnet2Name),
						SubnetPropertiesFormat: &network.SubnetPropertiesFormat{
							AddressPrefix: to.StringPtr("10.1.0.0/16"),
						},
					},
				},
			},
		})

	if err != nil {
		return vnet, fmt.Errorf("cannot create virtual network: %v", err)
	}

	err = future.WaitForCompletion(ctx, vnetClient.Client)
	if err != nil {
		return vnet, fmt.Errorf("cannot get the vnet create or update future response: %v", err)
	}

	return future.Result(vnetClient)
}

// DeleteVirtualNetwork deletes a virtual network given an existing virtual network
func DeleteVirtualNetwork(ctx context.Context, vnetName string) (result network.VirtualNetworksDeleteFuture, err error) {
	vnetClient := getVnetClient()
	return vnetClient.Delete(ctx, helpers.ResourceGroupName(), vnetName)
}

// VNet Subnets

func getSubnetsClient() network.SubnetsClient {
	token, _ := iam.GetResourceManagementToken(iam.AuthGrantType())
	subnetsClient := network.NewSubnetsClient(helpers.SubscriptionID())
	subnetsClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return subnetsClient
}

// CreateVirtualNetworkSubnet creates a subnet
func CreateVirtualNetworkSubnet() {}

// DeleteVirtualNetworkSubnet deletes a subnet
func DeleteVirtualNetworkSubnet() {}

// GetVirtualNetworkSubnet returns an existing subnet from a virtual network
func GetVirtualNetworkSubnet(ctx context.Context, vnetName string, subnetName string) (network.Subnet, error) {
	subnetsClient := getSubnetsClient()
	return subnetsClient.Get(ctx, helpers.ResourceGroupName(), vnetName, subnetName, "")
}

// Network Security Groups

func getNsgClient() network.SecurityGroupsClient {
	token, _ := iam.GetResourceManagementToken(iam.AuthGrantType())
	nsgClient := network.NewSecurityGroupsClient(helpers.SubscriptionID())
	nsgClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return nsgClient
}

// CreateNetworkSecurityGroup creates a new network security group
func CreateNetworkSecurityGroup(ctx context.Context, nsgName string) (nsg network.SecurityGroup, err error) {
	nsgClient := getNsgClient()
	future, err := nsgClient.CreateOrUpdate(
		ctx,
		helpers.ResourceGroupName(),
		nsgName,
		network.SecurityGroup{
			Location: to.StringPtr(helpers.Location()),
			SecurityGroupPropertiesFormat: &network.SecurityGroupPropertiesFormat{
				SecurityRules: &[]network.SecurityRule{
					{
						Name: to.StringPtr("allow_ssh"),
						SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
							Protocol:                 network.SecurityRuleProtocolTCP,
							SourceAddressPrefix:      to.StringPtr("0.0.0.0/0"),
							SourcePortRange:          to.StringPtr("1-65535"),
							DestinationAddressPrefix: to.StringPtr("0.0.0.0/0"),
							DestinationPortRange:     to.StringPtr("22"),
							Access:                   network.SecurityRuleAccessAllow,
							Direction:                network.SecurityRuleDirectionInbound,
							Priority:                 to.Int32Ptr(100),
						},
					},
					{
						Name: to.StringPtr("allow_https"),
						SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
							Protocol:                 network.SecurityRuleProtocolTCP,
							SourceAddressPrefix:      to.StringPtr("0.0.0.0/0"),
							SourcePortRange:          to.StringPtr("1-65535"),
							DestinationAddressPrefix: to.StringPtr("0.0.0.0/0"),
							DestinationPortRange:     to.StringPtr("443"),
							Access:                   network.SecurityRuleAccessAllow,
							Direction:                network.SecurityRuleDirectionInbound,
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

	err = future.WaitForCompletion(ctx, nsgClient.Client)
	if err != nil {
		return nsg, fmt.Errorf("cannot get nsg create or update future response: %v", err)
	}

	return future.Result(nsgClient)
}

// DeleteNetworkSecurityGroup deletes an existing network security group
func DeleteNetworkSecurityGroup(ctx context.Context, nsgName string) (result network.SecurityGroupsDeleteFuture, err error) {
	nsgClient := getNsgClient()
	return nsgClient.Delete(ctx, helpers.ResourceGroupName(), nsgName)
}

// GetNetworkSecurityGroup returns an existing network security group
func GetNetworkSecurityGroup(ctx context.Context, nsgName string) (network.SecurityGroup, error) {
	nsgClient := getNsgClient()
	return nsgClient.Get(ctx, helpers.ResourceGroupName(), nsgName, "")
}

// Network Security Group Rules

// CreateNetworkSecurityGroupRule creates a network security group rule
func CreateNetworkSecurityGroupRule() {}

// DeleteNetworkSecurityGroupRule deletes a network security group rule
func DeleteNetworkSecurityGroupRule() {}

// Network Interfaces (NIC's)

func getNicClient() network.InterfacesClient {
	token, _ := iam.GetResourceManagementToken(iam.AuthGrantType())
	nicClient := network.NewInterfacesClient(helpers.SubscriptionID())
	nicClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return nicClient
}

// CreateNIC creates a new network interface
func CreateNIC(ctx context.Context, vnetName, subnetName, nsgName, ipName, nicName string) (nic network.Interface, err error) {
	nsg, err := GetNetworkSecurityGroup(ctx, nsgName)
	if err != nil {
		log.Fatalf("failed to get nsg: %v", err)
	}

	subnet, err := GetVirtualNetworkSubnet(ctx, vnetName, subnetName)
	if err != nil {
		log.Fatalf("failed to get subnet: %v", err)
	}

	ip, err := GetPublicIP(ctx, ipName)
	if err != nil {
		log.Fatalf("failed to get ip address: %v", err)
	}

	nicClient := getNicClient()
	future, err := nicClient.CreateOrUpdate(
		ctx,
		helpers.ResourceGroupName(),
		nicName,
		network.Interface{
			Name:     to.StringPtr(nicName),
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
		return nic, fmt.Errorf("cannot create nic: %v", err)
	}

	err = future.WaitForCompletion(ctx, nicClient.Client)
	if err != nil {
		return nic, fmt.Errorf("cannot get nic create or update future response: %v", err)
	}

	return future.Result(nicClient)
}

// GetNic returns an existing network interface
func GetNic(ctx context.Context, nicName string) (network.Interface, error) {
	nicClient := getNicClient()
	return nicClient.Get(ctx, helpers.ResourceGroupName(), nicName, "")
}

// DeleteNic deletes an existing network interface
func DeleteNic(ctx context.Context, nic string) (result network.InterfacesDeleteFuture, err error) {
	nicClient := getNicClient()
	return nicClient.Delete(ctx, helpers.ResourceGroupName(), nic)
}

// Public IP Addresses

func getIPClient() network.PublicIPAddressesClient {
	token, _ := iam.GetResourceManagementToken(iam.AuthGrantType())
	ipClient := network.NewPublicIPAddressesClient(helpers.SubscriptionID())
	ipClient.Authorizer = autorest.NewBearerAuthorizer(token)
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
