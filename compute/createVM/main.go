package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var subscriptionId string

const (
	resourceGroupName = "sample-resource-group"
	vmName            = "sample-vm"
	vnetName          = "sample-vnet"
	subnetName        = "sample-subnet"
	nsgName           = "sample-nsg"
	nicName           = "sample-nic"
	diskName          = "sample-disk"
	publicIPName      = "sample-public-ip"
	location          = "westus2"
)

func main() {
	subscriptionId = os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionId) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set.")
	}
	//create virtual machine
	createVM()

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		//delete virtual machine
		cleanup()
	}
}

func createVM() {
	conn, err := connectionAzure()
	if err != nil {
		log.Fatalf("cannot connect to Azure:%+v", err)
	}
	ctx := context.Background()

	log.Println("start creating virtual machine...")
	resourceGroup, err := createResourceGroup(ctx, conn)
	if err != nil {
		log.Fatalf("cannot create resource group:%+v", err)
	}
	log.Printf("Created resource group: %s", *resourceGroup.ID)

	virtualNetwork, err := createVirtualNetwork(ctx, conn)
	if err != nil {
		log.Fatalf("cannot create virtual network:%+v", err)
	}
	log.Printf("Created virtual network: %s", *virtualNetwork.ID)

	subnet, err := createSubnets(ctx, conn)
	if err != nil {
		log.Fatalf("cannot create subnet:%+v", err)
	}
	log.Printf("Created subnet: %s", *subnet.ID)

	publicIP, err := createPublicIP(ctx, conn)
	if err != nil {
		log.Fatalf("cannot create public IP address:%+v", err)
	}
	log.Printf("Created public IP address: %s", *publicIP.ID)

	// network security group
	nsg, err := createNetworkSecurityGroup(ctx, conn)
	if err != nil {
		log.Fatalf("cannot create network security group:%+v", err)
	}
	log.Printf("Created network security group: %s", *nsg.ID)

	netWorkInterface, err := createNetWorkInterface(ctx, conn, *subnet.ID, *publicIP.ID, *nsg.ID)
	if err != nil {
		log.Fatalf("cannot create network interface:%+v", err)
	}
	log.Printf("Created network interface: %s", *netWorkInterface.ID)

	networkInterfaceID := netWorkInterface.ID
	virtualMachine, err := createVirtualMachine(ctx, conn, *networkInterfaceID)
	if err != nil {
		log.Fatalf("cannot create virual machine:%+v", err)
	}
	log.Printf("Created network virual machine: %s", *virtualMachine.ID)

	log.Println("Virtual machine created successfully")
}

func cleanup() {
	conn, err := connectionAzure()
	if err != nil {
		log.Fatalf("cannot connection Azure:%+v", err)
	}
	ctx := context.Background()

	log.Println("start deleting virtual machine...")
	_, err = deleteVirtualMachine(ctx, conn)
	if err != nil {
		log.Fatalf("cannot delete virtual machine:%+v", err)
	}
	log.Println("deleted virtual machine")

	_, err = deleteDisk(ctx, conn)
	if err != nil {
		log.Fatalf("cannot delete disk:%+v", err)
	}
	log.Println("deleted disk")

	_, err = deleteNetWorkInterface(ctx, conn)
	if err != nil {
		log.Fatalf("cannot delete network interface:%+v", err)
	}
	log.Println("deleted network interface")

	_, err = deleteNetworkSecurityGroup(ctx, conn)
	if err != nil {
		log.Fatalf("cannot delete network security group:%+v", err)
	}
	log.Println("deleted network security group")

	_, err = deletePublicIP(ctx, conn)
	if err != nil {
		log.Fatalf("cannot delete public IP address:%+v", err)
	}
	log.Println("deleted public IP address")

	_, err = deleteSubnets(ctx, conn)
	if err != nil {
		log.Fatalf("cannot delete subnet:%+v", err)
	}
	log.Println("deleted subnet")

	_, err = deleteVirtualNetWork(ctx, conn)
	if err != nil {
		log.Fatalf("cannot delete virtual network:%+v", err)
	}
	log.Println("deleted virtual network")

	_, err = deleteResourceGroup(ctx, conn)
	if err != nil {
		log.Fatalf("cannot delete resource group:%+v", err)
	}
	log.Println("deleted resource group")
	log.Println("success deleted virtual machine.")
}

func connectionAzure() (azcore.TokenCredential, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	return cred, nil
}

func createResourceGroup(ctx context.Context, cred azcore.TokenCredential) (*armresources.ResourceGroup, error) {
	resourceGroup := armresources.NewResourceGroupsClient(subscriptionId, cred, nil)

	parameters := armresources.ResourceGroup{
		Location: to.StringPtr(location),
		Tags:     map[string]*string{"sample-rs-tag": to.StringPtr("sample-tag")}, // resource group update tags
	}

	resp, err := resourceGroup.CreateOrUpdate(ctx, resourceGroupName, parameters, nil)
	if err != nil {
		return nil, err
	}

	return &resp.ResourceGroup, nil
}

func deleteResourceGroup(ctx context.Context, cred azcore.TokenCredential) (*http.Response, error) {
	resourceGroup := armresources.NewResourceGroupsClient(subscriptionId, cred, nil)

	pollerResponse, err := resourceGroup.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}

	return resp.RawResponse, nil
}

func createVirtualNetwork(ctx context.Context, cred azcore.TokenCredential) (*armnetwork.VirtualNetwork, error) {
	vnetClient := armnetwork.NewVirtualNetworksClient(subscriptionId, cred, nil)

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

	return &resp.VirtualNetwork, nil
}

func deleteVirtualNetWork(ctx context.Context, cred azcore.TokenCredential) (*http.Response, error) {
	vnetClient := armnetwork.NewVirtualNetworksClient(subscriptionId, cred, nil)

	pollerResponse, err := vnetClient.BeginDelete(ctx, resourceGroupName, vnetName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}

	return resp.RawResponse, nil
}

func createSubnets(ctx context.Context, cred azcore.TokenCredential) (*armnetwork.Subnet, error) {
	subnetClient := armnetwork.NewSubnetsClient(subscriptionId, cred, nil)

	parameters := armnetwork.Subnet{
		Properties: &armnetwork.SubnetPropertiesFormat{
			AddressPrefix: to.StringPtr("10.1.10.0/24"),
		},
	}

	pollerResponse, err := subnetClient.BeginCreateOrUpdate(ctx, resourceGroupName, vnetName, subnetName, parameters, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}

	return &resp.Subnet, nil
}

func deleteSubnets(ctx context.Context, cred azcore.TokenCredential) (*http.Response, error) {
	subnetClient := armnetwork.NewSubnetsClient(subscriptionId, cred, nil)

	pollerResponse, err := subnetClient.BeginDelete(ctx, resourceGroupName, vnetName, subnetName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}

	return resp.RawResponse, nil
}

func createNetworkSecurityGroup(ctx context.Context, cred azcore.TokenCredential) (*armnetwork.NetworkSecurityGroup, error) {
	nsgClient := armnetwork.NewNetworkSecurityGroupsClient(subscriptionId, cred, nil)

	parameters := armnetwork.NetworkSecurityGroup{
		Resource: armnetwork.Resource{
			Location: to.StringPtr(location),
		},
		Properties: &armnetwork.NetworkSecurityGroupPropertiesFormat{
			SecurityRules: []*armnetwork.SecurityRule{
				// Windows connection to virtual machine needs to open port 3389,RDP
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
	return &resp.NetworkSecurityGroup, nil
}

func deleteNetworkSecurityGroup(ctx context.Context, cred azcore.TokenCredential) (*http.Response, error) {
	nsgClient := armnetwork.NewNetworkSecurityGroupsClient(subscriptionId, cred, nil)

	pollerResponse, err := nsgClient.BeginDelete(ctx, resourceGroupName, nsgName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp.RawResponse, nil
}

func createPublicIP(ctx context.Context, cred azcore.TokenCredential) (*armnetwork.PublicIPAddress, error) {
	publicIPAddressClient := armnetwork.NewPublicIPAddressesClient(subscriptionId, cred, nil)

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
	return &resp.PublicIPAddress, err
}

func deletePublicIP(ctx context.Context, cred azcore.TokenCredential) (*http.Response, error) {
	publicIPAddressClient := armnetwork.NewPublicIPAddressesClient(subscriptionId, cred, nil)

	pollerResponse, err := publicIPAddressClient.BeginDelete(ctx, resourceGroupName, publicIPName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp.RawResponse, err
}

func createNetWorkInterface(ctx context.Context, cred azcore.TokenCredential, subnetID string, publicIPID string, networkSecurityGroupID string) (*armnetwork.NetworkInterface, error) {
	nicClient := armnetwork.NewNetworkInterfacesClient(subscriptionId, cred, nil)

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
								ID: to.StringPtr(subnetID),
							},
						},
						PublicIPAddress: &armnetwork.PublicIPAddress{
							Resource: armnetwork.Resource{
								ID: to.StringPtr(publicIPID),
							},
						},
					},
				},
			},
			NetworkSecurityGroup: &armnetwork.NetworkSecurityGroup{
				Resource: armnetwork.Resource{
					ID: to.StringPtr(networkSecurityGroupID),
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

	return &resp.NetworkInterface, err
}

func deleteNetWorkInterface(ctx context.Context, cred azcore.TokenCredential) (*http.Response, error) {
	nicClient := armnetwork.NewNetworkInterfacesClient(subscriptionId, cred, nil)

	pollerResponse, err := nicClient.BeginDelete(ctx, resourceGroupName, nicName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}

	return resp.RawResponse, err
}

func createVirtualMachine(ctx context.Context, cred azcore.TokenCredential, networkInterfaceID string) (*armcompute.VirtualMachine, error) {
	vmClient := armcompute.NewVirtualMachinesClient(subscriptionId, cred, nil)

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
				VMSize: armcompute.VirtualMachineSizeTypes("Standard_F2s").ToPtr(), // VM size include vCPUs,RAM,Data Disks,Temp storage.
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

	return &resp.VirtualMachine, nil
}

func deleteVirtualMachine(ctx context.Context, cred azcore.TokenCredential) (*http.Response, error) {
	vmClient := armcompute.NewVirtualMachinesClient(subscriptionId, cred, nil)

	pollerResponse, err := vmClient.BeginDelete(ctx, resourceGroupName, vmName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}

	return resp.RawResponse, nil
}

func deleteDisk(ctx context.Context, cred azcore.TokenCredential) (*http.Response, error) {
	diskClient := armcompute.NewDisksClient(subscriptionId, cred, nil)

	pollerResponse, err := diskClient.BeginDelete(ctx, resourceGroupName, diskName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp.RawResponse, nil
}
