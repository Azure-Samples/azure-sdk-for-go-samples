// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"log"
	"os"
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
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createVirtualNetwork(ctx context.Context, cred azcore.TokenCredential) (*armnetwork.VirtualNetwork, error) {
	virtualNetworkClient, err := armnetwork.NewVirtualNetworksClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := virtualNetworkClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		virtualNetworkName,
		armnetwork.VirtualNetwork{
			Location: to.Ptr(location),
			Properties: &armnetwork.VirtualNetworkPropertiesFormat{
				AddressSpace: &armnetwork.AddressSpace{
					AddressPrefixes: []*string{
						to.Ptr("10.1.0.0/16"),
					},
				},
			},
		},
		nil)

	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.VirtualNetwork, nil
}

func createSubnet(ctx context.Context, cred azcore.TokenCredential) (*armnetwork.Subnet, error) {
	subnetsClient, err := armnetwork.NewSubnetsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := subnetsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		virtualNetworkName,
		subnetName,
		armnetwork.Subnet{
			Properties: &armnetwork.SubnetPropertiesFormat{
				AddressPrefix: to.Ptr("10.1.0.0/24"),
			},
		},
		nil)

	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Subnet, nil
}

func createNIC(ctx context.Context, cred azcore.TokenCredential, subnetID string) (*armnetwork.Interface, error) {
	nicClient, err := armnetwork.NewInterfacesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := nicClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		networkInterfaceName,
		armnetwork.Interface{
			Location: to.Ptr(location),
			Properties: &armnetwork.InterfacePropertiesFormat{
				IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
					{
						Name: to.Ptr("ipConfig"),
						Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
							Subnet: &armnetwork.Subnet{
								ID: to.Ptr(subnetID),
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

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Interface, nil
}

func createVMSS(ctx context.Context, cred azcore.TokenCredential, subnetID string) (*armcompute.VirtualMachineScaleSet, error) {
	vmssClient, err := armcompute.NewVirtualMachineScaleSetsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := vmssClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		vmScaleSetName,
		armcompute.VirtualMachineScaleSet{
			Location: to.Ptr(location),
			SKU: &armcompute.SKU{
				Name:     to.Ptr("Standard_D1_v2"),
				Capacity: to.Ptr[int64](2),
				Tier:     to.Ptr("Standard"),
			},
			Properties: &armcompute.VirtualMachineScaleSetProperties{
				Overprovision: to.Ptr(true),
				UpgradePolicy: &armcompute.UpgradePolicy{
					Mode: to.Ptr(armcompute.UpgradeModeManual),
				},
				VirtualMachineProfile: &armcompute.VirtualMachineScaleSetVMProfile{
					OSProfile: &armcompute.VirtualMachineScaleSetOSProfile{
						ComputerNamePrefix: to.Ptr("vmss"),
						AdminUsername:      to.Ptr("sample-user"),
						AdminPassword:      to.Ptr("Password01!@#"),
					},
					StorageProfile: &armcompute.VirtualMachineScaleSetStorageProfile{
						ImageReference: &armcompute.ImageReference{
							Offer:     to.Ptr("WindowsServer"),
							Publisher: to.Ptr("MicrosoftWindowsServer"),
							SKU:       to.Ptr("2019-Datacenter"),
							Version:   to.Ptr("latest"),
						},
						OSDisk: &armcompute.VirtualMachineScaleSetOSDisk{
							Caching: to.Ptr(armcompute.CachingTypesReadWrite),
							ManagedDisk: &armcompute.VirtualMachineScaleSetManagedDiskParameters{
								StorageAccountType: to.Ptr(armcompute.StorageAccountTypesStandardLRS),
							},
							CreateOption: to.Ptr(armcompute.DiskCreateOptionTypesFromImage),
							DiskSizeGB:   to.Ptr[int32](128),
						},
					},
					NetworkProfile: &armcompute.VirtualMachineScaleSetNetworkProfile{
						NetworkInterfaceConfigurations: []*armcompute.VirtualMachineScaleSetNetworkConfiguration{
							{
								Name: to.Ptr(vmScaleSetName),
								Properties: &armcompute.VirtualMachineScaleSetNetworkConfigurationProperties{
									Primary:            to.Ptr(true),
									EnableIPForwarding: to.Ptr(true),
									IPConfigurations: []*armcompute.VirtualMachineScaleSetIPConfiguration{
										{
											Name: to.Ptr(vmScaleSetName),
											Properties: &armcompute.VirtualMachineScaleSetIPConfigurationProperties{
												Subnet: &armcompute.APIEntityReference{
													ID: to.Ptr(subnetID),
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

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.VirtualMachineScaleSet, nil
}

func createAutoscaleSetting(ctx context.Context, cred azcore.TokenCredential, resourceURI string) (*armmonitor.AutoscaleSettingResource, error) {
	autoscaleSettingsClient, err := armmonitor.NewAutoscaleSettingsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := autoscaleSettingsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		metricAlertName,
		armmonitor.AutoscaleSettingResource{
			Location: to.Ptr(location),
			Properties: &armmonitor.AutoscaleSetting{
				Profiles: []*armmonitor.AutoscaleProfile{
					{
						Name: to.Ptr("adios"),
						Capacity: &armmonitor.ScaleCapacity{
							Default: to.Ptr("1"),
							Maximum: to.Ptr("10"),
							Minimum: to.Ptr("1"),
						},
						Rules: []*armmonitor.ScaleRule{},
					},
				},
				Enabled:           to.Ptr(true),
				TargetResourceURI: to.Ptr(resourceURI),
				Notifications: []*armmonitor.AutoscaleNotification{
					{
						Operation: to.Ptr("Scale"),
						Email: &armmonitor.EmailNotification{
							CustomEmails: []*string{
								to.Ptr("ga@ms.com"),
								to.Ptr("hu@ws.com"),
							},
							SendToSubscriptionAdministrator:    to.Ptr(true),
							SendToSubscriptionCoAdministrators: to.Ptr(true),
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
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resourceGroupResp, err := resourceGroupClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		armresources.ResourceGroup{
			Location: to.Ptr(location),
		},
		nil)
	if err != nil {
		return nil, err
	}
	return &resourceGroupResp.ResourceGroup, nil
}

func cleanup(ctx context.Context, cred azcore.TokenCredential) error {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	pollerResp, err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return err
	}

	_, err = pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}
	return nil
}
