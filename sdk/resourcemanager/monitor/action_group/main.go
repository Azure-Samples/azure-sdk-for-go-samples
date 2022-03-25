// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
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
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createActionGroup(ctx context.Context, cred azcore.TokenCredential) (*armmonitor.ActionGroupResource, error) {
	actionGroupsClient := armmonitor.NewActionGroupsClient(subscriptionID, cred, nil)

	resp, err := actionGroupsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		actionGroupName,
		armmonitor.ActionGroupResource{
			Location: to.StringPtr("global"),
			Properties: &armmonitor.ActionGroup{
				GroupShortName: to.StringPtr("sample"),
				Enabled:        to.BoolPtr(true),
				EmailReceivers: []*armmonitor.EmailReceiver{
					{
						Name:                 to.StringPtr("John Doe's email"),
						EmailAddress:         to.StringPtr("johndoe@eamil.com"),
						UseCommonAlertSchema: to.BoolPtr(false),
					},
				},
				SmsReceivers: []*armmonitor.SmsReceiver{
					{
						Name:        to.StringPtr("Jhon Doe's mobile"),
						CountryCode: to.StringPtr("1"),
						PhoneNumber: to.StringPtr("1234567890"),
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
	actionGroupsClient := armmonitor.NewActionGroupsClient(subscriptionID, cred, nil)

	resp, err := actionGroupsClient.Get(ctx, resourceGroupName, actionGroupName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ActionGroupResource, nil
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
