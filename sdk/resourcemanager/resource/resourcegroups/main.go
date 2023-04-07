// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"log"
	"os"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
)

var (
	resourcesClientFactory *armresources.ClientFactory
)

var (
	resourceGroupClient *armresources.ResourceGroupsClient
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

	exits, err := checkExistenceResourceGroup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group already exist:", exits)

	if !exits {
		resourceGroup, err := createResourceGroup(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("resources group:", *resourceGroup.ID)
	}

	resourceGroup, err := getResourceGroup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get resources group:", *resourceGroup.ID)

	resourceGroups, err := listResourceGroup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for _, resource := range resourceGroups {
		log.Printf("Resource Group Name: %s,ID: %s", *resource.Name, *resource.ID)
	}

	template, err := exportTemplateResourceGroup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("export template: %#v", template.Template)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
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

func getResourceGroup(ctx context.Context) (*armresources.ResourceGroup, error) {

	resourceGroupResp, err := resourceGroupClient.Get(ctx, resourceGroupName, nil)
	if err != nil {
		return nil, err
	}
	return &resourceGroupResp.ResourceGroup, nil
}

func listResourceGroup(ctx context.Context) ([]*armresources.ResourceGroup, error) {

	resultPager := resourceGroupClient.NewListPager(nil)

	resourceGroups := make([]*armresources.ResourceGroup, 0)
	for resultPager.More() {
		pageResp, err := resultPager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		resourceGroups = append(resourceGroups, pageResp.ResourceGroupListResult.Value...)
	}
	return resourceGroups, nil
}

func checkExistenceResourceGroup(ctx context.Context) (bool, error) {

	boolResp, err := resourceGroupClient.CheckExistence(ctx, resourceGroupName, nil)
	if err != nil {
		return false, err
	}
	return boolResp.Success, nil
}

func exportTemplateResourceGroup(ctx context.Context) (*armresources.ResourceGroupExportResult, error) {

	pollerResp, err := resourceGroupClient.BeginExportTemplate(
		ctx,
		resourceGroupName,
		armresources.ExportTemplateRequest{
			Resources: []*string{
				to.Ptr("*"),
			},
		},
		nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ResourceGroupExportResult, nil
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
