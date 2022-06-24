package main

import (
	"context"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/logic/armlogic"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	workflowName      = "sample-logic-workflow"
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

	workflow, err := createWorkflow(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("logic workflows:", *workflow.ID)

	workflow, err = getWorkflow(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get logic workflows:", *workflow.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createWorkflow(ctx context.Context, cred azcore.TokenCredential) (*armlogic.Workflow, error) {
	workflowsClient, err := armlogic.NewWorkflowsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := workflowsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		workflowName,
		armlogic.Workflow{
			Location: to.Ptr(location),
			Properties: &armlogic.WorkflowProperties{
				Definition: map[string]interface{}{
					"$schema":        "https://schema.management.azure.com/providers/Microsoft.Logic/schemas/2016-06-01/workflowdefinition.json#",
					"contentVersion": "1.0.0.0",
					//"parameters":     {},
					//"triggers":       {},
					//"actions":        {},
					//"outputs":        {},
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &resp.Workflow, nil
}

func getWorkflow(ctx context.Context, cred azcore.TokenCredential) (*armlogic.Workflow, error) {
	workflowsClient, err := armlogic.NewWorkflowsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := workflowsClient.Get(ctx, resourceGroupName, workflowName, nil)
	if err != nil {
		return nil, err
	}

	return &resp.Workflow, nil
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
