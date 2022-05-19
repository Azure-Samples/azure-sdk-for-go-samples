// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"log"
	"os"
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
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createRegistry(ctx context.Context, cred azcore.TokenCredential) (*armcontainerregistry.Registry, error) {
	registriesClient, err := armcontainerregistry.NewRegistriesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := registriesClient.BeginCreate(
		ctx,
		resourceGroupName,
		registryName,
		armcontainerregistry.Registry{
			Location: to.Ptr(location),
			Tags: map[string]*string{
				"key": to.Ptr("value"),
			},
			SKU: &armcontainerregistry.SKU{
				Name: to.Ptr(armcontainerregistry.SKUNamePremium),
			},
			Properties: &armcontainerregistry.RegistryProperties{
				AdminUserEnabled: to.Ptr(true),
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
	return &resp.Registry, nil
}

func createTask(ctx context.Context, cred azcore.TokenCredential) (*armcontainerregistry.Task, error) {
	tasksClient, err := armcontainerregistry.NewTasksClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := tasksClient.BeginCreate(
		ctx,
		resourceGroupName,
		registryName,
		taskName,
		armcontainerregistry.Task{
			Location: to.Ptr(location),
			Properties: &armcontainerregistry.TaskProperties{
				Status: to.Ptr(armcontainerregistry.TaskStatusEnabled),
				Platform: &armcontainerregistry.PlatformProperties{
					OS:           to.Ptr(armcontainerregistry.OSLinux),
					Architecture: to.Ptr(armcontainerregistry.ArchitectureAmd64),
				},
				AgentConfiguration: &armcontainerregistry.AgentProperties{
					CPU: to.Ptr[int32](2),
				},
				Step: &armcontainerregistry.DockerBuildStep{
					Type:        to.Ptr(armcontainerregistry.StepTypeDocker),
					ContextPath: to.Ptr("https://github.com/SteveLasker/node-helloworld"),
					ImageNames: []*string{
						to.Ptr("testtask:v1"),
					},
					DockerFilePath: to.Ptr("Dockerfile"),
					IsPushEnabled:  to.Ptr(true),
					NoCache:        to.Ptr(false),
				},
				Trigger: &armcontainerregistry.TriggerProperties{
					BaseImageTrigger: &armcontainerregistry.BaseImageTrigger{
						Name:                     to.Ptr("myBaseImageTrigger"),
						BaseImageTriggerType:     to.Ptr(armcontainerregistry.BaseImageTriggerTypeRuntime),
						UpdateTriggerPayloadType: to.Ptr(armcontainerregistry.UpdateTriggerPayloadTypeDefault),
						Status:                   to.Ptr(armcontainerregistry.TriggerStatusEnabled),
					},
				},
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
	return &resp.Task, nil
}

func getTask(ctx context.Context, cred azcore.TokenCredential) (*armcontainerregistry.Task, error) {
	tasksClient, err := armcontainerregistry.NewTasksClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := tasksClient.Get(ctx, resourceGroupName, registryName, taskName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Task, nil
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
