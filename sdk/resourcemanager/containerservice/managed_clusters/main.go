// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID      string
	objectID            string
	clientSecret        string
	location            = "westus"
	resourceGroupName   = "sample-resource-group"
	managedClustersName = "sample-aks-cluster"
)

func main() {
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set.")
	}

	objectID = os.Getenv("AZURE_OBJECT_ID")
	if len(objectID) == 0 {
		log.Fatal("AZURE_OBJECT_ID is not set.")
	}

	clientSecret = os.Getenv("AZURE_CLIENT_SECRET")
	if len(clientSecret) == 0 {
		log.Fatal("AZURE_CLIENT_SECRET is not set.")
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

	managedCluster, err := createManagedCluster(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("managed cluster:", *managedCluster.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createManagedCluster(ctx context.Context, cred azcore.TokenCredential) (*armcontainerservice.ManagedCluster, error) {
	managedClustersClient, err := armcontainerservice.NewManagedClustersClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := managedClustersClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		managedClustersName,
		armcontainerservice.ManagedCluster{
			Location: to.Ptr(location),
			Properties: &armcontainerservice.ManagedClusterProperties{
				DNSPrefix: to.Ptr("aksgosdk"),
				AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
					{
						Name:              to.Ptr("askagent"),
						Count:             to.Ptr[int32](1),
						VMSize:            to.Ptr("Standard_DS2_v2"),
						MaxPods:           to.Ptr[int32](110),
						MinCount:          to.Ptr[int32](1),
						MaxCount:          to.Ptr[int32](100),
						OSType:            to.Ptr(armcontainerservice.OSTypeLinux),
						Type:              to.Ptr(armcontainerservice.AgentPoolTypeVirtualMachineScaleSets),
						EnableAutoScaling: to.Ptr(true),
						Mode:              to.Ptr(armcontainerservice.AgentPoolModeSystem),
					},
				},
				ServicePrincipalProfile: &armcontainerservice.ManagedClusterServicePrincipalProfile{
					ClientID: to.Ptr(objectID),
					Secret:   to.Ptr(clientSecret),
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
	return &resp.ManagedCluster, nil
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

	_, err = pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return err
	}
	return nil
}
