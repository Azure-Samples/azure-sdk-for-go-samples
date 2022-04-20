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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	registryName      = "sample2registry"
	taskRunName       = "sample-task-run"
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

	taskRun, err := createTaskRun(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("task run:", *taskRun.ID)

	taskRun, err = getTaskRun(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get task run:", *taskRun.ID)

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
	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.Registry, nil
}

func createTaskRun(ctx context.Context, cred azcore.TokenCredential) (*armcontainerregistry.TaskRun, error) {
	taskRunsClient, err := armcontainerregistry.NewTaskRunsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := taskRunsClient.BeginCreate(
		ctx,
		resourceGroupName,
		registryName,
		taskRunName,
		armcontainerregistry.TaskRun{
			Properties: &armcontainerregistry.TaskRunProperties{
				ForceUpdateTag: to.Ptr("test"),
				RunRequest: &armcontainerregistry.DockerBuildRequest{
					IsArchiveEnabled: to.Ptr(true),
					DockerFilePath:   to.Ptr("Dockerfile"),
					Platform: &armcontainerregistry.PlatformProperties{
						OS:           to.Ptr(armcontainerregistry.OSLinux),
						Architecture: to.Ptr(armcontainerregistry.ArchitectureAmd64),
					},
					ImageNames: []*string{
						to.Ptr("testtaskrun:v1"),
					},
					IsPushEnabled:  to.Ptr(true),
					NoCache:        to.Ptr(false),
					SourceLocation: to.Ptr("https://github.com/Azure-Samples/acr-build-helloworld-node.git"),
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
	return &resp.TaskRun, nil
}

func getTaskRun(ctx context.Context, cred azcore.TokenCredential) (*armcontainerregistry.TaskRun, error) {
	taskRunsClient, err := armcontainerregistry.NewTaskRunsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := taskRunsClient.Get(ctx, resourceGroupName, registryName, taskRunName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.TaskRun, nil
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
