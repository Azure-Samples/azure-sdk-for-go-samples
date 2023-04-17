// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/operationalinsights/armoperationalinsights"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"log"
	"os"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	workspaceName     = "sample-workspace"
	ruleName          = "sample-scheduled-query-rules"
)

var (
	resourcesClientFactory           *armresources.ClientFactory
	operationalinsightsClientFactory *armoperationalinsights.ClientFactory
	monitorClientFactory             *armmonitor.ClientFactory
)

var (
	resourceGroupClient       *armresources.ResourceGroupsClient
	workspacesClient          *armoperationalinsights.WorkspacesClient
	scheduledQueryRulesClient *armmonitor.ScheduledQueryRulesClient
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

	operationalinsightsClientFactory, err = armoperationalinsights.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	workspacesClient = operationalinsightsClientFactory.NewWorkspacesClient()

	monitorClientFactory, err = armmonitor.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	scheduledQueryRulesClient = monitorClientFactory.NewScheduledQueryRulesClient()

	resourceGroup, err := createResourceGroup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	workspace, err := createWorkspaces(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("workspace:", *workspace.ID)

	scheduledQueryRule, err := createScheduledQueryRule(ctx, *workspace.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("scheduled query rule:", *scheduledQueryRule.ID)

	scheduledQueryRule, err = getScheduledQueryRule(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get scheduled query rule:", *scheduledQueryRule.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createWorkspaces(ctx context.Context) (*armoperationalinsights.Workspace, error) {

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
	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Workspace, nil
}

func createScheduledQueryRule(ctx context.Context, workspaceID string) (*armmonitor.ScheduledQueryRuleResource, error) {

	resp, err := scheduledQueryRulesClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		ruleName,
		armmonitor.ScheduledQueryRuleResource{
			Location: to.Ptr(location),
			Properties: &armmonitor.ScheduledQueryRuleProperties{
				Description: to.Ptr("Performance rule"),
				Actions: &armmonitor.Actions{
					ActionGroups: []*string{
						to.Ptr(workspaceID),
					},
					CustomProperties: map[string]*string{
						"key11": to.Ptr("value11"),
						"key12": to.Ptr("value12"),
					},
				},
				CheckWorkspaceAlertsStorageConfigured: to.Ptr(true),
				Criteria: &armmonitor.ScheduledQueryRuleCriteria{
					AllOf: []*armmonitor.Condition{
						{
							Dimensions: []*armmonitor.Dimension{
								{
									Name:     to.Ptr("ComputerIp"),
									Operator: to.Ptr(armmonitor.DimensionOperatorExclude),
									Values: []*string{
										to.Ptr("192.168.1.1")},
								},
								{
									Name:     to.Ptr("OSType"),
									Operator: to.Ptr(armmonitor.DimensionOperatorInclude),
									Values: []*string{
										to.Ptr("*")},
								}},
							FailingPeriods: &armmonitor.ConditionFailingPeriods{
								MinFailingPeriodsToAlert:  to.Ptr[int64](1),
								NumberOfEvaluationPeriods: to.Ptr[int64](1),
							},
							MetricMeasureColumn: to.Ptr("% Processor Time"),
							Operator:            to.Ptr(armmonitor.ConditionOperatorGreaterThan),
							Query:               to.Ptr("Perf | where ObjectName == \"Processor\""),
							ResourceIDColumn:    to.Ptr("resourceId"),
							Threshold:           to.Ptr[float64](70),
							TimeAggregation:     to.Ptr(armmonitor.TimeAggregationAverage),
						}},
				},
				Enabled:             to.Ptr(true),
				EvaluationFrequency: to.Ptr("PT5M"),
				MuteActionsDuration: to.Ptr("PT30M"),
				RuleResolveConfiguration: &armmonitor.RuleResolveConfiguration{
					AutoResolved:  to.Ptr(true),
					TimeToResolve: to.Ptr("PT10M"),
				},
				Severity:            to.Ptr(armmonitor.AlertSeverity(4)),
				SkipQueryValidation: to.Ptr(true),
				WindowSize:          to.Ptr("PT10M"),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.ScheduledQueryRuleResource, nil
}

func getScheduledQueryRule(ctx context.Context) (*armmonitor.ScheduledQueryRuleResource, error) {

	resp, err := scheduledQueryRulesClient.Get(ctx, resourceGroupName, ruleName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ScheduledQueryRuleResource, nil
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
