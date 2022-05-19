// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"log"
	"os"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	actionGroupName   = "sample-action-group"
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

	actionGroup, err := createActionGroup(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("action group:", *actionGroup.ID)

	actionGroup, err = getActionGroup(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get action group:", *actionGroup.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createActionGroup(ctx context.Context, cred azcore.TokenCredential) (*armmonitor.ActionGroupResource, error) {
	actionGroupsClient, err := armmonitor.NewActionGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := actionGroupsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		actionGroupName,
		armmonitor.ActionGroupResource{
			Location: to.Ptr("global"),
			Properties: &armmonitor.ActionGroup{
				GroupShortName: to.Ptr("sample"),
				Enabled:        to.Ptr(true),
				EmailReceivers: []*armmonitor.EmailReceiver{
					{
						Name:                 to.Ptr("John Doe's email"),
						EmailAddress:         to.Ptr("johndoe@eamil.com"),
						UseCommonAlertSchema: to.Ptr(false),
					},
				},
				SmsReceivers: []*armmonitor.SmsReceiver{
					{
						Name:        to.Ptr("Jhon Doe's mobile"),
						CountryCode: to.Ptr("1"),
						PhoneNumber: to.Ptr("1234567890"),
					},
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.ActionGroupResource, nil
}

func getActionGroup(ctx context.Context, cred azcore.TokenCredential) (*armmonitor.ActionGroupResource, error) {
	actionGroupsClient, err := armmonitor.NewActionGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := actionGroupsClient.Get(ctx, resourceGroupName, actionGroupName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ActionGroupResource, nil
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
