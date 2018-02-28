// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

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
	vnetClient.AddToUserAgent(helpers.UserAgent())
	return vnetClient
}

// CreateVirtualNetwork creates a virtual network
func CreateVirtualNetwork(ctx context.Context, vnetName string) (vnet network.VirtualNetwork, err error) {
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
	nsgClient.AddToUserAgent(helpers.UserAgent())
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
	nicClient.AddToUserAgent(helpers.UserAgent())
	return nicClient
}

// CreateNIC creates a new network interface. The Network Security Group is not a required parameter
func CreateNIC(ctx context.Context, vnetName, subnetName, nsgName, ipName, nicName string) (nic network.Interface, err error) {
	subnet, err := GetVirtualNetworkSubnet(ctx, vnetName, subnetName)
	if err != nil {
		log.Fatalf("failed to get subnet: %v", err)
	}

	ip, err := GetPublicIP(ctx, ipName)
	if err != nil {
		log.Fatalf("failed to get ip address: %v", err)
	}

	nicParams := network.Interface{
		Name:     to.StringPtr(nicName),
		Location: to.StringPtr(helpers.Location()),
		InterfacePropertiesFormat: &network.InterfacePropertiesFormat{
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
	}

	if nsgName != "" {
		nsg, err := GetNetworkSecurityGroup(ctx, nsgName)
		if err != nil {
			log.Fatalf("failed to get nsg: %v", err)
		}
		nicParams.NetworkSecurityGroup = &nsg
	}

	nicClient := getNicClient()
	future, err := nicClient.CreateOrUpdate(ctx, helpers.ResourceGroupName(), nicName, nicParams)
	if err != nil {
		return nic, fmt.Errorf("cannot create nic: %v", err)
	}

	err = future.WaitForCompletion(ctx, nicClient.Client)
	if err != nil {
		return nic, fmt.Errorf("cannot get nic create or update future response: %v", err)
	}

	return future.Result(nicClient)
}

func CreateNICWithLoadBalancer(ctx context.Context, lbName, vnetName, subnetName, nicName string, natRule int) (nic network.Interface, err error) {
	subnet, err := GetVirtualNetworkSubnet(ctx, vnetName, subnetName)
	if err != nil {
		return
	}

	lb, err := GetLoadBalancer(ctx, lbName)
	if err != nil {
		return
	}

	nicClient := getNicClient()
	future, err := nicClient.CreateOrUpdate(ctx,
		helpers.ResourceGroupName(),
		nicName,
		network.Interface{
			Location: to.StringPtr(helpers.Location()),
			InterfacePropertiesFormat: &network.InterfacePropertiesFormat{
				IPConfigurations: &[]network.InterfaceIPConfiguration{
					{
						Name: to.StringPtr("pipConfig"),
						InterfaceIPConfigurationPropertiesFormat: &network.InterfaceIPConfigurationPropertiesFormat{
							Subnet: &network.Subnet{
								ID: subnet.ID,
							},
							LoadBalancerBackendAddressPools: &[]network.BackendAddressPool{
								{
									ID: (*lb.BackendAddressPools)[0].ID,
								},
							},
							LoadBalancerInboundNatRules: &[]network.InboundNatRule{
								{
									ID: (*lb.InboundNatRules)[natRule].ID,
								},
							},
						},
					},
				},
			},
		})
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

// Load balancers

func getLBClient() network.LoadBalancersClient {
	token, _ := iam.GetResourceManagementToken(iam.AuthGrantType())
	lbClient := network.NewLoadBalancersClient(helpers.SubscriptionID())
	lbClient.Authorizer = autorest.NewBearerAuthorizer(token)
	lbClient.AddToUserAgent(helpers.UserAgent())
	return lbClient
}

func GetLoadBalancer(ctx context.Context, lbName string) (network.LoadBalancer, error) {
	lbClient := getLBClient()
	return lbClient.Get(ctx, helpers.ResourceGroupName(), lbName, "")
}

func CreateLoadBalancer(ctx context.Context, lbName, pipName string) (lb network.LoadBalancer, err error) {
	probeName := "probe"
	frontEndIPConfigName := "fip"
	backEndAddressPoolName := "backEndPool"
	idPrefix := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/loadBalancers", helpers.SubscriptionID(), helpers.ResourceGroupName())

	pip, err := GetPublicIP(ctx, pipName)
	if err != nil {
		return
	}

	lbClient := getLBClient()
	future, err := lbClient.CreateOrUpdate(ctx,
		helpers.ResourceGroupName(),
		lbName,
		network.LoadBalancer{
			Location: to.StringPtr(helpers.Location()),
			LoadBalancerPropertiesFormat: &network.LoadBalancerPropertiesFormat{
				FrontendIPConfigurations: &[]network.FrontendIPConfiguration{
					{
						Name: &frontEndIPConfigName,
						FrontendIPConfigurationPropertiesFormat: &network.FrontendIPConfigurationPropertiesFormat{
							PrivateIPAllocationMethod: network.Dynamic,
							PublicIPAddress:           &pip,
						},
					},
				},
				BackendAddressPools: &[]network.BackendAddressPool{
					{
						Name: &backEndAddressPoolName},
				},
				Probes: &[]network.Probe{
					{
						Name: &probeName,
						ProbePropertiesFormat: &network.ProbePropertiesFormat{
							Protocol:          network.ProbeProtocolHTTP,
							Port:              to.Int32Ptr(80),
							IntervalInSeconds: to.Int32Ptr(15),
							NumberOfProbes:    to.Int32Ptr(4),
							RequestPath:       to.StringPtr("healthprobe.aspx"),
						},
					},
				},
				LoadBalancingRules: &[]network.LoadBalancingRule{
					{
						Name: to.StringPtr("lbRule"),
						LoadBalancingRulePropertiesFormat: &network.LoadBalancingRulePropertiesFormat{
							Protocol:             network.TransportProtocolTCP,
							FrontendPort:         to.Int32Ptr(80),
							BackendPort:          to.Int32Ptr(80),
							IdleTimeoutInMinutes: to.Int32Ptr(4),
							EnableFloatingIP:     to.BoolPtr(false),
							LoadDistribution:     network.Default,
							FrontendIPConfiguration: &network.SubResource{
								ID: to.StringPtr(fmt.Sprintf("/%s/%s/frontendIPConfigurations/%s", idPrefix, lbName, frontEndIPConfigName)),
							},
							BackendAddressPool: &network.SubResource{
								ID: to.StringPtr(fmt.Sprintf("/%s/%s/backendAddressPools/%s", idPrefix, lbName, backEndAddressPoolName)),
							},
							Probe: &network.SubResource{
								ID: to.StringPtr(fmt.Sprintf("/%s/%s/probes/%s", idPrefix, lbName, probeName)),
							},
						},
					},
				},
				InboundNatRules: &[]network.InboundNatRule{
					network.InboundNatRule{
						Name: to.StringPtr("natRule1"),
						InboundNatRulePropertiesFormat: &network.InboundNatRulePropertiesFormat{
							Protocol:             network.TransportProtocolTCP,
							FrontendPort:         to.Int32Ptr(21),
							BackendPort:          to.Int32Ptr(22),
							EnableFloatingIP:     to.BoolPtr(false),
							IdleTimeoutInMinutes: to.Int32Ptr(4),
							FrontendIPConfiguration: &network.SubResource{
								ID: to.StringPtr(fmt.Sprintf("/%s/%s/frontendIPConfigurations/%s", idPrefix, lbName, frontEndIPConfigName)),
							},
						},
					},
					network.InboundNatRule{
						Name: to.StringPtr("natRule2"),
						InboundNatRulePropertiesFormat: &network.InboundNatRulePropertiesFormat{
							Protocol:             network.TransportProtocolTCP,
							FrontendPort:         to.Int32Ptr(23),
							BackendPort:          to.Int32Ptr(22),
							EnableFloatingIP:     to.BoolPtr(false),
							IdleTimeoutInMinutes: to.Int32Ptr(4),
							FrontendIPConfiguration: &network.SubResource{
								ID: to.StringPtr(fmt.Sprintf("/%s/%s/frontendIPConfigurations/%s", idPrefix, lbName, frontEndIPConfigName)),
							},
						},
					},
				},
			},
		})

	if err != nil {
		return lb, fmt.Errorf("cannot create load balancer: %v", err)
	}

	err = future.WaitForCompletion(ctx, lbClient.Client)
	if err != nil {
		return lb, fmt.Errorf("cannot get load balancer create or update future response: %v", err)
	}

	return future.Result(lbClient)
}
