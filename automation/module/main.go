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

func createAutomationModule(ctx context.Context, cred azcore.TokenCredential) (*armautomation.Module, error) {
	moduleClient := armautomation.NewModuleClient(subscriptionID, cred, nil)
	resp, err := moduleClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		automationAccountName,
		moduleName,
		armautomation.ModuleCreateOrUpdateParameters{
			Properties: &armautomation.ModuleCreateOrUpdateProperties{
				ContentLink: &armautomation.ContentLink{
					URI: to.StringPtr("https://teststorage.blob.core.windows.net/dsccomposite/OmsCompositeResources.zip"),
					ContentHash: &armautomation.ContentHash{
						Algorithm: to.StringPtr("sha265"),
						Value:     to.StringPtr("07E108A962B81DD9C9BAA89BB47C0F6EE52B29E83758B07795E408D258B2B87A"),
					},
					Version: to.StringPtr("1.0.0.0"),
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
	moduleClient := armautomation.NewModuleClient(subscriptionID, cred, nil)
	resp, err := moduleClient.Get(ctx, resourceGroupName, automationAccountName, moduleName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Module, nil
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
