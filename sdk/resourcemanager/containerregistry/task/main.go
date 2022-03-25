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
	ctx := context.Background()

	resourceGroup, err := createResourceGroup(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	registry, err := createRegistry(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("registry:", *registry.ID)

	task, err := createTask(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("task:", *task.ID)

	task, err = getTask(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get task:", *task.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createRegistry(ctx context.Context, cred azcore.TokenCredential) (*armcontainerregistry.Registry, error) {
	registriesClient := armcontainerregistry.NewRegistriesClient(subscriptionID, cred, nil)

	pollerResp, err := registriesClient.BeginCreate(
		ctx,
		resourceGroupName,
		registryName,
		armcontainerregistry.Registry{
			Location: to.StringPtr(location),
			Tags: map[string]*string{
				"key": to.StringPtr("value"),
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

func createTask(ctx context.Context, cred azcore.TokenCredential) (*armcontainerregistry.Task, error) {
	tasksClient := armcontainerregistry.NewTasksClient(subscriptionID, cred, nil)

	pollerResp, err := tasksClient.BeginCreate(
		ctx,
		resourceGroupName,
		registryName,
		taskName,
		armcontainerregistry.Task{
			Location: to.StringPtr(location),
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
					Type:        armcontainerregistry.StepTypeDocker.ToPtr(),
					ContextPath: to.StringPtr("https://github.com/SteveLasker/node-helloworld"),
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

func getTask(ctx context.Context, cred azcore.TokenCredential) (*armcontainerregistry.Task, error) {
	tasksClient := armcontainerregistry.NewTasksClient(subscriptionID, cred, nil)

	resp, err := tasksClient.Get(ctx, resourceGroupName, registryName, taskName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Task, nil
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
