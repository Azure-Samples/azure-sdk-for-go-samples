// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/automation/armautomation"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID        string
	location              = "westus"
	resourceGroupName     = "sample-resource-group"
	automationAccountName = "sample-automation-account"
	moduleName            = "sample-automation-module"
)

var (
	resourcesClientFactory  *armresources.ClientFactory
	automationClientFactory *armautomation.ClientFactory
)

var (
	resourceGroupClient *armresources.ResourceGroupsClient
	accountClient       *armautomation.AccountClient
	moduleClient        *armautomation.ModuleClient
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

	automationClientFactory, err = armautomation.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	accountClient = automationClientFactory.NewAccountClient()
	moduleClient = automationClientFactory.NewModuleClient()

	resourceGroup, err := createResourceGroup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	account, err := createAutomationAccount(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("automation account:", *account.ID)

	module, err := createAutomationModule(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("automation module:", *module.ID)

	module, err = getAutomationModule(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get automation module:", *module.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createAutomationAccount(ctx context.Context) (*armautomation.Account, error) {

	resp, err := accountClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		automationAccountName,
		armautomation.AccountCreateOrUpdateParameters{
			Location: to.Ptr(location),
			Name:     to.Ptr(automationAccountName),
			Properties: &armautomation.AccountCreateOrUpdateProperties{
				SKU: &armautomation.SKU{
					Name: to.Ptr(armautomation.SKUNameEnumFree),
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &resp.Account, nil
}

func createAutomationModule(ctx context.Context) (*armautomation.Module, error) {

	resp, err := moduleClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		automationAccountName,
		moduleName,
		armautomation.ModuleCreateOrUpdateParameters{
			Properties: &armautomation.ModuleCreateOrUpdateProperties{
				ContentLink: &armautomation.ContentLink{
					URI: to.Ptr("https://teststorage.blob.core.windows.net/dsccomposite/OmsCompositeResources.zip"),
					ContentHash: &armautomation.ContentHash{
						Algorithm: to.Ptr("sha265"),
						Value:     to.Ptr("07E108A962B81DD9C9BAA89BB47C0F6EE52B29E83758B07795E408D258B2B87A"),
					},
					Version: to.Ptr("1.0.0.0"),
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &resp.Module, nil
}

func getAutomationModule(ctx context.Context) (*armautomation.Module, error) {

	resp, err := moduleClient.Get(ctx, resourceGroupName, automationAccountName, moduleName, nil)
	if err != nil {
		return nil, err
	}

	return &resp.Module, nil
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
