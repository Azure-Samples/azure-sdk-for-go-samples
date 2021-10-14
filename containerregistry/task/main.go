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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	registryName      = "sample2registry"
	taskName          = "sample-run"
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

	registry, err := createRegistry(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("registry:", *registry.ID)

	task, err := createTask(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("task:", *task.ID)

	task, err = getTask(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get task:", *task.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createRegistry(ctx context.Context, conn *arm.Connection) (*armcontainerregistry.Registry, error) {
	registriesClient := armcontainerregistry.NewRegistriesClient(conn, subscriptionID)

	pollerResp, err := registriesClient.BeginCreate(
		ctx,
		resourceGroupName,
		registryName,
		armcontainerregistry.Registry{
			Resource: armcontainerregistry.Resource{
				Location: to.StringPtr(location),
				Tags: map[string]*string{
					"key": to.StringPtr("value"),
				},
			},
			SKU: &armcontainerregistry.SKU{
				Name: armcontainerregistry.SKUNamePremium.ToPtr(),
			},
			Properties: &armcontainerregistry.RegistryProperties{
				AdminUserEnabled: to.BoolPtr(true),
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
	return &resp.Registry, nil
}

func createTask(ctx context.Context, conn *arm.Connection) (*armcontainerregistry.Task, error) {
	tasksClient := armcontainerregistry.NewTasksClient(conn, subscriptionID)

	pollerResp, err := tasksClient.BeginCreate(
		ctx,
		resourceGroupName,
		registryName,
		taskName,
		armcontainerregistry.Task{
			Resource: armcontainerregistry.Resource{
				Location: to.StringPtr(location),
			},
			Properties: &armcontainerregistry.TaskProperties{
				Status: armcontainerregistry.TaskStatusEnabled.ToPtr(),
				Platform: &armcontainerregistry.PlatformProperties{
					OS:           armcontainerregistry.OSLinux.ToPtr(),
					Architecture: armcontainerregistry.ArchitectureAmd64.ToPtr(),
				},
				AgentConfiguration: &armcontainerregistry.AgentProperties{
					CPU: to.Int32Ptr(2),
				},
				Step: &armcontainerregistry.DockerBuildStep{
					TaskStepProperties: armcontainerregistry.TaskStepProperties{
						Type:        armcontainerregistry.StepTypeDocker.ToPtr(),
						ContextPath: to.StringPtr("https://github.com/SteveLasker/node-helloworld"),
					},
					ImageNames: []*string{
						to.StringPtr("testtask:v1"),
					},
					DockerFilePath: to.StringPtr("Dockerfile"),
					IsPushEnabled:  to.BoolPtr(true),
					NoCache:        to.BoolPtr(false),
				},
				Trigger: &armcontainerregistry.TriggerProperties{
					BaseImageTrigger: &armcontainerregistry.BaseImageTrigger{
						Name:                     to.StringPtr("myBaseImageTrigger"),
						BaseImageTriggerType:     armcontainerregistry.BaseImageTriggerTypeRuntime.ToPtr(),
						UpdateTriggerPayloadType: armcontainerregistry.UpdateTriggerPayloadTypeDefault.ToPtr(),
						Status:                   armcontainerregistry.TriggerStatusEnabled.ToPtr(),
					},
				},
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
	return &resp.Task, nil
}

func getTask(ctx context.Context, conn *arm.Connection) (*armcontainerregistry.Task, error) {
	tasksClient := armcontainerregistry.NewTasksClient(conn, subscriptionID)

	resp, err := tasksClient.Get(ctx, resourceGroupName, registryName, taskName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Task, nil
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
