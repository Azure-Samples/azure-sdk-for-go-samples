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
)

var (
	resourcesClientFactory  *armresources.ClientFactory
	automationClientFactory *armautomation.ClientFactory
)

var (
	resourceGroupClient *armresources.ResourceGroupsClient
	accountClient       *armautomation.AccountClient
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

	account, err = updateAutomationAccount(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("update automation account:", *account.ID, *account.Tags["automation"])

	account, err = getAutomationAccount(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get automation account:", *account.ID)

	accounts, err := listAutomationAccount(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("list automation account:")
	for _, tmp := range accounts {
		log.Printf("\t%v", *tmp.ID)
	}

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

func updateAutomationAccount(ctx context.Context) (*armautomation.Account, error) {

	resp, err := accountClient.Update(
		ctx,
		resourceGroupName,
		automationAccountName,
		armautomation.AccountUpdateParameters{
			Tags: map[string]*string{
				"automation": to.Ptr("sample"),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &resp.Account, nil
}

func getAutomationAccount(ctx context.Context) (*armautomation.Account, error) {

	resp, err := accountClient.Get(ctx, resourceGroupName, automationAccountName, nil)
	if err != nil {
		return nil, err
	}

	return &resp.Account, nil
}

func listAutomationAccount(ctx context.Context) ([]*armautomation.Account, error) {

	list := accountClient.NewListPager(nil)
	accounts := make([]*armautomation.Account, 0)
	for list.More() {
		resp, err := list.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, resp.Value...)
	}

	return accounts, nil
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
