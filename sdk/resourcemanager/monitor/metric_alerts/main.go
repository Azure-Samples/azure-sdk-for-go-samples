// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID       string
	location             = "southcentralus"
	resourceGroupName    = "sample-resource-group"
	virtualNetworkName   = "sample-virtual-network"
	subnetName           = "sample-subnet"
	networkInterfaceName = "sample-network-interface"
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
	ctx := context.Background()

	resourceGroup, err := createResourceGroup(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	virtualNetwork, err := createVirtualNetwork(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("virtual network:", *virtualNetwork.ID)

	subnet, err := createSubnet(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("subnet:", *subnet.ID)

	nic, err := createNIC(ctx, cred, *subnet.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("network interface:", *nic.ID)

	virtualMachine, err := createVirtualMachine(ctx, cred, *nic.ID)
	if err != nil {
		log.Fatalf("cannot create virual machine:%+v", err)
	}
	log.Printf("virual machine: %s", *virtualMachine.ID)

	metricAlert, err := createMetricAlerts(ctx, cred, *virtualMachine.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("metric alert: %s", *metricAlert.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createVirtualNetwork(ctx context.Context, cred azcore.TokenCredential) (*armnetwork.VirtualNetwork, error) {
	virtualNetworkClient := armnetwork.NewVirtualNetworksClient(subscriptionID, cred, nil)

	pollerResp, err := virtualNetworkClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		virtualNetworkName,
		armnetwork.VirtualNetwork{
			Location: to.StringPtr(location),
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

func createSubnet(ctx context.Context, cred azcore.TokenCredential) (*armnetwork.Subnet, error) {
	subnetsClient := armnetwork.NewSubnetsClient(subscriptionID, cred, nil)

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

func createNIC(ctx context.Context, cred azcore.TokenCredential, subnetID string) (*armnetwork.Interface, error) {
	nicClient := armnetwork.NewInterfacesClient(subscriptionID, cred, nil)

	pollerResp, err := nicClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		networkInterfaceName,
		armnetwork.Interface{
			Location: to.StringPtr(location),
			Properties: &armnetwork.InterfacePropertiesFormat{
				IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
					{
						Name: to.StringPtr("ipConfig"),
						Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
							PrivateIPAllocationMethod: armnetwork.IPAllocationMethodDynamic.ToPtr(),
							Subnet: &armnetwork.Subnet{
								ID: to.StringPtr(subnetID),
							},
						},
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
	return &resp.Interface, nil
}

func createVirtualMachine(ctx context.Context, cred azcore.TokenCredential, networkInterfaceID string) (*armcompute.VirtualMachine, error) {
	vmClient := armcompute.NewVirtualMachinesClient(subscriptionID, cred, nil)

	parameters := armcompute.VirtualMachine{
		Location: to.StringPtr(location),
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
				VMSize: armcompute.VirtualMachineSizeTypes("Standard_F2s").ToPtr(), // VM size include vCPUs,RAM,Data Disks,Temp storage.
			},
			OSProfile: &armcompute.OSProfile{ //
				ComputerName:  to.StringPtr("sample-compute"),
				AdminUsername: to.StringPtr("sample-user"),
				AdminPassword: to.StringPtr("Password01!@#"),
			},
			NetworkProfile: &armcompute.NetworkProfile{
				NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
					{
						ID: to.StringPtr(networkInterfaceID),
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

func createMetricAlerts(ctx context.Context, cred azcore.TokenCredential, resourceURI string) (*armmonitor.MetricAlertResource, error) {
	metricAlertsClient := armmonitor.NewMetricAlertsClient(subscriptionID, cred, nil)

	resp, err := metricAlertsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		metricAlertName,
		armmonitor.MetricAlertResource{
			Location: to.StringPtr("global"),
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
					ODataType: armmonitor.OdatatypeMicrosoftAzureMonitorMultipleResourceMultipleMetricCriteria.ToPtr(),
					AllOf: []armmonitor.MultiMetricCriteriaClassification{
						&armmonitor.DynamicMetricCriteria{
							CriterionType:   armmonitor.CriterionTypeDynamicThresholdCriterion.ToPtr(),
							Name:            to.StringPtr("High_CPU_80"),
							MetricName:      to.StringPtr("Percentage CPU"),
							MetricNamespace: to.StringPtr("Microsoft.Compute/virtualMachines"),
							TimeAggregation: armmonitor.AggregationTypeEnumAverage.ToPtr(),
							Dimensions:      []*armmonitor.MetricDimension{},
							Operator:        armmonitor.DynamicThresholdOperatorGreaterOrLessThan.ToPtr(),
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

func createResourceGroup(ctx context.Context, cred azcore.TokenCredential) (*armresources.ResourceGroup, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)

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

func cleanup(ctx context.Context, cred azcore.TokenCredential) (*http.Response, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)

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
