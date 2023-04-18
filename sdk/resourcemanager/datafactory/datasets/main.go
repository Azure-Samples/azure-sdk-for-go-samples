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
	dataSetName       = "sample-data-set"
)

var (
	resourcesClientFactory   *armresources.ClientFactory
	datafactoryClientFactory *armdatafactory.ClientFactory
)

var (
	resourceGroupClient *armresources.ResourceGroupsClient
	factoriesClient     *armdatafactory.FactoriesClient
	datasetsClient      *armdatafactory.DatasetsClient
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
	datasetsClient = datafactoryClientFactory.NewDatasetsClient()

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

	dataSet, err := createDataSet(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("data set:", *dataSet.ID)

	dataSet, err = getDataSet(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get data set:", *dataSet.ID)

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

func createDataSet(ctx context.Context) (*armdatafactory.DatasetResource, error) {

	resp, err := datasetsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		factoryName,
		dataSetName,
		armdatafactory.DatasetResource{
			Properties: &armdatafactory.Dataset{},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &resp.DatasetResource, nil
}

func getDataSet(ctx context.Context) (*armdatafactory.DatasetResource, error) {

	resp, err := datasetsClient.Get(ctx, resourceGroupName, factoryName, dataSetName, nil)
	if err != nil {
		return nil, err
	}

	return &resp.DatasetResource, nil
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
