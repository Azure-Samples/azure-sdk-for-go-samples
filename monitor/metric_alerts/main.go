package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resources/armresources"
)

var (
	subscriptionID       string
	location             = "westus"
	resourceGroupName    = "sample-resource-group"
	virtualNetworkName   = "sample-virtual-network"
	subnetName           = "sample-subnet"
	publicIPAddressName  = "sample-public-ip"
	securityGroupName    = "sample-network-security-group"
	networkInterfaceName = "sample-network-interface"
	actionGroupName      = "sample-action-group"
	osDiskName           = "sample-os-disk"
	virtualMachineName   = "sample-virtual-machine"
	metricAlertName      = "sample-metric-alert"
)

func main() {
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set.")
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}

	conn := arm.NewDefaultConnection(cred, &arm.ConnectionOptions{
		Logging: policy.LogOptions{
			IncludeBody: true,
		},
	})
	ctx := context.Background()

	resourceGroup, err := createResourceGroup(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	virtualNetwork, err := createVirtualNetwork(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("virtual network:", *virtualNetwork.ID)

	subnet, err := createSubnet(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("subnet:", *subnet.ID)

	publicIP, err := createPublicIP(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("public ip:", *publicIP.ID)

	networkSecurityGroup, err := createNetworkSecurityGroup(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("network security group:", *networkSecurityGroup.ID)

	nic, err := createNIC(ctx, conn, *subnet.ID, *publicIP.ID, *networkSecurityGroup.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("network interface:", *nic.ID)

	virtualMachine, err := createVirtualMachine(ctx, conn, *nic.ID)
	if err != nil {
		log.Fatalf("cannot create virual machine:%+v", err)
	}
	log.Printf("virual machine: %s", *virtualMachine.ID)

	metricAlert, err := createMetricAlerts(ctx, conn, *virtualMachine.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("metric alert: %s", *metricAlert.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createActionGroup(ctx context.Context, conn *arm.Connection) (*armmonitor.ActionGroupResource, error) {
	actionGroupsClient := armmonitor.NewActionGroupsClient(conn, subscriptionID)

	resp, err := actionGroupsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		actionGroupName,
		armmonitor.ActionGroupResource{
			AzureResource: armmonitor.AzureResource{
				Location: to.StringPtr(location),
			},
			Properties: &armmonitor.ActionGroup{
				GroupShortName: to.StringPtr("sample"),
				Enabled:        to.BoolPtr(true),
				EmailReceivers: []*armmonitor.EmailReceiver{
					{
						Name:                 to.StringPtr("John Doe's email"),
						EmailAddress:         to.StringPtr("johndoe@eamil.com"),
						UseCommonAlertSchema: to.BoolPtr(false),
					},
				},
				SmsReceivers: []*armmonitor.SmsReceiver{
					{
						Name:        to.StringPtr("Jhon Doe's mobile"),
						CountryCode: to.StringPtr("1"),
						PhoneNumber: to.StringPtr("1234567890"),
					},
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.ActionGroupResource, nil
}

func createVirtualNetwork(ctx context.Context, conn *arm.Connection) (*armnetwork.VirtualNetwork, error) {
	virtualNetworkClient := armnetwork.NewVirtualNetworksClient(conn, subscriptionID)

	pollerResp, err := virtualNetworkClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		virtualNetworkName,
		armnetwork.VirtualNetwork{
			Resource: armnetwork.Resource{
				Location: to.StringPtr(location),
			},
			Properties: &armnetwork.VirtualNetworkPropertiesFormat{
				AddressSpace: &armnetwork.AddressSpace{
					AddressPrefixes: []*string{
						to.StringPtr("10.1.0.0/16"),
					},
				},
			},
		},
		nil)

	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.VirtualNetwork, nil
}

func createSubnet(ctx context.Context, conn *arm.Connection) (*armnetwork.Subnet, error) {
	subnetsClient := armnetwork.NewSubnetsClient(conn, subscriptionID)

	pollerResp, err := subnetsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		virtualNetworkName,
		subnetName,
		armnetwork.Subnet{
			Properties: &armnetwork.SubnetPropertiesFormat{
				AddressPrefix: to.StringPtr("10.1.0.0/24"),
			},
		},
		nil)

	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.Subnet, nil
}

func createPublicIP(ctx context.Context, conn *arm.Connection) (*armnetwork.PublicIPAddress, error) {
	publicIPClient := armnetwork.NewPublicIPAddressesClient(conn, subscriptionID)

	pollerResp, err := publicIPClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		publicIPAddressName,
		armnetwork.PublicIPAddress{
			Resource: armnetwork.Resource{
				Name:     to.StringPtr(publicIPAddressName),
				Location: to.StringPtr(location),
			},
			Properties: &armnetwork.PublicIPAddressPropertiesFormat{
				PublicIPAddressVersion:   armnetwork.IPVersionIPv4.ToPtr(),
				PublicIPAllocationMethod: armnetwork.IPAllocationMethodStatic.ToPtr(),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.PublicIPAddress, nil
}

func createNetworkSecurityGroup(ctx context.Context, conn *arm.Connection) (*armnetwork.NetworkSecurityGroup, error) {
	networkSecurityGroupClient := armnetwork.NewNetworkSecurityGroupsClient(conn, subscriptionID)

	pollerResp, err := networkSecurityGroupClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		securityGroupName,
		armnetwork.NetworkSecurityGroup{
			Resource: armnetwork.Resource{
				Location: to.StringPtr(location),
			},
			Properties: &armnetwork.NetworkSecurityGroupPropertiesFormat{
				SecurityRules: []*armnetwork.SecurityRule{
					{
						Name: to.StringPtr("allow_ssh"),
						Properties: &armnetwork.SecurityRulePropertiesFormat{
							Protocol:                 armnetwork.SecurityRuleProtocolTCP.ToPtr(),
							SourceAddressPrefix:      to.StringPtr("0.0.0.0/0"),
							SourcePortRange:          to.StringPtr("1-65535"),
							DestinationAddressPrefix: to.StringPtr("0.0.0.0/0"),
							DestinationPortRange:     to.StringPtr("22"),
							Access:                   armnetwork.SecurityRuleAccessAllow.ToPtr(),
							Direction:                armnetwork.SecurityRuleDirectionInbound.ToPtr(),
							Priority:                 to.Int32Ptr(100),
						},
					},
					{
						Name: to.StringPtr("allow_https"),
						Properties: &armnetwork.SecurityRulePropertiesFormat{
							Protocol:                 armnetwork.SecurityRuleProtocolTCP.ToPtr(),
							SourceAddressPrefix:      to.StringPtr("0.0.0.0/0"),
							SourcePortRange:          to.StringPtr("1-65535"),
							DestinationAddressPrefix: to.StringPtr("0.0.0.0/0"),
							DestinationPortRange:     to.StringPtr("443"),
							Access:                   armnetwork.SecurityRuleAccessAllow.ToPtr(),
							Direction:                armnetwork.SecurityRuleDirectionInbound.ToPtr(),
							Priority:                 to.Int32Ptr(200),
						},
					},
				},
			},
		},
		nil)

	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.NetworkSecurityGroup, nil
}

func createNIC(ctx context.Context, conn *arm.Connection, subnetID, publicIPID, networkSecurityGroupID string) (*armnetwork.NetworkInterface, error) {
	nicClient := armnetwork.NewNetworkInterfacesClient(conn, subscriptionID)

	pollerResp, err := nicClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		networkInterfaceName,
		armnetwork.NetworkInterface{
			Resource: armnetwork.Resource{
				Location: to.StringPtr(location),
			},
			Properties: &armnetwork.NetworkInterfacePropertiesFormat{
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
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.NetworkInterface, nil
}

func createVirtualMachine(ctx context.Context, connection *arm.Connection, networkInterfaceID string) (*armcompute.VirtualMachine, error) {
	vmClient := armcompute.NewVirtualMachinesClient(connection, subscriptionID)

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
				},
				OSDisk: &armcompute.OSDisk{
					Name:         to.StringPtr(osDiskName),
					CreateOption: armcompute.DiskCreateOptionTypesFromImage.ToPtr(),
					Caching:      armcompute.CachingTypesReadWrite.ToPtr(),
					ManagedDisk: &armcompute.ManagedDiskParameters{
						StorageAccountType: armcompute.StorageAccountTypesStandardLRS.ToPtr(), // OSDisk type Standard/Premium HDD/SSD
					},
				},
			},
			HardwareProfile: &armcompute.HardwareProfile{
				VMSize: armcompute.VirtualMachineSizeTypesStandardF2S.ToPtr(), // VM size include vCPUs,RAM,Data Disks,Temp storage.
			},
			OSProfile: &armcompute.OSProfile{ //
				ComputerName:  to.StringPtr("sample-compute"),
				AdminUsername: to.StringPtr("sample-user"),
				AdminPassword: to.StringPtr("Password01!@#"),
			},
			NetworkProfile: &armcompute.NetworkProfile{
				NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
					{
						SubResource: armcompute.SubResource{
							ID: to.StringPtr(networkInterfaceID),
						},
						Properties: &armcompute.NetworkInterfaceReferenceProperties{
							Primary: to.BoolPtr(true),
						},
					},
				},
			},
		},
	}

	pollerResponse, err := vmClient.BeginCreateOrUpdate(ctx, resourceGroupName, virtualMachineName, parameters, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}

	return &resp.VirtualMachine, nil
}

func createMetricAlerts(ctx context.Context, conn *arm.Connection, resourceURI string) (*armmonitor.MetricAlertResource, error) {
	metricAlertsClient := armmonitor.NewMetricAlertsClient(conn, subscriptionID)

	resp, err := metricAlertsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		metricAlertName,
		armmonitor.MetricAlertResource{
			Resource: armmonitor.Resource{
				Location: to.StringPtr(location),
			},
			Properties: &armmonitor.MetricAlertProperties{
				Description: to.StringPtr("This is the description of the rule"),
				Severity:    to.Int32Ptr(3),
				Enabled:     to.BoolPtr(true),
				Scopes: []*string{
					to.StringPtr(resourceURI),
				},
				EvaluationFrequency:  to.StringPtr("PT1M"),
				WindowSize:           to.StringPtr("PT15M"),
				TargetResourceType:   to.StringPtr("Microsoft.Compute/virtualMachines"),
				TargetResourceRegion: to.StringPtr("southcentralus"),
				Criteria: &armmonitor.MetricAlertMultipleResourceMultipleMetricCriteria{
					MetricAlertCriteria: armmonitor.MetricAlertCriteria{
						ODataType: armmonitor.OdatatypeMicrosoftAzureMonitorMultipleResourceMultipleMetricCriteria.ToPtr(),
					},
					AllOf: []armmonitor.MultiMetricCriteriaClassification{
						&armmonitor.DynamicMetricCriteria{
							MultiMetricCriteria: armmonitor.MultiMetricCriteria{
								CriterionType:   armmonitor.CriterionTypeDynamicThresholdCriterion.ToPtr(),
								Name:            to.StringPtr("High_CPU_80"),
								MetricName:      to.StringPtr("Percentage CPU"),
								MetricNamespace: to.StringPtr("Microsoft.Compute/virtualMachines"),
								TimeAggregation: armmonitor.AggregationTypeEnumAverage.ToPtr(),
								Dimensions:      []*armmonitor.MetricDimension{},
							},
							Operator: armmonitor.DynamicThresholdOperatorGreaterOrLessThan.ToPtr(),
							FailingPeriods: &armmonitor.DynamicThresholdFailingPeriods{
								NumberOfEvaluationPeriods: to.Float32Ptr(4),
								MinFailingPeriodsToAlert:  to.Float32Ptr(4),
							},
							AlertSensitivity: armmonitor.DynamicThresholdSensitivityMedium.ToPtr(),
						},
					},
				},
				AutoMitigate: to.BoolPtr(false),
				Actions:      []*armmonitor.MetricAlertAction{},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.MetricAlertResource, nil
}

func createResourceGroup(ctx context.Context, conn *arm.Connection) (*armresources.ResourceGroup, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn, subscriptionID)

	resourceGroupResp, err := resourceGroupClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		armresources.ResourceGroup{
			Location: to.StringPtr(location),
		},
		nil)
	if err != nil {
		return nil, err
	}
	return &resourceGroupResp.ResourceGroup, nil
}

func cleanup(ctx context.Context, conn *arm.Connection) (*http.Response, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn, subscriptionID)

	pollerResp, err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp.RawResponse, nil
}
