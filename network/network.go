package network

import (
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2017-09-01/network"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

// Vnets

func getVnetClient() network.VirtualNetworksClient {
	token, _ := iam.GetResourceManagementToken(iam.OAuthGrantTypeServicePrincipal)
	vnetClient := network.NewVirtualNetworksClient(helpers.SubscriptionID())
	vnetClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return vnetClient
}

func CreateVirtualNetworkAndSubnets(vnetName, subnet1Name, subnet2Name string) (<-chan network.VirtualNetwork, <-chan error) {
	vnetClient := getVnetClient()
	return vnetClient.CreateOrUpdate(
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
		},
		nil)
}
func DeleteVirtualNetwork(vnetName string) (<-chan autorest.Response, <-chan error) {
	vnetClient := getVnetClient()
	return vnetClient.Delete(helpers.ResourceGroupName(), vnetName, nil)
}

// VNet Subnets

func getSubnetsClient() network.SubnetsClient {
	token, _ := iam.GetResourceManagementToken(iam.OAuthGrantTypeServicePrincipal)
	subnetsClient := network.NewSubnetsClient(helpers.SubscriptionID())
	subnetsClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return subnetsClient
}

func CreateVirtualNetworkSubnet() {}
func DeleteVirtualNetworkSubnet() {}

func GetVirtualNetworkSubnet(vnetName string, subnetName string) (network.Subnet, error) {
	subnetsClient := getSubnetsClient()
	return subnetsClient.Get(helpers.ResourceGroupName(), vnetName, subnetName, "")
}

// Network Security Groups

func getNsgClient() network.SecurityGroupsClient {
	token, _ := iam.GetResourceManagementToken(iam.OAuthGrantTypeServicePrincipal)
	nsgClient := network.NewSecurityGroupsClient(helpers.SubscriptionID())
	nsgClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return nsgClient
}

func CreateNetworkSecurityGroup(nsgName string) (<-chan network.SecurityGroup, <-chan error) {
	nsgClient := getNsgClient()
	return nsgClient.CreateOrUpdate(
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
		nil,
	)
}

func DeleteNetworkSecurityGroup(nsgName string) (<-chan autorest.Response, <-chan error) {
	nsgClient := getNsgClient()
	return nsgClient.Delete(helpers.ResourceGroupName(), nsgName, nil)
}

func GetNetworkSecurityGroup(nsgName string) (network.SecurityGroup, error) {
	nsgClient := getNsgClient()
	return nsgClient.Get(helpers.ResourceGroupName(), nsgName, "")
}

// Network Security Group Rules

func CreateNetworkSecurityGroupRule() {}
func DeleteNetworkSecurityGroupRule() {}

// Network Interfaces (NIC's)

func getNicClient() network.InterfacesClient {
	token, _ := iam.GetResourceManagementToken(iam.OAuthGrantTypeServicePrincipal)
	nicClient := network.NewInterfacesClient(helpers.SubscriptionID())
	nicClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return nicClient
}

func CreateNic(vnetName, subnetName, nsgName, ipName, nicName string) (<-chan network.Interface, <-chan error) {
	nsg, err := GetNetworkSecurityGroup(nsgName)
	if err != nil {
		log.Fatalf("failed to get nsg: %v", err)
	}

	subnet, err := GetVirtualNetworkSubnet(vnetName, subnetName)
	if err != nil {
		log.Fatalf("failed to get subnet: %v", err)
	}

	ip, err := GetPublicIp(ipName)
	if err != nil {
		log.Fatalf("failed to get ip address: %v", err)
	}

	nicClient := getNicClient()
	return nicClient.CreateOrUpdate(
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
		nil,
	)
}

func DeleteNic(nic string) (<-chan autorest.Response, <-chan error) {
	nicClient := getNicClient()
	return nicClient.Delete(helpers.ResourceGroupName(), nic, nil)
}

// Public IP Addresses

func getIpClient() network.PublicIPAddressesClient {
	token, _ := iam.GetResourceManagementToken(iam.OAuthGrantTypeServicePrincipal)
	ipClient := network.NewPublicIPAddressesClient(helpers.SubscriptionID())
	ipClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return ipClient
}

func CreatePublicIp(ipName string) (<-chan network.PublicIPAddress, <-chan error) {
	ipClient := getIpClient()
	return ipClient.CreateOrUpdate(
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
		nil,
	)
}

func DeletePublicIp(ipName string) (<-chan autorest.Response, <-chan error) {
	ipClient := getIpClient()
	return ipClient.Delete(helpers.ResourceGroupName(), ipName, nil)
}

func GetPublicIp(ipName string) (network.PublicIPAddress, error) {
	ipClient := getIpClient()
	return ipClient.Get(helpers.ResourceGroupName(), ipName, "")
}
