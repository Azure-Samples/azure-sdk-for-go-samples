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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
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

	exits, err := checkExistenceResourceGroup(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group already exist:", exits)

	if !exits {
		resourceGroup, err := createResourceGroup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("resources group:", *resourceGroup.ID)
	}

	resourceGroup, err := getResourceGroup(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get resources group:", *resourceGroup.ID)

	resourceGroups := listResourceGroup(ctx, cred)
	for _, resource := range resourceGroups {
		log.Printf("Resource Group Name: %s,ID: %s", *resource.Name, *resource.ID)
	}

	template, err := exportTemplateResourceGroup(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("export template: %#v", template.Template)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
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

func getResourceGroup(ctx context.Context, cred azcore.TokenCredential) (*armresources.ResourceGroup, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)

	resourceGroupResp, err := resourceGroupClient.Get(ctx, resourceGroupName, nil)
	if err != nil {
		return nil, err
	}
	return &resourceGroupResp.ResourceGroup, nil
}

func listResourceGroup(ctx context.Context, cred azcore.TokenCredential) []*armresources.ResourceGroup {
	resourceGroupClient := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)

	resultPager := resourceGroupClient.List(nil)

	resourceGroups := make([]*armresources.ResourceGroup, 0)
	for resultPager.NextPage(ctx) {
		pageResp := resultPager.PageResponse()
		resourceGroups = append(resourceGroups, pageResp.ResourceGroupListResult.Value...)
	}
	return resourceGroups
}

func checkExistenceResourceGroup(ctx context.Context, cred azcore.TokenCredential) (bool, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)

	boolResp, err := resourceGroupClient.CheckExistence(ctx, resourceGroupName, nil)
	if err != nil {
		return false, err
	}
	return boolResp.Success, nil
}

func exportTemplateResourceGroup(ctx context.Context, cred azcore.TokenCredential) (*armresources.ResourceGroupExportResult, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)

	pollerResp, err := resourceGroupClient.BeginExportTemplate(
		ctx,
		resourceGroupName,
		armresources.ExportTemplateRequest{
			Resources: []*string{
				to.StringPtr("*"),
			},
		},
		nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.ResourceGroupExportResult, nil
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
