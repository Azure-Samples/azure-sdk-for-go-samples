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
	credentialName        = "sample-automation-credential"
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

	credential, err := createAutomationCredential(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("automation credential:", *credential.ID)

	credential, err = updateAutomationCredential(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("update automation credential:", *credential.ID, *credential.Properties.Description)

	credential, err = getAutomationCredential(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get automation credential:", *credential.ID)

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

func createAutomationCredential(ctx context.Context, cred azcore.TokenCredential) (*armautomation.Credential, error) {
	credentialClient := armautomation.NewCredentialClient(subscriptionID, cred, nil)
	resp, err := credentialClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		automationAccountName,
		credentialName,
		armautomation.CredentialCreateOrUpdateParameters{
			Name: to.StringPtr(credentialName),
			Properties: &armautomation.CredentialCreateOrUpdateProperties{
				UserName:    to.StringPtr("azuresdkforgo"),
				Password:    to.StringPtr("QWE!@#123"),
				Description: to.StringPtr("description goes here"),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.Credential, nil
}

func updateAutomationCredential(ctx context.Context, cred azcore.TokenCredential) (*armautomation.Credential, error) {
	credentialClient := armautomation.NewCredentialClient(subscriptionID, cred, nil)
	resp, err := credentialClient.Update(
		ctx,
		resourceGroupName,
		automationAccountName,
		credentialName,
		armautomation.CredentialUpdateParameters{
			Properties: &armautomation.CredentialUpdateProperties{
				UserName:    to.StringPtr("azuresdkforgo"),
				Description: to.StringPtr("updated description"),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.Credential, nil
}

func getAutomationCredential(ctx context.Context, cred azcore.TokenCredential) (*armautomation.Credential, error) {
	credentialClient := armautomation.NewCredentialClient(subscriptionID, cred, nil)
	resp, err := credentialClient.Get(ctx, resourceGroupName, automationAccountName, credentialName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Credential, nil
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
