package main

import (
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/arm/network"
	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
)

var (
	subscriptionID       string
	spToken              *adal.ServicePrincipalToken
	tenantID             string
	clientID             string
	networkGatewayClient network.VirtualNetworkGatewaysClient
	resourceGroupClient  resources.GroupsClient
	vNetClient           network.VirtualNetworksClient
	pipClient            network.PublicIPAddressesClient
	gatewaySubnet        network.Subnet

	vnetGatewayName string
	location        string
	resGroup        string
	gwSubnetID      string
	pipID           string
	pipName         string
	vnetName        string
)

func init() {

	fmt.Println("hello")

	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	tenantID = os.Getenv("AZURE_TENANT_ID")

	fmt.Print("OAuthConfig for tenant...")
	oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, tenantID)
	onErrorFail(err, "OAuthConfigForTenant failed")

	clientID = os.Getenv("AZURE_CLIENT_ID")
	clientSecret := os.Getenv("AZURE_CLIENT_SECRET")

	fmt.Print("New service principal token...")
	spToken, err = adal.NewServicePrincipalToken(*oauthConfig, clientID, clientSecret, azure.PublicCloud.ResourceManagerEndpoint)
	onErrorFail(err, "NewServicePrincipalToken failed")

	createClients()
}

func main() {

	vnetGatewayName = "net1gw"
	location = "westus"
	resGroup = "networkRg1"
	vnetName = "network1"
	pipName = "gwPip1"

	createResourceGroup()
	createVirtualNetwork()
	createPublicIP()

	createGateway()
	deleteGateway()

	deletePublicIP()
	deleteVirtualNetwork()
	deleteResourceGroup()

	fmt.Println("goodbye")
}

func createResourceGroup() {

	fmt.Print("Creating resource group... ")

	rgParms := resources.Group{
		Location: &location,
	}
	_, err := resourceGroupClient.CreateOrUpdate(resGroup, rgParms)
	onErrorFail(err, "failed")
}

func createVirtualNetwork() {

	fmt.Print("Creating virtual network... ")

	subnet1 := network.Subnet{
		Name: to.StringPtr("subnet1"),
		SubnetPropertiesFormat: &network.SubnetPropertiesFormat{
			AddressPrefix: to.StringPtr("10.0.1.0/24"),
		},
	}
	gatewaySubnet = network.Subnet{
		Name: to.StringPtr("GatewaySubnet"),
		SubnetPropertiesFormat: &network.SubnetPropertiesFormat{
			AddressPrefix: to.StringPtr("10.0.255.0/27"),
		},
	}
	vNetParameters := network.VirtualNetwork{
		Location: &location,
		VirtualNetworkPropertiesFormat: &network.VirtualNetworkPropertiesFormat{
			AddressSpace: &network.AddressSpace{
				AddressPrefixes: &[]string{"10.0.0.0/16"},
			},
			Subnets: &[]network.Subnet{subnet1, gatewaySubnet},
		},
	}
	_, errChan := vNetClient.CreateOrUpdate(resGroup, vnetName, vNetParameters, nil)
	onErrorFail(<-errChan, "failed")

	fmt.Print("Reading virtual network... ")
	result, err := vNetClient.Get(resGroup, vnetName, "")
	onErrorFail(err, "failed")

	if result.ID == nil {
		fmt.Println("Failed to read virtual network")
	}
	subnets := *result.VirtualNetworkPropertiesFormat.Subnets
	subnet := subnets[1]
	gwSubnetID = *subnet.ID

}

func createPublicIP() {

	fmt.Print("Creating public IP... ")

	pip := network.PublicIPAddress{
		Name:     &pipName,
		Location: &location,
		PublicIPAddressPropertiesFormat: &network.PublicIPAddressPropertiesFormat{
			PublicIPAllocationMethod: network.Dynamic,
		},
	}

	_, errChan := pipClient.CreateOrUpdate(resGroup, pipName, pip, nil)
	onErrorFail(<-errChan, "failed to create")

	fmt.Print("Reading public IP... ")
	result, err := pipClient.Get(resGroup, pipName, "")
	onErrorFail(err, "failed")

	if result.ID == nil {
		fmt.Printf("Cannot read Public IP %s (resource group %s) ID", pipName, resGroup)
	}
	pipID = *result.ID
}

func createGateway() {

	fmt.Print("Creating gateway... ")

	IpConfigProps := network.VirtualNetworkGatewayIPConfigurationPropertiesFormat{
		PrivateIPAllocationMethod: network.Dynamic,
		Subnet:          &network.SubResource{&gwSubnetID},
		PublicIPAddress: &network.SubResource{&pipID},
	}

	gwIpConfig1 := network.VirtualNetworkGatewayIPConfiguration{
		Name: to.StringPtr("gwIpConfig"),
		VirtualNetworkGatewayIPConfigurationPropertiesFormat: &IpConfigProps,
	}

	props := network.VirtualNetworkGatewayPropertiesFormat{
		IPConfigurations: &[]network.VirtualNetworkGatewayIPConfiguration{gwIpConfig1},
		GatewayType:      network.VirtualNetworkGatewayTypeVpn,
		VpnType:          network.RouteBased,
		EnableBgp:        to.BoolPtr(false),
		ActiveActive:     to.BoolPtr(false),
		Sku:              &network.VirtualNetworkGatewaySku{network.VirtualNetworkGatewaySkuNameVpnGw1, network.VirtualNetworkGatewaySkuTierVpnGw1, to.Int32Ptr(2)},
	}

	vnetGateway := network.VirtualNetworkGateway{
		Name:     &vnetGatewayName,
		Location: &location,
		VirtualNetworkGatewayPropertiesFormat: &props,
	}

	_, errChan := networkGatewayClient.CreateOrUpdate(resGroup, vnetGatewayName, vnetGateway, make(chan struct{}))
	onErrorFail(<-errChan, "failed")
}

func deleteGateway() {

	fmt.Print("Deleting gateway...")

	_, errChan := networkGatewayClient.Delete(resGroup, vnetGatewayName, make(chan struct{}))
	onErrorFail(<-errChan, "failed")

}

func deletePublicIP() {

	fmt.Print("Deleting public ip... ")

	_, errChan := pipClient.Delete(resGroup, pipName, nil)
	onErrorFail(<-errChan, "failed")
}

func deleteVirtualNetwork() {

	fmt.Print("Deleting virtual network... ")

	_, errChan := vNetClient.Delete(resGroup, vnetName, nil)
	onErrorFail(<-errChan, "failed")
}

func deleteResourceGroup() {

	fmt.Print("Deleting resource group... ")

	_, errChan := resourceGroupClient.Delete(resGroup, nil)
	onErrorFail(<-errChan, "failed")
}

func onErrorFail(err error, message string) {

	if err != nil {
		fmt.Printf("%s: %s", message, err)
		os.Exit(1)
	} else {
		fmt.Println("succeeded")
	}
}

func createClients() {

	networkGatewayClient = network.NewVirtualNetworkGatewaysClient(subscriptionID)
	networkGatewayClient.Authorizer = autorest.NewBearerAuthorizer(spToken)

	resourceGroupClient = resources.NewGroupsClient(subscriptionID)
	resourceGroupClient.Authorizer = autorest.NewBearerAuthorizer(spToken)

	vNetClient = network.NewVirtualNetworksClient(subscriptionID)
	vNetClient.Authorizer = autorest.NewBearerAuthorizer(spToken)

	pipClient = network.NewPublicIPAddressesClient(subscriptionID)
	pipClient.Authorizer = autorest.NewBearerAuthorizer(spToken)
}
