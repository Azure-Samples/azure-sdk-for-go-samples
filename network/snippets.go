package network

import (
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/management"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2017-09-01/network"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

// Vnets

func getVnetClient() network.VirtualNetworksClient {
	vnetClient := network.NewVirtualNetworksClient(management.GetSubID())
	vnetClient.Authorizer = management.GetToken()
	return vnetClient
}

func CreateVirtualNetworkAndSubnets(vNet, subnet1, subnet2 string) (<-chan network.VirtualNetwork, <-chan error) {
	vnetClient := getVnetClient()
	return vnetClient.CreateOrUpdate(
		management.GetResourceGroup(),
		vNet,
		network.VirtualNetwork{
			Location: to.StringPtr(management.GetLocation()),
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
	return vnetClient.Delete(management.GetResourceGroup(), vNet, nil)
}

// VNet Subnets

func CreateVirtualNetworkSubnet() {}
func DeleteVirtualNetworkSubnet() {}

func GetVirtualNetworkSubnet(vnet string, subnet string) (network.Subnet, error) {
	subnetClient := network.NewSubnetsClient(management.GetSubID())
	subnetClient.Authorizer = management.GetToken()
	return subnetClient.Get(management.GetResourceGroup(), vnet, subnet, "")
}

// Network Security Groups

func getNsgClient() network.SecurityGroupsClient {
	nsgClient := network.NewSecurityGroupsClient(management.GetSubID())
	nsgClient.Authorizer = management.GetToken()
	return nsgClient
}

func CreateNetworkSecurityGroup(nsg string) (<-chan network.SecurityGroup, <-chan error) {
	nsgClient := getNsgClient()
	return nsgClient.CreateOrUpdate(
		management.GetResourceGroup(),
		nsg,
		network.SecurityGroup{
			Location: to.StringPtr(management.GetLocation()),
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
	return nsgClient.Delete(management.GetResourceGroup(), nsg, nil)
}

func GetNetworkSecurityGroup(nsg string) (network.SecurityGroup, error) {
	nsgClient := getNsgClient()
	return nsgClient.Get(management.GetResourceGroup(), nsg, "")
}

// Network Security Group Rules

func CreateNetworkSecurityGroupRule() {}
func DeleteNetworkSecurityGroupRule() {}

// Network Interfaces (NIC's)

func getNicClient() network.InterfacesClient {
	nicClient := network.NewInterfacesClient(management.GetSubID())
	nicClient.Authorizer = management.GetToken()
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
		management.GetResourceGroup(),
		nicName,
		network.Interface{
			Name:     to.StringPtr(nicName),
			Location: to.StringPtr(management.GetLocation()),
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

func DeleteNic(nic string) (<-chan autorest.Response, <-chan error) {
	nicClient := getNicClient()
	return nicClient.Delete(management.GetResourceGroup(), nic, nil)
}

// Public IP Addresses

func getPipClient() network.PublicIPAddressesClient {
	pipClient := network.NewPublicIPAddressesClient(management.GetSubID())
	pipClient.Authorizer = management.GetToken()
	return pipClient
}

func CreatePublicIp(ip string) (<-chan network.PublicIPAddress, <-chan error) {
	pipClient := getPipClient()
	return pipClient.CreateOrUpdate(
		management.GetResourceGroup(),
		ip,
		network.PublicIPAddress{
			Name:     to.StringPtr(ip),
			Location: to.StringPtr(management.GetLocation()),
			PublicIPAddressPropertiesFormat: &network.PublicIPAddressPropertiesFormat{
				PublicIPAddressVersion:   network.IPv4,
				PublicIPAllocationMethod: network.Static,
			},
		},
		nil,
	)
}

func DeletePublicIp(ip string) (<-chan autorest.Response, <-chan error) {
	pipClient := getPipClient()
	return pipClient.Delete(management.GetResourceGroup(), ip, nil)
}

func GetPublicIp(ip string) (network.PublicIPAddress, error) {
	pipClient := getPipClient()
	return pipClient.Get(management.GetResourceGroup(), ip, "")
}
