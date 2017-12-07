package network

import (
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2017-09-01/network"
	"github.com/subosito/gotenv"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

var (
	virtualNetworkName string
	subnet1Name        string
	subnet2Name        string
	nsgName            string
	nicName            string
	ip1Name            string
	clients            map[string]interface{}
)

func init() {
	gotenv.Load()
	virtualNetworkName = helpers.GetEnvVarOrFail("AZURE_VNET_NAME")
	nsgName = "basic_services"
	nicName = "nic1"
	subnet1Name = "subnet1"
	subnet2Name = "subnet2"
	ip1Name = "ip1"
	clients = make(map[string]interface{})
}

func getNetworkClients() (map[string]interface{}, error) {
	if len(clients) > 0 {
		return clients, nil
	}

	token, err := iam.GetResourceManagementToken(iam.OAuthGrantTypeServicePrincipal)
	if err != nil {
		log.Fatalf("%s: %v", "failed to get auth token", err)
	}

	vnetClient := network.NewVirtualNetworksClient(helpers.SubscriptionID)
	vnetClient.Authorizer = autorest.NewBearerAuthorizer(token)
	clients["vnet"] = vnetClient

	subnetClient := network.NewSubnetsClient(helpers.SubscriptionID)
	subnetClient.Authorizer = autorest.NewBearerAuthorizer(token)
	clients["subnet"] = subnetClient

	nsgClient := network.NewSecurityGroupsClient(helpers.SubscriptionID)
	nsgClient.Authorizer = autorest.NewBearerAuthorizer(token)
	clients["nsg"] = nsgClient

	ipAddressClient := network.NewPublicIPAddressesClient(helpers.SubscriptionID)
	ipAddressClient.Authorizer = autorest.NewBearerAuthorizer(token)
	clients["ip"] = ipAddressClient

	nicClient := network.NewInterfacesClient(helpers.SubscriptionID)
	nicClient.Authorizer = autorest.NewBearerAuthorizer(token)
	clients["nic"] = nicClient

	return clients, nil
}

// Vnets

func CreateVirtualNetwork() (<-chan network.VirtualNetwork, <-chan error) {
	clients, _ := getNetworkClients()
	vnetClient, _ := clients["vnet"].(network.VirtualNetworksClient)

	return vnetClient.CreateOrUpdate(
		helpers.ResourceGroupName,
		virtualNetworkName,
		network.VirtualNetwork{
			Location: to.StringPtr(helpers.Location),
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
func DeleteVirtualNetwork() (<-chan autorest.Response, <-chan error) {
	clients, _ := getNetworkClients()
	vnetClient, _ := clients["vnet"].(network.VirtualNetworksClient)

	return vnetClient.Delete(helpers.ResourceGroupName, virtualNetworkName, nil)
}

// VNet Subnets

func CreateVirtualNetworkSubnet() {}
func DeleteVirtualNetworkSubnet() {}
func GetVirtualNetworkSubnet(_vnetName string, _subnetName string) (network.Subnet, error) {
	clients, _ := getNetworkClients()
	subnetClient, _ := clients["subnet"].(network.SubnetsClient)

	return subnetClient.Get(helpers.ResourceGroupName, _vnetName, _subnetName, "")
}

// Network Security Groups

func CreateNetworkSecurityGroup() (<-chan network.SecurityGroup, <-chan error) {
	clients, _ := getNetworkClients()
	nsgClient, _ := clients["nsg"].(network.SecurityGroupsClient)

	return nsgClient.CreateOrUpdate(
		helpers.ResourceGroupName,
		nsgName,
		network.SecurityGroup{
			Location: to.StringPtr(helpers.Location),
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

func DeleteNetworkSecurityGroup() (<-chan autorest.Response, <-chan error) {
	clients, _ := getNetworkClients()
	nsgClient, _ := clients["nsg"].(network.SecurityGroupsClient)

	return nsgClient.Delete(helpers.ResourceGroupName, nsgName, nil)
}

func GetNetworkSecurityGroup(_nsgName string) (network.SecurityGroup, error) {
	clients, _ := getNetworkClients()
	nsgClient, _ := clients["nsg"].(network.SecurityGroupsClient)

	return nsgClient.Get(helpers.ResourceGroupName, _nsgName, "")
}

// Network Security Group Rules

func CreateNetworkSecurityGroupRule() {}
func DeleteNetworkSecurityGroupRule() {}

// Network Interfaces (NIC's)

func CreateNic() (<-chan network.Interface, <-chan error) {
	clients, _ := getNetworkClients()
	nicClient, _ := clients["nic"].(network.InterfacesClient)

	nsg, err := GetNetworkSecurityGroup(nsgName)
	if err != nil {
		log.Fatalf("failed to get nsg: %v", err)
	}

	subnet, err := GetVirtualNetworkSubnet(virtualNetworkName, subnet1Name)
	if err != nil {
		log.Fatalf("failed to get subnet: %v", err)
	}

	ip, err := GetPublicIp(ip1Name)
	if err != nil {
		log.Fatalf("failed to get ip address: %v", err)
	}

	return nicClient.CreateOrUpdate(
		helpers.ResourceGroupName,
		nicName,
		network.Interface{
			Name:     to.StringPtr(nicName),
			Location: to.StringPtr(helpers.Location),
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

func DeleteNic() (<-chan autorest.Response, <-chan error) {
	clients, _ := getNetworkClients()
	nicClient, _ := clients["nic"].(network.InterfacesClient)

	return nicClient.Delete(helpers.ResourceGroupName, nicName, nil)
}

// Public IP Addresses

func CreatePublicIp() (<-chan network.PublicIPAddress, <-chan error) {
	clients, _ := getNetworkClients()
	ipClient, _ := clients["ip"].(network.PublicIPAddressesClient)

	return ipClient.CreateOrUpdate(
		helpers.ResourceGroupName,
		ip1Name,
		network.PublicIPAddress{
			Name:     to.StringPtr(ip1Name),
			Location: to.StringPtr(helpers.Location),
			PublicIPAddressPropertiesFormat: &network.PublicIPAddressPropertiesFormat{
				PublicIPAddressVersion:   network.IPv4,
				PublicIPAllocationMethod: network.Static,
			},
		},
		nil,
	)
}

func DeletePublicIp() (<-chan autorest.Response, <-chan error) {
	clients, _ := getNetworkClients()
	ipClient, _ := clients["ip"].(network.PublicIPAddressesClient)

	return ipClient.Delete(helpers.ResourceGroupName, ip1Name, nil)
}

func GetPublicIp(_ipName string) (network.PublicIPAddress, error) {
	clients, _ := getNetworkClients()
	ipClient, _ := clients["ip"].(network.PublicIPAddressesClient)

	return ipClient.Get(helpers.ResourceGroupName, _ipName, "")
}
