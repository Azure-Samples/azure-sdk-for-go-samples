// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
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

	account, err := createAutomationAccount(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("automation account:", *account.ID)

	module, err := createAutomationModule(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("automation module:", *module.ID)

	module, err = getAutomationModule(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get automation module:", *module.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createAutomationAccount(ctx context.Context, cred azcore.TokenCredential) (*armautomation.Account, error) {
	accountClient, err := armautomation.NewAccountClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

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

func createAutomationModule(ctx context.Context, cred azcore.TokenCredential) (*armautomation.Module, error) {
	moduleClient, err := armautomation.NewModuleClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

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

func getAutomationModule(ctx context.Context, cred azcore.TokenCredential) (*armautomation.Module, error) {
	moduleClient, err := armautomation.NewModuleClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := moduleClient.Get(ctx, resourceGroupName, automationAccountName, moduleName, nil)
	if err != nil {
		return nil, err
	}

	return &resp.Module, nil
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
