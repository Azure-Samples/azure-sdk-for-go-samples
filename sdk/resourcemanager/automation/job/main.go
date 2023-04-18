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
	runbookName           = "Get-AzureVMTutorial"
	jobName               = "sample-automation-job"
)

var (
	resourcesClientFactory  *armresources.ClientFactory
	automationClientFactory *armautomation.ClientFactory
)

var (
	resourceGroupClient *armresources.ResourceGroupsClient
	accountClient       *armautomation.AccountClient
	runbookClient       *armautomation.RunbookClient
	jobClient           *armautomation.JobClient
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
	jobClient = automationClientFactory.NewJobClient()

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

	job, err := createAutomationJob(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("automation job:", *job.ID)

	job, err = getAutomationJob(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get automation job:", *job.ID)

	jobOutput, err := getJobOutput(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get automation job output:", jobOutput)

	runbookContent, err := getRunbookContent(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get automation job runbook content:", runbookContent)

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
				RunbookType: to.Ptr(armautomation.RunbookTypeEnumPowerShellWorkflow),
				LogVerbose:  to.Ptr(false),
				LogProgress: to.Ptr(true),
				PublishContentLink: &armautomation.ContentLink{
					URI: to.Ptr("https://raw.githubusercontent.com/Azure/azure-quickstart-templates/0.0.0.3/101-automation-runbook-getvms/Runbooks/Get-AzureVMTutorial.ps1"),
					ContentHash: &armautomation.ContentHash{
						Algorithm: to.Ptr("SHA256"),
						Value:     to.Ptr("4fab357cab33adbe9af72ae4b1203e601e30e44de271616e376c08218fd10d1c"),
					},
				},
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

func createAutomationJob(ctx context.Context) (*armautomation.Job, error) {

	resp, err := jobClient.Create(
		ctx,
		resourceGroupName,
		automationAccountName,
		jobName,
		armautomation.JobCreateParameters{
			Properties: &armautomation.JobCreateProperties{
				Runbook: &armautomation.RunbookAssociationProperty{
					Name: to.Ptr(runbookName),
				},
				Parameters: map[string]*string{
					"key1": to.Ptr("value1"),
					"key2": to.Ptr("value2"),
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &resp.Job, nil
}

func getAutomationJob(ctx context.Context) (*armautomation.Job, error) {

	resp, err := jobClient.Get(ctx, resourceGroupName, automationAccountName, jobName, nil)
	if err != nil {
		return nil, err
	}

	return &resp.Job, nil
}

func getJobOutput(ctx context.Context) (string, error) {

	resp, err := jobClient.GetOutput(ctx, resourceGroupName, automationAccountName, jobName, nil)
	if err != nil {
		return "", err
	}

	return *resp.Value, nil
}

func getRunbookContent(ctx context.Context) (string, error) {

	resp, err := jobClient.GetRunbookContent(ctx, resourceGroupName, automationAccountName, jobName, nil)
	if err != nil {
		return "", err
	}

	return *resp.Value, nil
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
