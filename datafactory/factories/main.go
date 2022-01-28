package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datafactory/armdatafactory"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	factoryName       = "sample-data2factory"
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

	dataFactory, err = getDataFactory(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get data factory:", *dataFactory.ID)

	factories := getFactories(ctx, cred)
	log.Printf("list data factory:size(%d)\n", len(factories))
	for _, f := range factories {
		fmt.Printf("\t%v\n", *f.ID)
	}

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
				//RepoConfiguration: &armdatafactory.FactoryGitHubConfiguration{
				//	AccountName:         to.StringPtr("your github account"),
				//	RepositoryName:      to.StringPtr("your github repo"),
				//	CollaborationBranch: to.StringPtr("checkout branch"),
				//	RootFolder:          to.StringPtr(""),
				//},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.Factory, nil
}

func getDataFactory(ctx context.Context, cred azcore.TokenCredential) (*armdatafactory.Factory, error) {
	factoriesClient := armdatafactory.NewFactoriesClient(subscriptionID, cred, nil)
	resp, err := factoriesClient.Get(ctx, resourceGroupName, factoryName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Factory, nil
}

func getFactories(ctx context.Context, cred azcore.TokenCredential) []*armdatafactory.Factory {
	factoriesClient := armdatafactory.NewFactoriesClient(subscriptionID, cred, nil)
	list := factoriesClient.List(nil)
	var factories []*armdatafactory.Factory
	for list.NextPage(ctx) {
		resp := list.PageResponse()
		factories = append(factories, resp.Value...)
	}
	return factories
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
