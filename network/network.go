// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package network

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2017-09-01/network"
	"github.com/Azure/go-autorest/autorest/to"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
)

// Vnets

func getVnetClient() network.VirtualNetworksClient {
	vnetClient := network.NewVirtualNetworksClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	vnetClient.Authorizer = a
	vnetClient.AddToUserAgent(config.UserAgent())
	return vnetClient
}

// CreateVirtualNetwork creates a virtual network
func CreateVirtualNetwork(ctx context.Context, vnetName string) (vnet network.VirtualNetwork, err error) {
	vnetClient := getVnetClient()
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
	return vnetClient.Delete(ctx, config.GroupName(), vnetName)
}

// VNet Subnets

func getSubnetsClient() network.SubnetsClient {
	subnetsClient := network.NewSubnetsClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	subnetsClient.Authorizer = a
	subnetsClient.AddToUserAgent(config.UserAgent())
	return subnetsClient
}

// CreateVirtualNetworkSubnet creates a subnet in an existing vnet
func CreateVirtualNetworkSubnet(ctx context.Context, vnetName, subnetName string) (subnet network.Subnet, err error) {
	subnetsClient := getSubnetsClient()

	future, err := subnetsClient.CreateOrUpdate(
		ctx,
		config.GroupName(),
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

// CreateSubnetWithNetowrkSecurityGroup create a subnet referencing a network secuiry group
func CreateSubnetWithNetowrkSecurityGroup(ctx context.Context, vnetName, subnetName, addressPrefix, nsgName string) (subnet network.Subnet, err error) {
	nsg, err := GetNetworkSecurityGroup(ctx, nsgName)
	if err != nil {
		return subnet, fmt.Errorf("cannot get nsg: %v", err)
	}

	subnetsClient := getSubnetsClient()
	future, err := subnetsClient.CreateOrUpdate(
		ctx,
		config.GroupName(),
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
	return subnetsClient.Get(ctx, config.GroupName(), vnetName, subnetName, "")
}

// Network Security Groups

func getNsgClient() network.SecurityGroupsClient {
	nsgClient := network.NewSecurityGroupsClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	nsgClient.Authorizer = a
	nsgClient.AddToUserAgent(config.UserAgent())
	return nsgClient
}

// CreateNetworkSecurityGroup creates a new network security group with rules set for allowing SSH and HTTPS use
func CreateNetworkSecurityGroup(ctx context.Context, nsgName string) (nsg network.SecurityGroup, err error) {
	nsgClient := getNsgClient()
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

// CreateSimpleNetworkSecurityGroup creates a new network security group, without rules (rules can be set later)
func CreateSimpleNetworkSecurityGroup(ctx context.Context, nsgName string) (nsg network.SecurityGroup, err error) {
	nsgClient := getNsgClient()
	future, err := nsgClient.CreateOrUpdate(
		ctx,
		config.GroupName(),
		nsgName,
		network.SecurityGroup{
			Location: to.StringPtr(config.Location()),
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
	return nsgClient.Delete(ctx, config.GroupName(), nsgName)
}

// GetNetworkSecurityGroup returns an existing network security group
func GetNetworkSecurityGroup(ctx context.Context, nsgName string) (network.SecurityGroup, error) {
	nsgClient := getNsgClient()
	return nsgClient.Get(ctx, config.GroupName(), nsgName, "")
}

// Network Security Group Rules

// CreateNetworkSecurityGroupRule creates a network security group rule
func CreateNetworkSecurityGroupRule() {}

// DeleteNetworkSecurityGroupRule deletes a network security group rule
func DeleteNetworkSecurityGroupRule() {}

// Network Interfaces (NIC's)

func getNicClient() network.InterfacesClient {
	nicClient := network.NewInterfacesClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	nicClient.Authorizer = a
	nicClient.AddToUserAgent(config.UserAgent())
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
		Location: to.StringPtr(config.Location()),
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
	future, err := nicClient.CreateOrUpdate(ctx, config.GroupName(), nicName, nicParams)
	if err != nil {
		return nic, fmt.Errorf("cannot create nic: %v", err)
	}

	err = future.WaitForCompletion(ctx, nicClient.Client)
	if err != nil {
		return nic, fmt.Errorf("cannot get nic create or update future response: %v", err)
	}

	return future.Result(nicClient)
}

// CreateNICWithLoadBalancer creats a network interface, wich is set up with a loadbalancer's NAT rule
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
		config.GroupName(),
		nicName,
		network.Interface{
			Location: to.StringPtr(config.Location()),
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
	return nicClient.Get(ctx, config.GroupName(), nicName, "")
}

// DeleteNic deletes an existing network interface
func DeleteNic(ctx context.Context, nic string) (result network.InterfacesDeleteFuture, err error) {
	nicClient := getNicClient()
	return nicClient.Delete(ctx, config.GroupName(), nic)
}

// Public IP Addresses

func getIPClient() network.PublicIPAddressesClient {
	ipClient := network.NewPublicIPAddressesClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	ipClient.Authorizer = a
	ipClient.AddToUserAgent(config.UserAgent())
	return ipClient
}

// CreatePublicIP creates a new public IP
func CreatePublicIP(ctx context.Context, ipName string) (ip network.PublicIPAddress, err error) {
	ipClient := getIPClient()
	future, err := ipClient.CreateOrUpdate(
		ctx,
		config.GroupName(),
		ipName,
		network.PublicIPAddress{
			Name:     to.StringPtr(ipName),
			Location: to.StringPtr(config.Location()),
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
	return ipClient.Get(ctx, config.GroupName(), ipName, "")
}

// DeletePublicIP deletes an existing public IP
func DeletePublicIP(ctx context.Context, ipName string) (result network.PublicIPAddressesDeleteFuture, err error) {
	ipClient := getIPClient()
	return ipClient.Delete(ctx, config.GroupName(), ipName)
}

func getSecurityRulesClient() network.SecurityRulesClient {
	rulesClient := network.NewSecurityRulesClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	rulesClient.Authorizer = a
	rulesClient.AddToUserAgent(config.UserAgent())
	return rulesClient
}

// CreateSSHRule creates an inbound network security rule that allows using port 22
func CreateSSHRule(ctx context.Context, nsgName string) (rule network.SecurityRule, err error) {
	rulesClient := getSecurityRulesClient()
	future, err := rulesClient.CreateOrUpdate(ctx,
		config.GroupName(),
		nsgName,
		"ALLOW-SSH",
		network.SecurityRule{
			SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
				Access: network.SecurityRuleAccessAllow,
				DestinationAddressPrefix: to.StringPtr("*"),
				DestinationPortRange:     to.StringPtr("22"),
				Direction:                network.SecurityRuleDirectionInbound,
				Description:              to.StringPtr("Allow SSH"),
				Priority:                 to.Int32Ptr(103),
				Protocol:                 network.SecurityRuleProtocolTCP,
				SourceAddressPrefix:      to.StringPtr("*"),
				SourcePortRange:          to.StringPtr("*"),
			},
		})
	if err != nil {
		return rule, fmt.Errorf("cannot create SSH security rule: %v", err)
	}

	err = future.WaitForCompletion(ctx, rulesClient.Client)
	if err != nil {
		return rule, fmt.Errorf("cannot get security rule create or update future response: %v", err)
	}

	return future.Result(rulesClient)
}

// CreateHTTPRule creates an inbound network security rule that allows using port 80
func CreateHTTPRule(ctx context.Context, nsgName string) (rule network.SecurityRule, err error) {
	rulesClient := getSecurityRulesClient()
	future, err := rulesClient.CreateOrUpdate(ctx,
		config.GroupName(),
		nsgName,
		"ALLOW-HTTP",
		network.SecurityRule{
			SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
				Access: network.SecurityRuleAccessAllow,
				DestinationAddressPrefix: to.StringPtr("*"),
				DestinationPortRange:     to.StringPtr("80"),
				Direction:                network.SecurityRuleDirectionInbound,
				Description:              to.StringPtr("Allow HTTP"),
				Priority:                 to.Int32Ptr(101),
				Protocol:                 network.SecurityRuleProtocolTCP,
				SourceAddressPrefix:      to.StringPtr("*"),
				SourcePortRange:          to.StringPtr("*"),
			},
		})
	if err != nil {
		return rule, fmt.Errorf("cannot create HTTP security rule: %v", err)
	}

	err = future.WaitForCompletion(ctx, rulesClient.Client)
	if err != nil {
		return rule, fmt.Errorf("cannot get security rule create or update future response: %v", err)
	}

	return future.Result(rulesClient)
}

// CreateSQLRule creates an inbound network security rule that allows using port 1433
func CreateSQLRule(ctx context.Context, nsgName, frontEndAddressPrefix string) (rule network.SecurityRule, err error) {
	rulesClient := getSecurityRulesClient()
	future, err := rulesClient.CreateOrUpdate(ctx,
		config.GroupName(),
		nsgName,
		"ALLOW-SQL",
		network.SecurityRule{
			SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
				Access: network.SecurityRuleAccessAllow,
				DestinationAddressPrefix: to.StringPtr("*"),
				DestinationPortRange:     to.StringPtr("1433"),
				Direction:                network.SecurityRuleDirectionInbound,
				Description:              to.StringPtr("Allow SQL"),
				Priority:                 to.Int32Ptr(102),
				Protocol:                 network.SecurityRuleProtocolTCP,
				SourceAddressPrefix:      &frontEndAddressPrefix,
				SourcePortRange:          to.StringPtr("*"),
			},
		})
	if err != nil {
		return rule, fmt.Errorf("cannot create SQL security rule: %v", err)
	}

	err = future.WaitForCompletion(ctx, rulesClient.Client)
	if err != nil {
		return rule, fmt.Errorf("cannot get security rule create or update future response: %v", err)
	}

	return future.Result(rulesClient)
}

// CreateDenyOutRule creates an network security rule that denies outbound traffic
func CreateDenyOutRule(ctx context.Context, nsgName string) (rule network.SecurityRule, err error) {
	rulesClient := getSecurityRulesClient()
	future, err := rulesClient.CreateOrUpdate(ctx,
		config.GroupName(),
		nsgName,
		"DENY-OUT",
		network.SecurityRule{
			SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
				Access: network.SecurityRuleAccessDeny,
				DestinationAddressPrefix: to.StringPtr("*"),
				DestinationPortRange:     to.StringPtr("*"),
				Direction:                network.SecurityRuleDirectionOutbound,
				Description:              to.StringPtr("Deny outbound traffic"),
				Priority:                 to.Int32Ptr(100),
				Protocol:                 network.SecurityRuleProtocolAsterisk,
				SourceAddressPrefix:      to.StringPtr("*"),
				SourcePortRange:          to.StringPtr("*"),
			},
		})
	if err != nil {
		return rule, fmt.Errorf("cannot create deny out security rule: %v", err)
	}

	err = future.WaitForCompletion(ctx, rulesClient.Client)
	if err != nil {
		return rule, fmt.Errorf("cannot get security rule create or update future response: %v", err)
	}

	return future.Result(rulesClient)
}

// Load balancers

func getLBClient() network.LoadBalancersClient {
	lbClient := network.NewLoadBalancersClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	lbClient.Authorizer = a
	lbClient.AddToUserAgent(config.UserAgent())
	return lbClient
}

// GetLoadBalancer gets info on a loadbalancer
func GetLoadBalancer(ctx context.Context, lbName string) (network.LoadBalancer, error) {
	lbClient := getLBClient()
	return lbClient.Get(ctx, config.GroupName(), lbName, "")
}

// CreateLoadBalancer creates a load balancer with 2 inbound NAT rules.
func CreateLoadBalancer(ctx context.Context, lbName, pipName string) (lb network.LoadBalancer, err error) {
	probeName := "probe"
	frontEndIPConfigName := "fip"
	backEndAddressPoolName := "backEndPool"
	idPrefix := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/loadBalancers", config.SubscriptionID(), config.GroupName())

	pip, err := GetPublicIP(ctx, pipName)
	if err != nil {
		return
	}

	lbClient := getLBClient()
	future, err := lbClient.CreateOrUpdate(ctx,
		config.GroupName(),
		lbName,
		network.LoadBalancer{
			Location: to.StringPtr(config.Location()),
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
						Name: &backEndAddressPoolName,
					},
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
					{
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
					{
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
