// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datafactory/armdatafactory/v3"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"log"
	"os"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	factoryName       = "sample-data2factory"
	dataflowName      = "sample-dataflow"
)

var (
	resourcesClientFactory   *armresources.ClientFactory
	datafactoryClientFactory *armdatafactory.ClientFactory
)

var (
	resourceGroupClient *armresources.ResourceGroupsClient
	factoriesClient     *armdatafactory.FactoriesClient
	dataFlowsClient     *armdatafactory.DataFlowsClient
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

	resourcesClientFactory, err = armresources.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	resourceGroupClient = resourcesClientFactory.NewResourceGroupsClient()

	datafactoryClientFactory, err = armdatafactory.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	factoriesClient = datafactoryClientFactory.NewFactoriesClient()
	dataFlowsClient = datafactoryClientFactory.NewDataFlowsClient()

	resourceGroup, err := createResourceGroup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	dataFactory, err := createDataFactory(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("data factory:", *dataFactory.ID)

	dataFlow, err := createDataFlow(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("data flow:", *dataFlow.ID)

	dataFlow, err = getDataFlow(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get data flow:", *dataFlow.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createDataFactory(ctx context.Context) (*armdatafactory.Factory, error) {

	resp, err := factoriesClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		factoryName,
		armdatafactory.Factory{
			Location: to.Ptr(location),
			Properties: &armdatafactory.FactoryProperties{
				PublicNetworkAccess: to.Ptr(armdatafactory.PublicNetworkAccessEnabled),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &resp.Factory, nil
}

func createDataFlow(ctx context.Context) (*armdatafactory.DataFlowResource, error) {

	resp, err := dataFlowsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		factoryName,
		dataflowName,
		armdatafactory.DataFlowResource{
			Properties: &armdatafactory.DataFlow{
				//Type: to.StringPtr(""),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &resp.DataFlowResource, nil
}

func getDataFlow(ctx context.Context) (*armdatafactory.DataFlowResource, error) {

	resp, err := dataFlowsClient.Get(ctx, resourceGroupName, factoryName, dataflowName, nil)
	if err != nil {
		return nil, err
	}

	return &resp.DataFlowResource, nil
}

func createResourceGroup(ctx context.Context) (*armresources.ResourceGroup, error) {

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

func cleanup(ctx context.Context) error {

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
