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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/automation/armautomation"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID        string
	location              = "westus"
	resourceGroupName     = "sample-resource-group"
	automationAccountName = "sample-automation-account"
	runbookName           = "Get-AzureVMTutorial"
	webhookName           = "sample-automation-webhook"
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

	runbook, err := createAutomationRunbook(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("automation runbook:", *runbook.ID)

	webhook, err := createAutomationWebhook(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("automation webhook:", *webhook.ID)

	webhook, err = getAutomationWebhook(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get automation webhook:", *webhook.ID)

	webhookURI, err := generateURI(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("webhook uri:", webhookURI)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createAutomationAccount(ctx context.Context, cred azcore.TokenCredential) (*armautomation.Account, error) {
	accountClient := armautomation.NewAccountClient(subscriptionID, cred, nil)
	resp, err := accountClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		automationAccountName,
		armautomation.AccountCreateOrUpdateParameters{
			Location: to.StringPtr(location),
			Name:     to.StringPtr(automationAccountName),
			Properties: &armautomation.AccountCreateOrUpdateProperties{
				SKU: &armautomation.SKU{
					Name: armautomation.SKUNameEnumFree.ToPtr(),
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

func createAutomationRunbook(ctx context.Context, cred azcore.TokenCredential) (*armautomation.Runbook, error) {
	runBookClient := armautomation.NewRunbookClient(subscriptionID, cred, nil)
	resp, err := runBookClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		automationAccountName,
		runbookName,
		armautomation.RunbookCreateOrUpdateParameters{
			Location: to.StringPtr(location),
			Properties: &armautomation.RunbookCreateOrUpdateProperties{
				RunbookType: armautomation.RunbookTypeEnumPowerShellWorkflow.ToPtr(),
				LogVerbose:  to.BoolPtr(false),
				LogProgress: to.BoolPtr(true),
				PublishContentLink: &armautomation.ContentLink{
					URI: to.StringPtr("https://raw.githubusercontent.com/Azure/azure-quickstart-templates/0.0.0.3/101-automation-runbook-getvms/Runbooks/Get-AzureVMTutorial.ps1"),
					ContentHash: &armautomation.ContentHash{
						Algorithm: to.StringPtr("SHA256"),
						Value:     to.StringPtr("4fab357cab33adbe9af72ae4b1203e601e30e44de271616e376c08218fd10d1c"),
					},
				},
				Description:      to.StringPtr("Description of the Runbook"),
				LogActivityTrace: to.Int32Ptr(1),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.Runbook, nil
}

func createAutomationWebhook(ctx context.Context, cred azcore.TokenCredential) (*armautomation.Webhook, error) {
	webhookClient := armautomation.NewWebhookClient(subscriptionID, cred, nil)
	resp, err := webhookClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		automationAccountName,
		webhookName,
		armautomation.WebhookCreateOrUpdateParameters{
			Name: to.StringPtr(webhookName),
			Properties: &armautomation.WebhookCreateOrUpdateProperties{
				IsEnabled:  to.BoolPtr(true),
				URI:        to.StringPtr("https://s1events.azure-automation.net/webhooks?token=7u3KfQvM1vUPWaDMFRv2%2fAA4Jqx8QwS8aBuyO6Xsdcw%3d"),
				ExpiryTime: to.TimePtr(time.Now().AddDate(0, 0, 7)),
				Runbook: &armautomation.RunbookAssociationProperty{
					Name: to.StringPtr(runbookName),
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.Webhook, nil
}

func getAutomationWebhook(ctx context.Context, cred azcore.TokenCredential) (*armautomation.Webhook, error) {
	webhookClient := armautomation.NewWebhookClient(subscriptionID, cred, nil)
	resp, err := webhookClient.Get(ctx, resourceGroupName, automationAccountName, webhookName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Webhook, nil
}

func generateURI(ctx context.Context, cred azcore.TokenCredential) (string, error) {
	webhookClient := armautomation.NewWebhookClient(subscriptionID, cred, nil)
	resp, err := webhookClient.GenerateURI(ctx, resourceGroupName, automationAccountName, nil)
	if err != nil {
		return "", err
	}
	return *resp.Value, nil
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
