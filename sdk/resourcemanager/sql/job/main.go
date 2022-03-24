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
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createServer(ctx context.Context, cred azcore.TokenCredential) (*armsql.Server, error) {
	serversClient := armsql.NewServersClient(subscriptionID, cred, nil)

	pollerResp, err := serversClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		armsql.Server{
			Location: to.StringPtr(location),
			Properties: &armsql.ServerProperties{
				AdministratorLogin:         to.StringPtr("dummylogin"),
				AdministratorLoginPassword: to.StringPtr("QWE123!@#"),
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
	databasesClient := armsql.NewDatabasesClient(subscriptionID, cred, nil)

	pollerResp, err := databasesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		databaseName,
		armsql.Database{
			Location: to.StringPtr(location),
			Properties: &armsql.DatabaseProperties{
				ReadScale: armsql.DatabaseReadScaleDisabled.ToPtr(),
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
	jobAgentsClient := armsql.NewJobAgentsClient(subscriptionID, cred, nil)

	pollerResp, err := jobAgentsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		jobAgentName,
		armsql.JobAgent{
			Location: to.StringPtr(location),
			Properties: &armsql.JobAgentProperties{
				DatabaseID: to.StringPtr(databaseID),
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
	jobCredentialsClient := armsql.NewJobCredentialsClient(subscriptionID, cred, nil)

	resp, err := jobCredentialsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		jobAgentName,
		credentialName,
		armsql.JobCredential{
			Properties: &armsql.JobCredentialProperties{
				Username: to.StringPtr("dummylogin"),
				Password: to.StringPtr("QWE123!@#"),
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
	jobTargetGroupsClient := armsql.NewJobTargetGroupsClient(subscriptionID, cred, nil)

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
	jobsClient := armsql.NewJobsClient(subscriptionID, cred, nil)

	startTime, _ := time.Parse("2012-01-02 15:04:05 06", "2021-09-18T18:30:01Z")
	endTime, _ := time.Parse("2012-01-02 15:04:05 06", "2021-09-18T23:59:59Z")

	resp, err := jobsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		jobAgentName,
		jobName,
		armsql.Job{
			Properties: &armsql.JobProperties{
				Description: to.StringPtr("my favourite job"),
				Schedule: &armsql.JobSchedule{
					StartTime: to.TimePtr(startTime),
					EndTime:   to.TimePtr(endTime),
					Type:      armsql.JobScheduleTypeRecurring.ToPtr(),
					Interval:  to.StringPtr("PT5M"),
					Enabled:   to.BoolPtr(true),
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
	jobStepsClient := armsql.NewJobStepsClient(subscriptionID, cred, nil)

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
					Value: to.StringPtr("select 1"),
				},
				Credential:  to.StringPtr(credentialID),
				TargetGroup: to.StringPtr(targetGroupID),
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
	jobExecutionsClient := armsql.NewJobExecutionsClient(subscriptionID, cred, nil)

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
	return &resp.JobExecution, nil
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
