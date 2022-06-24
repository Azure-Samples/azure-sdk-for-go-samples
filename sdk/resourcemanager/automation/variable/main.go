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
	variableName          = "sample-automation-variable"
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

	variable, err := createVariable(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("automation variable:", *variable.ID)

	variable, err = getVariable(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get automation variable:", *variable.ID)

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

func createVariable(ctx context.Context, cred azcore.TokenCredential) (*armautomation.Variable, error) {
	variableClient, err := armautomation.NewVariableClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := variableClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		automationAccountName,
		variableName,
		armautomation.VariableCreateOrUpdateParameters{
			Name: to.Ptr(variableName),
			Properties: &armautomation.VariableCreateOrUpdateProperties{
				Value:       to.Ptr("\"AutomationrName.domain.com\""),
				Description: to.Ptr("Description variable"),
				IsEncrypted: to.Ptr(false),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &resp.Variable, nil
}

func getVariable(ctx context.Context, cred azcore.TokenCredential) (*armautomation.Variable, error) {
	variableClient, err := armautomation.NewVariableClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := variableClient.Get(ctx, resourceGroupName, automationAccountName, variableName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Variable, nil
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
