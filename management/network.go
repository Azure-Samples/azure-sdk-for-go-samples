package management

import (
	"log"

	"github.com/Azure/azure-sdk-for-go/profiles/preview/network/mgmt/network"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

// Vnets

func getVnetClient() network.VirtualNetworksClient {
	vnetClient := network.NewVirtualNetworksClient(subscriptionId)
	vnetClient.Authorizer = token
	return vnetClient
}

func CreateVirtualNetworkAndSubnets(vNet, subnet1, subnet2 string) (<-chan network.VirtualNetwork, <-chan error) {
	vnetClient := getVnetClient()
	return vnetClient.CreateOrUpdate(
		resourceGroupName,
		vNet,
		network.VirtualNetwork{
			Location: to.StringPtr(location),
			VirtualNetworkPropertiesFormat: &network.VirtualNetworkPropertiesFormat{
				AddressSpace: &network.AddressSpace{
					AddressPrefixes: &[]string{"10.0.0.0/8"},
				},
				Subnets: &[]network.Subnet{
					{
						Name: to.StringPtr(subnet1),
						SubnetPropertiesFormat: &network.SubnetPropertiesFormat{
							AddressPrefix: to.StringPtr("10.0.0.0/16"),
						},
					},
					{
						Name: to.StringPtr(subnet2),
						SubnetPropertiesFormat: &network.SubnetPropertiesFormat{
							AddressPrefix: to.StringPtr("10.1.0.0/16"),
						},
					},
				},
			},
		},
		nil)
}
func DeleteVirtualNetwork(vNet string) (<-chan autorest.Response, <-chan error) {
	vnetClient := getVnetClient()
	return vnetClient.Delete(resourceGroupName, vNet, nil)
}

// VNet Subnets

func CreateVirtualNetworkSubnet() {}
func DeleteVirtualNetworkSubnet() {}
func GetVirtualNetworkSubnet(_vnetName string, _subnetName string) (network.Subnet, error) {
	subnetClient := network.NewSubnetsClient(subscriptionId)
	subnetClient.Authorizer = token

	return subnetClient.Get(resourceGroupName, _vnetName, _subnetName, "")
}

// Network Security Groups

func getNsgClient() network.SecurityGroupsClient {
	nsgClient := network.NewSecurityGroupsClient(subscriptionId)
	nsgClient.Authorizer = token
	return nsgClient
}

func CreateNetworkSecurityGroup(nsg string) (<-chan network.SecurityGroup, <-chan error) {
	nsgClient := getNsgClient()
	return nsgClient.CreateOrUpdate(
		resourceGroupName,
		nsg,
		network.SecurityGroup{
			Location: to.StringPtr(location),
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
		nil,
	)
}

func DeleteNetworkSecurityGroup(nsg string) (<-chan autorest.Response, <-chan error) {
	nsgClient := getNsgClient()
	return nsgClient.Delete(resourceGroupName, nsg, nil)
}

func GetNetworkSecurityGroup(_nsgName string) (network.SecurityGroup, error) {
	nsgClient := getNsgClient()
	return nsgClient.Get(resourceGroupName, _nsgName, "")
}

// Network Security Group Rules

func CreateNetworkSecurityGroupRule() {}
func DeleteNetworkSecurityGroupRule() {}

// Network Interfaces (NIC's)

func getNicClient() network.InterfacesClient {
	nicClient := network.NewInterfacesClient(subscriptionId)
	nicClient.Authorizer = token
	return nicClient
}

func CreateNic(vNetName, subnetName, nsgName, ipName, nicName string) (<-chan network.Interface, <-chan error) {
	nsg, err := GetNetworkSecurityGroup(nsgName)
	if err != nil {
		log.Fatalf("failed to get nsg: %v", err)
	}

	subnet, err := GetVirtualNetworkSubnet(vNetName, subnetName)
	if err != nil {
		log.Fatalf("failed to get subnet: %v", err)
	}

	ip, err := GetPublicIp(ipName)
	if err != nil {
		log.Fatalf("failed to get ip address: %v", err)
	}

	nicClient := getNicClient()
	return nicClient.CreateOrUpdate(
		resourceGroupName,
		nicName,
		network.Interface{
			Name:     to.StringPtr(nicName),
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
		nil,
	)
}

func DeleteNic(nicName string) (<-chan autorest.Response, <-chan error) {
	nicClient := getNicClient()
	return nicClient.Delete(resourceGroupName, nicName, nil)
}

// Public IP Addresses

func getPipClient() network.PublicIPAddressesClient {
	pipClient := network.NewPublicIPAddressesClient(subscriptionId)
	pipClient.Authorizer = token
	return pipClient
}

func CreatePublicIp(ipName string) (<-chan network.PublicIPAddress, <-chan error) {
	pipClient := getPipClient()

	return pipClient.CreateOrUpdate(
		resourceGroupName,
		ipName,
		network.PublicIPAddress{
			Name:     to.StringPtr(ipName),
			Location: to.StringPtr(location),
			PublicIPAddressPropertiesFormat: &network.PublicIPAddressPropertiesFormat{
				PublicIPAddressVersion:   network.IPv4,
				PublicIPAllocationMethod: network.Static,
			},
		},
		nil,
	)
}

func DeletePublicIp(ipName string) (<-chan autorest.Response, <-chan error) {
	pipClient := getPipClient()
	return pipClient.Delete(resourceGroupName, ipName, nil)
}

func GetPublicIp(ipName string) (network.PublicIPAddress, error) {
	pipClient := getPipClient()
	return pipClient.Get(resourceGroupName, ipName, "")
}
