package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/network/armnetwork"

	"github.com/Azure/azure-sdk-for-go/sdk/armcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/to"
)

var (
	subscriptionId string
	location       = "westus2"
)

// Instance
var (
	resourceGroupName = "sample-resource-group"
	vmName            = "sample-vm"
	vnetName          = "sample-vnet"
	subnetName        = "sample-subnet"
	nsgName           = "sample-nsg"
	nicName           = "sample-nic"
	diskName          = "sample-disk"
	publicIPName      = "sample-public-ip"
)

func main() {
	subscriptionId = os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionId) == 0 {
		log.Println("AZURE_SUBSCRIPTION_ID is not set.")
		return
	}
	log.Println("AZURE_SUBSCRIPTION_ID:", subscriptionId)
	//create virtual machine
	createVM()

	//delete virtual machine
	// deleteVM()
}

func createVM() {

	conn, err := connectionAzure()
	handlerErr(err)
	ctx := context.Background()

	log.Println("start creating virtual machine...")
	resourceGroup, err := createResourceGroup(ctx, conn, resourceGroupName)
	handlerErr(err)
	toString(resourceGroup)

	virtualNetwork, err := createVirtualNetwork(ctx, conn)
	handlerErr(err)
	toString(virtualNetwork)

	subnet, err := createSubnets(ctx, conn, vnetName, subnetName)
	handlerErr(err)
	toString(subnet)

	publicIP, err := createPublicIP(ctx, conn, publicIPName)
	handlerErr(err)
	toString(publicIP)

	// network security group
	nsg, err := createNetworkSecurityGroup(ctx, conn, nsgName)
	handlerErr(err)
	toString(nsg)

	subnetID := subnet.ID
	publicID := publicIP.ID
	networkSecurityGroupID := nsg.ID
	netWorkInterface, err := createNetWorkInterface(ctx, conn, subnetID, publicID, networkSecurityGroupID)
	handlerErr(err)
	toString(netWorkInterface)

	networkInterfaceID := netWorkInterface.ID
	virtualMachine, err := createVirtualMachine(ctx, conn, *networkInterfaceID)
	handlerErr(err)
	toString(virtualMachine)

	log.Println("success created virtual machine")
}

func deleteVM() {
	conn, err := connectionAzure()
	handlerErr(err)
	ctx := context.Background()

	log.Println("start deleting virtual machine...")
	_, err = deleteVirtualMachine(ctx, conn)
	handlerErr(err)
	log.Println("deleted virtual machine")

	_, err = deleteDisk(ctx, conn, diskName)
	handlerErr(err)
	log.Println("deleted disk")

	_, err = deleteNetWorkInterface(ctx, conn)
	handlerErr(err)
	log.Println("deleted network interface")

	_, err = deleteNetworkSecurityGroup(ctx, conn, nsgName)
	handlerErr(err)
	log.Println("deleted network security group")

	_, err = deletePublicIP(ctx, conn, publicIPName)
	handlerErr(err)
	log.Println("deleted public IP")

	_, err = deleteSubnets(ctx, conn, vnetName, subnetName)
	handlerErr(err)
	log.Println("deleted subnet")

	_, err = deleteVirtualNetWork(ctx, conn)
	handlerErr(err)
	log.Println("deleted virtual network")

	_, err = deleteResourceGroup(ctx, conn)
	handlerErr(err)
	log.Println("deleted resource group")
	log.Println("success deleted virtual machine.")
}

func handlerErr(err error) {
	if err != nil {
		panic(err)
	}
}

func toString(marshaler json.Marshaler) {
	data, err := marshaler.MarshalJSON()
	handlerErr(err)

	var str string
	var id string
	switch marshaler.(type) {
	case *armresources.ResourceGroup:
		var x = marshaler.(*armresources.ResourceGroup)
		id = *x.ID
		str = "ResourceGroup"
	case *armnetwork.VirtualNetwork:
		var x = marshaler.(*armnetwork.VirtualNetwork)
		id = *x.ID
		str = "VirtualNetwork"
	case *armnetwork.Subnet:
		var x = marshaler.(*armnetwork.Subnet)
		id = *x.ID
		str = "Subnet"
	case *armnetwork.PublicIPAddress:
		var x = marshaler.(*armnetwork.PublicIPAddress)
		id = *x.ID
		str = "PublicIPAddress"
	case *armnetwork.NetworkSecurityGroup:
		var x = marshaler.(*armnetwork.NetworkSecurityGroup)
		id = *x.ID
		str = "NetworkSecurityGroup"
	case *armnetwork.NetworkInterface:
		var x = marshaler.(*armnetwork.NetworkInterface)
		id = *x.ID
		str = "NetworkInterface"
	case *armcompute.VirtualMachine:
		var x = marshaler.(*armcompute.VirtualMachine)
		id = *x.ID
		str = "VirtualMachine"
	default:
		id = ""
		str = "Azure"
	}
	log.Printf("%s:\n\t%s\n\t%s", str, string(data), id)
}

func connectionAzure() (*armcore.Connection, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	conn := armcore.NewDefaultConnection(cred, &armcore.ConnectionOptions{
		Logging: azcore.LogOptions{
			IncludeBody: true,
		},
	})
	return conn, nil
}

func createResourceGroup(ctx context.Context, connection *armcore.Connection, rgName string) (*armresources.ResourceGroup, error) {
	resourceGroup := armresources.NewResourceGroupsClient(connection, subscriptionId)

	parameters := armresources.ResourceGroup{
		Location: to.StringPtr(location),
		Tags:     map[string]*string{"sample-rs-tag": to.StringPtr("sample-tag")}, // resource group update tags
	}

	resp, err := resourceGroup.CreateOrUpdate(ctx, rgName, parameters, nil)
	if err != nil {
		return nil, err
	}

	return resp.ResourceGroup, nil
}

func deleteResourceGroup(ctx context.Context, connection *armcore.Connection) (*http.Response, error) {
	resourceGroup := armresources.NewResourceGroupsClient(connection, subscriptionId)

	pollerResponse, err := resourceGroup.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func createVirtualNetwork(ctx context.Context, connection *armcore.Connection) (*armnetwork.VirtualNetwork, error) {
	vnetClient := armnetwork.NewVirtualNetworksClient(connection, subscriptionId)

	parameters := armnetwork.VirtualNetwork{
		Resource: armnetwork.Resource{
			Location: to.StringPtr(location),
		},
		Properties: &armnetwork.VirtualNetworkPropertiesFormat{
			AddressSpace: &armnetwork.AddressSpace{
				AddressPrefixes: []*string{
					to.StringPtr("10.1.0.0/16"), // example 10.1.0.0/16
				},
			},
			//Subnets: []*armnetwork.Subnet{
			//	{
			//		Name: to.StringPtr(subnetName+"3"),
			//		Properties: &armnetwork.SubnetPropertiesFormat{
			//			AddressPrefix: to.StringPtr("10.1.0.0/24"),
			//		},
			//	},
			//},
		},
	}

	pollerResponse, err := vnetClient.BeginCreateOrUpdate(ctx, resourceGroupName, vnetName, parameters, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}

	return resp.VirtualNetwork, nil
}

func deleteVirtualNetWork(ctx context.Context, connection *armcore.Connection) (*http.Response, error) {
	vnetClient := armnetwork.NewVirtualNetworksClient(connection, subscriptionId)

	pollerResponse, err := vnetClient.BeginDelete(ctx, resourceGroupName, vnetName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func createSubnets(ctx context.Context, connection *armcore.Connection, virtualNetworkName string, subnetName string) (*armnetwork.Subnet, error) {
	subnetClient := armnetwork.NewSubnetsClient(connection, subscriptionId)

	parameters := armnetwork.Subnet{
		Properties: &armnetwork.SubnetPropertiesFormat{
			AddressPrefix: to.StringPtr("10.1.10.0/24"),
		},
	}

	pollerResponse, err := subnetClient.BeginCreateOrUpdate(ctx, resourceGroupName, virtualNetworkName, subnetName, parameters, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}

	return resp.Subnet, nil
}

func deleteSubnets(ctx context.Context, connection *armcore.Connection, virtualNetworkName string, subnetName string) (*http.Response, error) {
	subnetClient := armnetwork.NewSubnetsClient(connection, subscriptionId)

	pollerResponse, err := subnetClient.BeginDelete(ctx, resourceGroupName, virtualNetworkName, subnetName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func listSubnets(ctx context.Context, connection *armcore.Connection, virtualNetworkName string) ([]*armnetwork.Subnet, error) {
	subnetClient := armnetwork.NewSubnetsClient(connection, subscriptionId)

	resultPager := subnetClient.List(resourceGroupName, virtualNetworkName, nil)

	subnetLists := make([]*armnetwork.Subnet, 0)
	for resultPager.NextPage(ctx) {
		pageResponse := resultPager.PageResponse()
		subnetLists = append(subnetLists, pageResponse.SubnetListResult.Value...)
	}

	return subnetLists, nil
}

func createNetworkSecurityGroup(ctx context.Context, connection *armcore.Connection, nsgName string) (*armnetwork.NetworkSecurityGroup, error) {
	nsgClient := armnetwork.NewNetworkSecurityGroupsClient(connection, subscriptionId)

	parameters := armnetwork.NetworkSecurityGroup{
		Resource: armnetwork.Resource{
			Location: to.StringPtr(location),
		},
		Properties: &armnetwork.NetworkSecurityGroupPropertiesFormat{
			SecurityRules: []*armnetwork.SecurityRule{
				// window connection to virtual machine needs to open port 3389,RDP
				// inbound
				{
					Name: to.StringPtr("sample_inbound_22"), //
					Properties: &armnetwork.SecurityRulePropertiesFormat{
						SourceAddressPrefix:      to.StringPtr("0.0.0.0/0"),
						SourcePortRange:          to.StringPtr("*"),
						DestinationAddressPrefix: to.StringPtr("0.0.0.0/0"),
						DestinationPortRange:     to.StringPtr("22"),
						Protocol:                 armnetwork.SecurityRuleProtocolTCP.ToPtr(),
						Access:                   armnetwork.SecurityRuleAccessAllow.ToPtr(),
						Priority:                 to.Int32Ptr(100),
						Description:              to.StringPtr("sample network security group inbound port 22"),
						Direction:                armnetwork.SecurityRuleDirectionInbound.ToPtr(),
					},
				},
				// outbound
				{
					Name: to.StringPtr("sample_outbound_22"), //
					Properties: &armnetwork.SecurityRulePropertiesFormat{
						SourceAddressPrefix:      to.StringPtr("0.0.0.0/0"),
						SourcePortRange:          to.StringPtr("*"),
						DestinationAddressPrefix: to.StringPtr("0.0.0.0/0"),
						DestinationPortRange:     to.StringPtr("22"),
						Protocol:                 armnetwork.SecurityRuleProtocolTCP.ToPtr(),
						Access:                   armnetwork.SecurityRuleAccessAllow.ToPtr(),
						Priority:                 to.Int32Ptr(100),
						Description:              to.StringPtr("sample network security group outbound port 22"),
						Direction:                armnetwork.SecurityRuleDirectionOutbound.ToPtr(),
					},
				},
			},
		},
	}

	pollerResponse, err := nsgClient.BeginCreateOrUpdate(ctx, resourceGroupName, nsgName, parameters, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp.NetworkSecurityGroup, nil
}

func deleteNetworkSecurityGroup(ctx context.Context, connection *armcore.Connection, nsgName string) (*http.Response, error) {
	nsgClient := armnetwork.NewNetworkSecurityGroupsClient(connection, subscriptionId)

	pollerResponse, err := nsgClient.BeginDelete(ctx, resourceGroupName, nsgName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func createPublicIP(ctx context.Context, connection *armcore.Connection, publicIPName string) (*armnetwork.PublicIPAddress, error) {
	publicIPAddressClient := armnetwork.NewPublicIPAddressesClient(connection, subscriptionId)

	parameters := armnetwork.PublicIPAddress{
		Resource: armnetwork.Resource{
			Location: to.StringPtr(location),
		},
		Properties: &armnetwork.PublicIPAddressPropertiesFormat{
			PublicIPAllocationMethod: armnetwork.IPAllocationMethodStatic.ToPtr(), // Static or Dynamic
		},
	}

	pollerResponse, err := publicIPAddressClient.BeginCreateOrUpdate(ctx, resourceGroupName, publicIPName, parameters, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp.PublicIPAddress, err
}

func deletePublicIP(ctx context.Context, connection *armcore.Connection, publicIPName string) (*http.Response, error) {
	publicIPAddressClient := armnetwork.NewPublicIPAddressesClient(connection, subscriptionId)

	pollerResponse, err := publicIPAddressClient.BeginDelete(ctx, resourceGroupName, publicIPName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp, err
}

func createNetWorkInterface(ctx context.Context, connection *armcore.Connection, subnetID *string, publicIPID *string, networkSecurityGroupID *string) (*armnetwork.NetworkInterface, error) {

	nicClient := armnetwork.NewNetworkInterfacesClient(connection, subscriptionId)

	parameters := armnetwork.NetworkInterface{
		Resource: armnetwork.Resource{
			Location: to.StringPtr(location),
		},
		Properties: &armnetwork.NetworkInterfacePropertiesFormat{
			//NetworkSecurityGroup:
			IPConfigurations: []*armnetwork.NetworkInterfaceIPConfiguration{
				{
					Name: to.StringPtr("ipConfig"),
					Properties: &armnetwork.NetworkInterfaceIPConfigurationPropertiesFormat{
						PrivateIPAllocationMethod: armnetwork.IPAllocationMethodDynamic.ToPtr(),
						Subnet: &armnetwork.Subnet{
							SubResource: armnetwork.SubResource{
								ID: subnetID,
							},
						},
						PublicIPAddress: &armnetwork.PublicIPAddress{
							Resource: armnetwork.Resource{
								ID: publicIPID,
							},
						},
					},
				},
			},
			NetworkSecurityGroup: &armnetwork.NetworkSecurityGroup{
				Resource: armnetwork.Resource{
					ID: networkSecurityGroupID,
				},
			},
		},
	}

	pollerResponse, err := nicClient.BeginCreateOrUpdate(ctx, resourceGroupName, nicName, parameters, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}

	return resp.NetworkInterface, err
}

func deleteNetWorkInterface(ctx context.Context, connection *armcore.Connection) (*http.Response, error) {
	nicClient := armnetwork.NewNetworkInterfacesClient(connection, subscriptionId)

	pollerResponse, err := nicClient.BeginDelete(ctx, resourceGroupName, nicName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}

	return resp, err
}

func createVirtualMachine(ctx context.Context, connection *armcore.Connection, networkInterfaceID string) (*armcompute.VirtualMachine, error) {
	vmClient := armcompute.NewVirtualMachinesClient(connection, subscriptionId)

	//require ssh key for authentication on linux
	//sshPublicKeyPath := "/home/user/.ssh/id_rsa.pub"
	//var sshBytes []byte
	//_,err := os.Stat(sshPublicKeyPath)
	//if err == nil {
	//	sshBytes,err = ioutil.ReadFile(sshPublicKeyPath)
	//	if err != nil {
	//		return nil, err
	//	}
	//}

	parameters := armcompute.VirtualMachine{
		Resource: armcompute.Resource{
			Location: to.StringPtr(location),
		},
		Identity: &armcompute.VirtualMachineIdentity{
			Type: armcompute.ResourceIdentityTypeNone.ToPtr(),
		},
		Properties: &armcompute.VirtualMachineProperties{
			StorageProfile: &armcompute.StorageProfile{
				ImageReference: &armcompute.ImageReference{
					// search image reference
					// az vm image list --output table
					Offer:     to.StringPtr("WindowsServer"),
					Publisher: to.StringPtr("MicrosoftWindowsServer"),
					SKU:       to.StringPtr("2019-Datacenter"),
					Version:   to.StringPtr("latest"),
					//require ssh key for authentication on linux
					//Offer:     to.StringPtr("UbuntuServer"),
					//Publisher: to.StringPtr("Canonical"),
					//SKU:       to.StringPtr("18.04-LTS"),
					//Version:   to.StringPtr("latest"),
				},
				OSDisk: &armcompute.OSDisk{
					Name:         to.StringPtr(diskName),
					CreateOption: armcompute.DiskCreateOptionTypesFromImage.ToPtr(),
					Caching:      armcompute.CachingTypesReadWrite.ToPtr(),
					ManagedDisk: &armcompute.ManagedDiskParameters{
						StorageAccountType: armcompute.StorageAccountTypesStandardLRS.ToPtr(), // OSDisk type Standard/Premium HDD/SSD
					},
					//DiskSizeGB: to.Int32Ptr(100), // default 127G
				},
			},
			HardwareProfile: &armcompute.HardwareProfile{
				VMSize: armcompute.VirtualMachineSizeTypesStandardF2S.ToPtr(), // VM size include vCPUs,RAM,Data Disks,Temp storage.
			},
			OSProfile: &armcompute.OSProfile{ //
				ComputerName:  to.StringPtr("sample-compute"),
				AdminUsername: to.StringPtr("sample-user"),
				AdminPassword: to.StringPtr("Password01!@#"),
				//require ssh key for authentication on linux
				//LinuxConfiguration: &armcompute.LinuxConfiguration{
				//	DisablePasswordAuthentication: to.BoolPtr(true),
				//	SSH: &armcompute.SSHConfiguration{
				//		PublicKeys: []*armcompute.SSHPublicKey{
				//			{
				//				Path:    to.StringPtr(fmt.Sprintf("/home/%s/.ssh/authorized_keys", "sample-user")),
				//				KeyData: to.StringPtr(string(sshBytes)),
				//			},
				//		},
				//	},
				//},
			},
			NetworkProfile: &armcompute.NetworkProfile{
				NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
					{
						SubResource: armcompute.SubResource{
							ID: to.StringPtr(networkInterfaceID),
						},
					},
				},
			},
		},
	}

	pollerResponse, err := vmClient.BeginCreateOrUpdate(ctx, resourceGroupName, vmName, parameters, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}

	return resp.VirtualMachine, nil
}

func deleteVirtualMachine(ctx context.Context, connection *armcore.Connection) (*http.Response, error) {
	vmClient := armcompute.NewVirtualMachinesClient(connection, subscriptionId)

	pollerResponse, err := vmClient.BeginDelete(ctx, resourceGroupName, vmName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func createDisk(ctx context.Context, connection *armcore.Connection, diskName string, diskSize int32) (*armcompute.Disk, error) {
	diskClient := armcompute.NewDisksClient(connection, subscriptionId)

	disk := armcompute.Disk{
		Resource: armcompute.Resource{
			Location: to.StringPtr(location),
		},
		Properties: &armcompute.DiskProperties{
			CreationData: &armcompute.CreationData{
				CreateOption: armcompute.DiskCreateOptionEmpty.ToPtr(), // create
			},
			DiskSizeGB: to.Int32Ptr(diskSize),
		},
	}

	pollerResponse, err := diskClient.BeginCreateOrUpdate(ctx, resourceGroupName, diskName, disk, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp.Disk, nil
}

func deleteDisk(ctx context.Context, connection *armcore.Connection, diskName string) (*http.Response, error) {
	diskClient := armcompute.NewDisksClient(connection, subscriptionId)

	pollerResponse, err := diskClient.BeginDelete(ctx, resourceGroupName, diskName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
