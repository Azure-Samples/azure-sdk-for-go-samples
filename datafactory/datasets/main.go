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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datafactory/armdatafactory"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	factoryName       = "sample-data2factory"
	dataSetName       = "sample-data-set"
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

	dataFactory, err := createDataFactory(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("data factory:", *dataFactory.ID)

	dataSet, err := createDataSet(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("data set:", *dataSet.ID)

	dataSet, err = getDataSet(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get data set:", *dataSet.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createDataFactory(ctx context.Context, cred azcore.TokenCredential) (*armdatafactory.Factory, error) {
	factoriesClient := armdatafactory.NewFactoriesClient(subscriptionID, cred, nil)
	resp, err := factoriesClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		factoryName,
		armdatafactory.Factory{
			Location: to.StringPtr(location),
			Properties: &armdatafactory.FactoryProperties{
				PublicNetworkAccess: armdatafactory.PublicNetworkAccessEnabled.ToPtr(),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.Factory, nil
}

func createDataSet(ctx context.Context, cred azcore.TokenCredential) (*armdatafactory.DatasetResource, error) {
	dataSetsClient := armdatafactory.NewDatasetsClient(subscriptionID, cred, nil)
	resp, err := dataSetsClient.CreateOrUpdate(
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

func getDataSet(ctx context.Context, cred azcore.TokenCredential) (*armdatafactory.DatasetResource, error) {
	dataSetsClient := armdatafactory.NewDatasetsClient(subscriptionID, cred, nil)
	resp, err := dataSetsClient.Get(ctx, resourceGroupName, factoryName, dataSetName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.DatasetResource, nil
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
