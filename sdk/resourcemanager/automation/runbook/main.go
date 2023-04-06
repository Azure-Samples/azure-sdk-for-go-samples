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
	runbookName           = "sample-automation-runbook"
)

var (
	resourcesClientFactory  *armresources.ClientFactory
	automationClientFactory *armautomation.ClientFactory
)

var (
	resourceGroupClient *armresources.ResourceGroupsClient
	accountClient       *armautomation.AccountClient
	runbookClient       *armautomation.RunbookClient
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
	runbookClient = automationClientFactory.NewRunbookClient()

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

	runbook, err := createAutomationRunbook(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("automation runbook:", *runbook.ID)

	runbook, err = getAutomationRunbook(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get automation runbook:", *runbook.ID)

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

func createAutomationRunbook(ctx context.Context) (*armautomation.Runbook, error) {

	resp, err := runbookClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		automationAccountName,
		runbookName,
		armautomation.RunbookCreateOrUpdateParameters{
			Location: to.Ptr(location),
			Properties: &armautomation.RunbookCreateOrUpdateProperties{
				RunbookType:      to.Ptr(armautomation.RunbookTypeEnumPowerShellWorkflow),
				LogVerbose:       to.Ptr(false),
				LogProgress:      to.Ptr(false),
				Description:      to.Ptr("Description of the Runbook"),
				LogActivityTrace: to.Ptr[int32](1),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &resp.Runbook, nil
}

func getAutomationRunbook(ctx context.Context) (*armautomation.Runbook, error) {

	resp, err := runbookClient.Get(ctx, resourceGroupName, automationAccountName, runbookName, nil)
	if err != nil {
		return nil, err
	}

	return &resp.Runbook, nil
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
