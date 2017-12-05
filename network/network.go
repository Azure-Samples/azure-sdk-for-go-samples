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

// Vnets

func getVnetClient() network.VirtualNetworksClient {
	vnetClient := network.NewVirtualNetworksClient(subscriptionId)
	vnetClient.Authorizer = token
	return vnetClient
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

func CreateVirtualNetworkAndSubnets(vNet, subnet1, subnet2 string) (<-chan network.VirtualNetwork, <-chan error) {
	vnetClient := getVnetClient()
	return vnetClient.CreateOrUpdate(
<<<<<<< HEAD:management/network.go
		resourceGroupName,
		vNet,
=======
		helpers.ResourceGroupName,
		virtualNetworkName,
>>>>>>> aea69c03ad165135312fd9c42e51e5b3b87a5fc6:network/network.go
		network.VirtualNetwork{
			Location: to.StringPtr(helpers.Location),
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
<<<<<<< HEAD:management/network.go
func DeleteVirtualNetwork(vNet string) (<-chan autorest.Response, <-chan error) {
	vnetClient := getVnetClient()
	return vnetClient.Delete(resourceGroupName, vNet, nil)
=======
func DeleteVirtualNetwork() (<-chan autorest.Response, <-chan error) {
	clients, _ := getNetworkClients()
	vnetClient, _ := clients["vnet"].(network.VirtualNetworksClient)

	return vnetClient.Delete(helpers.ResourceGroupName, virtualNetworkName, nil)
>>>>>>> aea69c03ad165135312fd9c42e51e5b3b87a5fc6:network/network.go
}

// VNet Subnets

func CreateVirtualNetworkSubnet() {}
func DeleteVirtualNetworkSubnet() {}
func GetVirtualNetworkSubnet(_vnetName string, _subnetName string) (network.Subnet, error) {
	subnetClient := network.NewSubnetsClient(subscriptionId)
	subnetClient.Authorizer = token

	return subnetClient.Get(helpers.ResourceGroupName, _vnetName, _subnetName, "")
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
<<<<<<< HEAD:management/network.go
		resourceGroupName,
		nsg,
=======
		helpers.ResourceGroupName,
		nsgName,
>>>>>>> aea69c03ad165135312fd9c42e51e5b3b87a5fc6:network/network.go
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

<<<<<<< HEAD:management/network.go
func DeleteNetworkSecurityGroup(nsg string) (<-chan autorest.Response, <-chan error) {
	nsgClient := getNsgClient()
	return nsgClient.Delete(resourceGroupName, nsg, nil)
}

func GetNetworkSecurityGroup(_nsgName string) (network.SecurityGroup, error) {
	nsgClient := getNsgClient()
	return nsgClient.Get(resourceGroupName, _nsgName, "")
=======
func DeleteNetworkSecurityGroup() (<-chan autorest.Response, <-chan error) {
	clients, _ := getNetworkClients()
	nsgClient, _ := clients["nsg"].(network.SecurityGroupsClient)

	return nsgClient.Delete(helpers.ResourceGroupName, nsgName, nil)
}

func GetNetworkSecurityGroup(_nsgName string) (network.SecurityGroup, error) {
	clients, _ := getNetworkClients()
	nsgClient, _ := clients["nsg"].(network.SecurityGroupsClient)

	return nsgClient.Get(helpers.ResourceGroupName, _nsgName, "")
>>>>>>> aea69c03ad165135312fd9c42e51e5b3b87a5fc6:network/network.go
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

<<<<<<< HEAD:management/network.go
func DeleteNic(nicName string) (<-chan autorest.Response, <-chan error) {
	nicClient := getNicClient()
	return nicClient.Delete(resourceGroupName, nicName, nil)
=======
func DeleteNic() (<-chan autorest.Response, <-chan error) {
	clients, _ := getNetworkClients()
	nicClient, _ := clients["nic"].(network.InterfacesClient)

	return nicClient.Delete(helpers.ResourceGroupName, nicName, nil)
>>>>>>> aea69c03ad165135312fd9c42e51e5b3b87a5fc6:network/network.go
}

// Public IP Addresses

func getPipClient() network.PublicIPAddressesClient {
	pipClient := network.NewPublicIPAddressesClient(subscriptionId)
	pipClient.Authorizer = token
	return pipClient
}

func CreatePublicIp(ipName string) (<-chan network.PublicIPAddress, <-chan error) {
	pipClient := getPipClient()

<<<<<<< HEAD:management/network.go
	return pipClient.CreateOrUpdate(
		resourceGroupName,
		ipName,
		network.PublicIPAddress{
			Name:     to.StringPtr(ipName),
			Location: to.StringPtr(location),
=======
	return ipClient.CreateOrUpdate(
		helpers.ResourceGroupName,
		ip1Name,
		network.PublicIPAddress{
			Name:     to.StringPtr(ip1Name),
			Location: to.StringPtr(helpers.Location),
>>>>>>> aea69c03ad165135312fd9c42e51e5b3b87a5fc6:network/network.go
			PublicIPAddressPropertiesFormat: &network.PublicIPAddressPropertiesFormat{
				PublicIPAddressVersion:   network.IPv4,
				PublicIPAllocationMethod: network.Static,
			},
		},
		nil,
	)
}

<<<<<<< HEAD:management/network.go
func DeletePublicIp(ipName string) (<-chan autorest.Response, <-chan error) {
	pipClient := getPipClient()
	return pipClient.Delete(resourceGroupName, ipName, nil)
}

func GetPublicIp(ipName string) (network.PublicIPAddress, error) {
	pipClient := getPipClient()
	return pipClient.Get(resourceGroupName, ipName, "")
=======
func DeletePublicIp() (<-chan autorest.Response, <-chan error) {
	clients, _ := getNetworkClients()
	ipClient, _ := clients["ip"].(network.PublicIPAddressesClient)

	return ipClient.Delete(helpers.ResourceGroupName, ip1Name, nil)
}

func GetPublicIp(_ipName string) (network.PublicIPAddress, error) {
	clients, _ := getNetworkClients()
	ipClient, _ := clients["ip"].(network.PublicIPAddressesClient)

	return ipClient.Get(helpers.ResourceGroupName, _ipName, "")
>>>>>>> aea69c03ad165135312fd9c42e51e5b3b87a5fc6:network/network.go
}
