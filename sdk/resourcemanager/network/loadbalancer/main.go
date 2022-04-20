// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"log"
	"os"
	"time"
)

var (
	subscriptionID      string
	location            = "westus2"
	resourceGroupName   = "sample-resources-group"
	publicIPAddressName = "sample-public-ip"
	loadBalancerName    = "sample-load-balancer"
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

	publicIP, err := createPublicIP(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("public ip:", *publicIP.ID)

	loadBalancer, err := createLoadBalancer(ctx, cred, publicIP)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("load balancer:", *loadBalancer.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
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

func createPublicIP(ctx context.Context, cred azcore.TokenCredential) (*armnetwork.PublicIPAddress, error) {
	publicIPClient, err := armnetwork.NewPublicIPAddressesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := publicIPClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		publicIPAddressName,
		armnetwork.PublicIPAddress{
			Name:     to.Ptr(publicIPAddressName),
			Location: to.Ptr(location),
			Properties: &armnetwork.PublicIPAddressPropertiesFormat{
				PublicIPAddressVersion:   to.Ptr(armnetwork.IPVersionIPv4),
				PublicIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodStatic),
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

func createLoadBalancer(ctx context.Context, cred azcore.TokenCredential, pip *armnetwork.PublicIPAddress) (*armnetwork.LoadBalancer, error) {
	probeName := "probe"
	frontEndIPConfigName := "fip"
	backEndAddressPoolName := "backEndPool"
	idPrefix := fmt.Sprintf("subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/loadBalancers", subscriptionID, resourceGroupName)

	lbClient, err := armnetwork.NewLoadBalancersClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := lbClient.BeginCreateOrUpdate(ctx,
		resourceGroupName,
		loadBalancerName,
		armnetwork.LoadBalancer{
			Location: to.Ptr(location),
			Properties: &armnetwork.LoadBalancerPropertiesFormat{
				FrontendIPConfigurations: []*armnetwork.FrontendIPConfiguration{
					{
						Name: &frontEndIPConfigName,
						Properties: &armnetwork.FrontendIPConfigurationPropertiesFormat{
							PrivateIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodDynamic),
							PublicIPAddress:           pip,
						},
					},
				},
				BackendAddressPools: []*armnetwork.BackendAddressPool{
					{
						Name: &backEndAddressPoolName,
					},
				},
				Probes: []*armnetwork.Probe{
					{
						Name: &probeName,
						Properties: &armnetwork.ProbePropertiesFormat{
							Protocol:          to.Ptr(armnetwork.ProbeProtocolHTTP),
							Port:              to.Ptr[int32](80),
							IntervalInSeconds: to.Ptr[int32](15),
							NumberOfProbes:    to.Ptr[int32](4),
							RequestPath:       to.Ptr("healthprobe.aspx"),
						},
					},
				},
				LoadBalancingRules: []*armnetwork.LoadBalancingRule{
					{
						Name: to.Ptr("lbRule"),
						Properties: &armnetwork.LoadBalancingRulePropertiesFormat{
							Protocol:             to.Ptr(armnetwork.TransportProtocolTCP),
							FrontendPort:         to.Ptr[int32](80),
							BackendPort:          to.Ptr[int32](80),
							IdleTimeoutInMinutes: to.Ptr[int32](4),
							EnableFloatingIP:     to.Ptr(false),
							LoadDistribution:     to.Ptr(armnetwork.LoadDistributionDefault),
							FrontendIPConfiguration: &armnetwork.SubResource{
								ID: to.Ptr(fmt.Sprintf("/%s/%s/frontendIPConfigurations/%s", idPrefix, loadBalancerName, frontEndIPConfigName)),
							},
							BackendAddressPool: &armnetwork.SubResource{
								ID: to.Ptr(fmt.Sprintf("/%s/%s/backendAddressPools/%s", idPrefix, loadBalancerName, backEndAddressPoolName)),
							},
							Probe: &armnetwork.SubResource{
								ID: to.Ptr(fmt.Sprintf("/%s/%s/probes/%s", idPrefix, loadBalancerName, probeName)),
							},
						},
					},
				},
				InboundNatRules: []*armnetwork.InboundNatRule{
					{
						Name: to.Ptr("natRule1"),
						Properties: &armnetwork.InboundNatRulePropertiesFormat{
							Protocol:             to.Ptr(armnetwork.TransportProtocolTCP),
							FrontendPort:         to.Ptr[int32](21),
							BackendPort:          to.Ptr[int32](22),
							EnableFloatingIP:     to.Ptr(false),
							IdleTimeoutInMinutes: to.Ptr[int32](4),
							FrontendIPConfiguration: &armnetwork.SubResource{
								ID: to.Ptr(fmt.Sprintf("/%s/%s/frontendIPConfigurations/%s", idPrefix, loadBalancerName, frontEndIPConfigName)),
							},
						},
					},
					{
						Name: to.Ptr("natRule2"),
						Properties: &armnetwork.InboundNatRulePropertiesFormat{
							Protocol:             to.Ptr(armnetwork.TransportProtocolTCP),
							FrontendPort:         to.Ptr[int32](23),
							BackendPort:          to.Ptr[int32](22),
							EnableFloatingIP:     to.Ptr(false),
							IdleTimeoutInMinutes: to.Ptr[int32](4),
							FrontendIPConfiguration: &armnetwork.SubResource{
								ID: to.Ptr(fmt.Sprintf("/%s/%s/frontendIPConfigurations/%s", idPrefix, loadBalancerName, frontEndIPConfigName)),
							},
						},
					},
				},
			},
		}, nil)

	if err != nil {
		return nil, fmt.Errorf("cannot create load balancer: %v", err)
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.LoadBalancer, nil
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

	_, err = pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return err
	}
	return nil
}
