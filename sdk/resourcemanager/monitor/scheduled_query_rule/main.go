// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/operationalinsights/armoperationalinsights"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	workspaceName     = "sample-workspace"
	ruleName          = "sample-scheduled-query-rules"
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

	workspace, err := createWorkspaces(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("workspace:", *workspace.ID)

	scheduledQueryRule, err := createScheduledQueryRule(ctx, cred, *workspace.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("scheduled query rule:", *scheduledQueryRule.ID)

	scheduledQueryRule, err = getScheduledQueryRule(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get scheduled query rule:", *scheduledQueryRule.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createWorkspaces(ctx context.Context, cred azcore.TokenCredential) (*armoperationalinsights.Workspace, error) {
	workspacesClient, err := armoperationalinsights.NewWorkspacesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := workspacesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		workspaceName,
		armoperationalinsights.Workspace{
			Location: to.Ptr(location),
			Tags: map[string]*string{
				"tag1": to.Ptr("value1"),
			},
			Properties: &armoperationalinsights.WorkspaceProperties{
				SKU: &armoperationalinsights.WorkspaceSKU{
					Name: to.Ptr(armoperationalinsights.WorkspaceSKUNameEnumPerNode),
				},
				RetentionInDays: to.Ptr[int32](30),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.Workspace, nil
}

func createScheduledQueryRule(ctx context.Context, cred azcore.TokenCredential, workspaceID string) (*armmonitor.LogSearchRuleResource, error) {
	scheduledQueryRulesClient, err := armmonitor.NewScheduledQueryRulesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := scheduledQueryRulesClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		ruleName,
		armmonitor.LogSearchRuleResource{
			Location: to.Ptr(location),
			Properties: &armmonitor.LogSearchRule{
				Action: &armmonitor.AlertingAction{
					Severity: to.Ptr(armmonitor.AlertSeverityOne),
					Trigger: &armmonitor.TriggerCondition{
						Threshold:         to.Ptr[float64](3),
						ThresholdOperator: to.Ptr(armmonitor.ConditionalOperatorGreaterThan),
						MetricTrigger: &armmonitor.LogMetricTrigger{
							MetricColumn:      to.Ptr("Computer"),
							MetricTriggerType: to.Ptr(armmonitor.MetricTriggerTypeConsecutive),
							Threshold:         to.Ptr[float64](5),
							ThresholdOperator: to.Ptr(armmonitor.ConditionalOperatorGreaterThan),
						},
					},
				},
				Source: &armmonitor.Source{
					DataSourceID: to.Ptr(workspaceID),
					Query:        to.Ptr("Heartbeat | summarize AggregatedValue = count() by bin(TimeGenerated, 5m)"),
					QueryType:    to.Ptr(armmonitor.QueryTypeResultCount),
				},
				Description: to.Ptr("log search rule description"),
				Enabled:     to.Ptr(armmonitor.EnabledTrue),
				Schedule: &armmonitor.Schedule{
					FrequencyInMinutes:  to.Ptr[int32](15),
					TimeWindowInMinutes: to.Ptr[int32](15),
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.LogSearchRuleResource, nil
}

func getScheduledQueryRule(ctx context.Context, cred azcore.TokenCredential) (*armmonitor.LogSearchRuleResource, error) {
	scheduledQueryRulesClient, err := armmonitor.NewScheduledQueryRulesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := scheduledQueryRulesClient.Get(ctx, resourceGroupName, ruleName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.LogSearchRuleResource, nil
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

	_, err = pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return err
	}
	return nil
}
