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
	location             = "eastus"
	resourceGroupName    = "sample-resource-group"
	virtualNetworkName   = "sample-virtual-network"
	subnetName           = "sample-subnet"
	networkInterfaceName = "sample-network-interface"
	metricAlertName      = "sample-metric-alert"
	vmScaleSetName       = "sample-vm-scale-set"
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

	vmss, err := createVMSS(ctx, cred, *subnet.ID)
	if err != nil {
		log.Fatalf("cannot create virtual machine scale sets:%+v", err)
	}
	log.Printf("virtual machine scale sets: %s", *vmss.ID)

	autoscaleSetting, err := createAutoscaleSetting(ctx, cred, *vmss.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("autoscale Setting: %s", *autoscaleSetting.ID)

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

func createVMSS(ctx context.Context, cred azcore.TokenCredential, subnetID string) (*armcompute.VirtualMachineScaleSet, error) {
	vmssClient := armcompute.NewVirtualMachineScaleSetsClient(subscriptionID, cred, nil)

	pollerResp, err := vmssClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		vmScaleSetName,
		armcompute.VirtualMachineScaleSet{
			Location: to.StringPtr(location),
			SKU: &armcompute.SKU{
				Name:     to.StringPtr("Standard_D1_v2"),
				Capacity: to.Int64Ptr(2),
				Tier:     to.StringPtr("Standard"),
			},
			Properties: &armcompute.VirtualMachineScaleSetProperties{
				Overprovision: to.BoolPtr(true),
				UpgradePolicy: &armcompute.UpgradePolicy{
					Mode: armcompute.UpgradeModeManual.ToPtr(),
				},
				VirtualMachineProfile: &armcompute.VirtualMachineScaleSetVMProfile{
					OSProfile: &armcompute.VirtualMachineScaleSetOSProfile{
						ComputerNamePrefix: to.StringPtr("vmss"),
						AdminUsername:      to.StringPtr("sample-user"),
						AdminPassword:      to.StringPtr("Password01!@#"),
					},
					StorageProfile: &armcompute.VirtualMachineScaleSetStorageProfile{
						ImageReference: &armcompute.ImageReference{
							Offer:     to.StringPtr("WindowsServer"),
							Publisher: to.StringPtr("MicrosoftWindowsServer"),
							SKU:       to.StringPtr("2019-Datacenter"),
							Version:   to.StringPtr("latest"),
						},
						OSDisk: &armcompute.VirtualMachineScaleSetOSDisk{
							Caching: armcompute.CachingTypesReadWrite.ToPtr(),
							ManagedDisk: &armcompute.VirtualMachineScaleSetManagedDiskParameters{
								StorageAccountType: armcompute.StorageAccountTypesStandardLRS.ToPtr(),
							},
							CreateOption: armcompute.DiskCreateOptionTypesFromImage.ToPtr(),
							DiskSizeGB:   to.Int32Ptr(128),
						},
					},
					NetworkProfile: &armcompute.VirtualMachineScaleSetNetworkProfile{
						NetworkInterfaceConfigurations: []*armcompute.VirtualMachineScaleSetNetworkConfiguration{
							{
								Name: to.StringPtr(vmScaleSetName),
								Properties: &armcompute.VirtualMachineScaleSetNetworkConfigurationProperties{
									Primary:            to.BoolPtr(true),
									EnableIPForwarding: to.BoolPtr(true),
									IPConfigurations: []*armcompute.VirtualMachineScaleSetIPConfiguration{
										{
											Name: to.StringPtr(vmScaleSetName),
											Properties: &armcompute.VirtualMachineScaleSetIPConfigurationProperties{
												Subnet: &armcompute.APIEntityReference{
													ID: to.StringPtr(subnetID),
												},
											},
										},
									},
								},
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
	return &resp.VirtualMachineScaleSet, nil
}

func createAutoscaleSetting(ctx context.Context, cred azcore.TokenCredential, resourceURI string) (*armmonitor.AutoscaleSettingResource, error) {
	autoscaleSettingsClient := armmonitor.NewAutoscaleSettingsClient(subscriptionID, cred, nil)

	resp, err := autoscaleSettingsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		metricAlertName,
		armmonitor.AutoscaleSettingResource{
			Location: to.StringPtr(location),
			Properties: &armmonitor.AutoscaleSetting{
				Profiles: []*armmonitor.AutoscaleProfile{
					{
						Name: to.StringPtr("adios"),
						Capacity: &armmonitor.ScaleCapacity{
							Default: to.StringPtr("1"),
							Maximum: to.StringPtr("10"),
							Minimum: to.StringPtr("1"),
						},
						Rules: []*armmonitor.ScaleRule{},
					},
				},
				Enabled:           to.BoolPtr(true),
				TargetResourceURI: to.StringPtr(resourceURI),
				Notifications: []*armmonitor.AutoscaleNotification{
					{
						Operation: to.StringPtr("Scale"),
						Email: &armmonitor.EmailNotification{
							CustomEmails: []*string{
								to.StringPtr("ga@ms.com"),
								to.StringPtr("hu@ws.com"),
							},
							SendToSubscriptionAdministrator:    to.BoolPtr(true),
							SendToSubscriptionCoAdministrators: to.BoolPtr(true),
						},
						Webhooks: []*armmonitor.WebhookNotification{},
					},
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.AutoscaleSettingResource, nil
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
