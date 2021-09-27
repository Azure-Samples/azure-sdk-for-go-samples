package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/operationalinsights/armoperationalinsights"
	"github.com/Azure/azure-sdk-for-go/sdk/resources/armresources"
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

	conn := arm.NewDefaultConnection(cred, &arm.ConnectionOptions{
		Logging: policy.LogOptions{
			IncludeBody: true,
		},
	})
	ctx := context.Background()

	resourceGroup, err := createResourceGroup(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	workspace, err := createWorkspaces(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("workspace:", *workspace.ID)

	scheduledQueryRule, err := createScheduledQueryRule(ctx, conn, *workspace.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("scheduled query rule:", *scheduledQueryRule.ID)

	scheduledQueryRule, err = getScheduledQueryRule(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get scheduled query rule:", *scheduledQueryRule.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createWorkspaces(ctx context.Context, conn *arm.Connection) (*armoperationalinsights.Workspace, error) {
	workspacesClient := armoperationalinsights.NewWorkspacesClient(conn, subscriptionID)

	pollerResp, err := workspacesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		workspaceName,
		armoperationalinsights.Workspace{
			TrackedResource: armoperationalinsights.TrackedResource{
				Location: to.StringPtr(location),
				Tags: map[string]*string{
					"tag1": to.StringPtr("value1"),
				},
			},
			Properties: &armoperationalinsights.WorkspaceProperties{
				SKU: &armoperationalinsights.WorkspaceSKU{
					Name: armoperationalinsights.WorkspaceSKUNameEnumPerNode.ToPtr(),
				},
				RetentionInDays: to.Int32Ptr(30),
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

func createScheduledQueryRule(ctx context.Context, conn *arm.Connection, workspaceID string) (*armmonitor.LogSearchRuleResource, error) {
	scheduledQueryRulesClient := armmonitor.NewScheduledQueryRulesClient(conn, subscriptionID)

	resp, err := scheduledQueryRulesClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		ruleName,
		armmonitor.LogSearchRuleResource{
			TrackedEntityResource: armmonitor.TrackedEntityResource{
				Location: to.StringPtr(location),
			},
			Properties: &armmonitor.LogSearchRule{
				Action: &armmonitor.AlertingAction{
					Severity: armmonitor.AlertSeverityOne.ToPtr(),
					Trigger: &armmonitor.TriggerCondition{
						Threshold:         to.Float64Ptr(3),
						ThresholdOperator: armmonitor.ConditionalOperatorGreaterThan.ToPtr(),
						MetricTrigger: &armmonitor.LogMetricTrigger{
							MetricColumn:      to.StringPtr("Computer"),
							MetricTriggerType: armmonitor.MetricTriggerTypeConsecutive.ToPtr(),
							Threshold:         to.Float64Ptr(5),
							ThresholdOperator: armmonitor.ConditionalOperatorGreaterThan.ToPtr(),
						},
					},
				},
				Source: &armmonitor.Source{
					DataSourceID: to.StringPtr(workspaceID),
					Query:        to.StringPtr("Heartbeat | summarize AggregatedValue = count() by bin(TimeGenerated, 5m)"),
					QueryType:    armmonitor.QueryTypeResultCount.ToPtr(),
				},
				Description: to.StringPtr("log search rule description"),
				Enabled:     armmonitor.EnabledTrue.ToPtr(),
				Schedule: &armmonitor.Schedule{
					FrequencyInMinutes:  to.Int32Ptr(15),
					TimeWindowInMinutes: to.Int32Ptr(15),
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

func getScheduledQueryRule(ctx context.Context, conn *arm.Connection) (*armmonitor.LogSearchRuleResource, error) {
	scheduledQueryRulesClient := armmonitor.NewScheduledQueryRulesClient(conn, subscriptionID)

	resp, err := scheduledQueryRulesClient.Get(ctx, resourceGroupName, ruleName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.LogSearchRuleResource, nil
}

func createResourceGroup(ctx context.Context, conn *arm.Connection) (*armresources.ResourceGroup, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn, subscriptionID)

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

func cleanup(ctx context.Context, conn *arm.Connection) (*http.Response, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn, subscriptionID)

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
