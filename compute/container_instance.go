// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package compute

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
	"github.com/Azure/azure-sdk-for-go/services/containerinstance/mgmt/2018-10-01/containerinstance"
	"github.com/Azure/go-autorest/autorest/to"
)

func getContainerGroupsClient() (containerinstance.ContainerGroupsClient, error) {
	containerGroupsClient := containerinstance.NewContainerGroupsClient(config.SubscriptionID())
	auth, _ := iam.GetResourceManagementAuthorizer()
	containerGroupsClient.Authorizer = auth
	containerGroupsClient.AddToUserAgent(config.UserAgent())
	return containerGroupsClient, nil
}

// CreateContainerGroup creates a new container group given a container group name, location and resoruce group
func CreateContainerGroup(ctx context.Context, containerGroupName, location, resourceGroupName string) (c containerinstance.ContainerGroup, err error) {
	containerGroupsClient, err := getContainerGroupsClient()
	if err != nil {
		return c, fmt.Errorf("cannot get container group client: %v", err)
	}

	future, err := containerGroupsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		containerGroupName,
		containerinstance.ContainerGroup{
			Name:     &containerGroupName,
			Location: &location,
			ContainerGroupProperties: &containerinstance.ContainerGroupProperties{
				IPAddress: &containerinstance.IPAddress{
					Type: containerinstance.Public,
					Ports: &[]containerinstance.Port{
						{
							Port:     to.Int32Ptr(80),
							Protocol: containerinstance.TCP,
						},
					},
				},
				OsType: containerinstance.Linux,
				Containers: &[]containerinstance.Container{
					{
						Name: to.StringPtr("gosdk-container"),
						ContainerProperties: &containerinstance.ContainerProperties{
							Ports: &[]containerinstance.ContainerPort{
								{
									Port: to.Int32Ptr(80),
								},
							},
							Image: to.StringPtr("nginx:latest"),
							Resources: &containerinstance.ResourceRequirements{
								Limits: &containerinstance.ResourceLimits{
									MemoryInGB: to.Float64Ptr(1),
									CPU:        to.Float64Ptr(1),
								},
								Requests: &containerinstance.ResourceRequests{
									MemoryInGB: to.Float64Ptr(1),
									CPU:        to.Float64Ptr(1),
								},
							},
						},
					},
				},
			},
		})

	if err != nil {
		log.Fatalf("cannot create container group: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, containerGroupsClient.Client)
	if err != nil {
		log.Fatalf("cannot create container group: %v", err)
	}
	return future.Result(containerGroupsClient)
}

// GetContainerGroup returns an existing container group given a resource group name and container group name
func GetContainerGroup(ctx context.Context, resourceGroupName, containerGroupName string) (c containerinstance.ContainerGroup, err error) {
	containerGroupsClient, err := getContainerGroupsClient()
	if err != nil {
		return c, fmt.Errorf("cannot get container group client: %v", err)
	}

	c, err = containerGroupsClient.Get(ctx, resourceGroupName, containerGroupName)
	if err != nil {
		return c, fmt.Errorf("cannot get container group %v from resource group %v: %v", containerGroupName, resourceGroupName, err)
	}

	return c, nil
}

// UpdateContainerGroup updates the image of the first container of an existing container group
// given a resource group name and container group name
func UpdateContainerGroup(ctx context.Context, resourceGroupName, containerGroupName string) (c containerinstance.ContainerGroup, err error) {
	containerGroupsClient, err := getContainerGroupsClient()
	if err != nil {
		return c, fmt.Errorf("cannot get container group client: %v", err)
	}

	c, err = GetContainerGroup(ctx, resourceGroupName, containerGroupName)
	if err != nil {
		return c, fmt.Errorf("cannot get container group %v from resource group %v: %v", containerGroupName, resourceGroupName, err)
	}
	// updating the image of the first container in the group
	// here you can also update other properties of the container group
	(*c.Containers)[0].Image = to.StringPtr("microsoft/aci-helloworld")

	future, err := containerGroupsClient.CreateOrUpdate(context.Background(), resourceGroupName, containerGroupName, c)

	if err != nil {
		log.Fatalf("cannot create container group: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, containerGroupsClient.Client)
	if err != nil {
		log.Fatalf("cannot create container group: %v", err)
	}
	return future.Result(containerGroupsClient)
}

// DeleteContainerGroup deletes an existing container group given a resource group name and container group name
func DeleteContainerGroup(ctx context.Context, resourceGroupName, containerGroupName string) (c containerinstance.ContainerGroup, err error) {
	containerGroupsClient, err := getContainerGroupsClient()
	if err != nil {
		return c, fmt.Errorf("cannot get container group client: %v", err)
	}

	return containerGroupsClient.Delete(ctx, resourceGroupName, containerGroupName)
}
