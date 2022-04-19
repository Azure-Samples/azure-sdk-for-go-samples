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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
)

var (
	subscriptionID    string
	location          = "eastus"
	resourceGroupName = "sample-resource-group"
	serverName        = "sample2server"
	databaseName      = "sample-database"
	jobAgentName      = "sample-job-agent"
	credentialName    = "sample-credential"
	targetGroupName   = "sample-target-group"
	jobName           = "sample-job"
	jobStepName       = "sample-job-step"
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

	server, err := createServer(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("server:", *server.ID)

	database, err := createDatabase(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("database:", *database.ID)

	jobAgent, err := createJobAgent(ctx, cred, *database.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("job agent:", *jobAgent.ID)

	jobCredential, err := createJobCredential(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("job credential:", *jobCredential.ID)

	jobTargetGroup, err := createJobTargetGroup(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("job target group:", *jobTargetGroup.ID)

	job, err := createJob(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("job:", *job.ID)

	jobStep, err := createJobStep(ctx, cred, *jobCredential.ID, *jobTargetGroup.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("job step:", *jobStep.ID)

	jobExecution, err := createJobExecution(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("job execution:", *jobExecution.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createServer(ctx context.Context, cred azcore.TokenCredential) (*armsql.Server, error) {
	serversClient, err := armsql.NewServersClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := serversClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		armsql.Server{
			Location: to.Ptr(location),
			Properties: &armsql.ServerProperties{
				AdministratorLogin:         to.Ptr("dummylogin"),
				AdministratorLoginPassword: to.Ptr("QWE123!@#"),
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
	return &resp.Server, nil
}

func createDatabase(ctx context.Context, cred azcore.TokenCredential) (*armsql.Database, error) {
	databasesClient, err := armsql.NewDatabasesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := databasesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		databaseName,
		armsql.Database{
			Location: to.Ptr(location),
			Properties: &armsql.DatabaseProperties{
				ReadScale: to.Ptr(armsql.DatabaseReadScaleDisabled),
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
	return &resp.Database, nil
}

func createJobAgent(ctx context.Context, cred azcore.TokenCredential, databaseID string) (*armsql.JobAgent, error) {
	jobAgentsClient, err := armsql.NewJobAgentsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := jobAgentsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		jobAgentName,
		armsql.JobAgent{
			Location: to.Ptr(location),
			Properties: &armsql.JobAgentProperties{
				DatabaseID: to.Ptr(databaseID),
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
	return &resp.JobAgent, nil
}

func createJobCredential(ctx context.Context, cred azcore.TokenCredential) (*armsql.JobCredential, error) {
	jobCredentialsClient, err := armsql.NewJobCredentialsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := jobCredentialsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		jobAgentName,
		credentialName,
		armsql.JobCredential{
			Properties: &armsql.JobCredentialProperties{
				Username: to.Ptr("dummylogin"),
				Password: to.Ptr("QWE123!@#"),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.JobCredential, nil
}

func createJobTargetGroup(ctx context.Context, cred azcore.TokenCredential) (*armsql.JobTargetGroup, error) {
	jobTargetGroupsClient, err := armsql.NewJobTargetGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := jobTargetGroupsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		jobAgentName,
		targetGroupName,
		armsql.JobTargetGroup{
			Properties: &armsql.JobTargetGroupProperties{
				Members: []*armsql.JobTarget{},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.JobTargetGroup, nil
}

func createJob(ctx context.Context, cred azcore.TokenCredential) (*armsql.Job, error) {
	jobsClient, err := armsql.NewJobsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	startTime, _ := time.Parse("2006-01-02 15:04:05 06", "2021-09-18T18:30:01Z")
	endTime, _ := time.Parse("2006-01-02 15:04:05 06", "2021-09-18T23:59:59Z")

	resp, err := jobsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		jobAgentName,
		jobName,
		armsql.Job{
			Properties: &armsql.JobProperties{
				Description: to.Ptr("my favourite job"),
				Schedule: &armsql.JobSchedule{
					StartTime: to.Ptr(startTime),
					EndTime:   to.Ptr(endTime),
					Type:      to.Ptr(armsql.JobScheduleTypeRecurring),
					Interval:  to.Ptr("PT5M"),
					Enabled:   to.Ptr(true),
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.Job, nil
}

func createJobStep(ctx context.Context, cred azcore.TokenCredential, credentialID, targetGroupID string) (*armsql.JobStep, error) {
	jobStepsClient, err := armsql.NewJobStepsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := jobStepsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		jobAgentName,
		jobName,
		jobStepName,
		armsql.JobStep{
			Properties: &armsql.JobStepProperties{
				Action: &armsql.JobStepAction{
					Value: to.Ptr("select 1"),
				},
				Credential:  to.Ptr(credentialID),
				TargetGroup: to.Ptr(targetGroupID),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.JobStep, nil
}

func createJobExecution(ctx context.Context, cred azcore.TokenCredential) (*armsql.JobExecution, error) {
	jobExecutionsClient, err := armsql.NewJobExecutionsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := jobExecutionsClient.BeginCreate(
		ctx,
		resourceGroupName,
		serverName,
		jobAgentName,
		jobName,
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.JobExecution, nil
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
