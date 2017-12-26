package containerinstance

import (
	"fmt"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/services/containerinstance/mgmt/2017-08-01-preview/containerinstance"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

func getContainerGroupsClient() (containerinstance.ContainerGroupsClient, error) {
	token, err := iam.GetResourceManagementToken(iam.OAuthGrantTypeServicePrincipal)
	if err != nil {
		return containerinstance.ContainerGroupsClient{}, fmt.Errorf("cannot get token: %v", err)
	}

	containerGroupsClient := containerinstance.NewContainerGroupsClient(helpers.SubscriptionID())
	containerGroupsClient.Authorizer = autorest.NewBearerAuthorizer(token)

	return containerGroupsClient, nil
}

// CreateContainerGroup creates a new container group given a container group name, location and resoruce group
func CreateContainerGroup(containerGroupName, location, resourceGroupName string) (c containerinstance.ContainerGroup, err error) {

	containerGroupsClient, err := getContainerGroupsClient()
	if err != nil {
		return c, fmt.Errorf("cannot get container group client: %v", err)
	}

	parameters := containerinstance.ContainerGroup{
		Name:     &containerGroupName,
		Location: &location,
		ContainerGroupProperties: &containerinstance.ContainerGroupProperties{
			IPAddress: &containerinstance.IPAddress{
				Type: to.StringPtr("Public"),
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
					Name: to.StringPtr("az-samples-go-container"),
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
	}

	c, err = containerGroupsClient.CreateOrUpdate(resourceGroupName, containerGroupName, parameters)
	if err != nil {
		log.Fatalf("cannot create container group: %v", err)
	}

	return c, nil
}

// GetContainerGroup returns an existing container group given a resource group name and container group name
func GetContainerGroup(resourceGroupName, containerGroupName string) (c containerinstance.ContainerGroup, err error) {

	containerGroupsClient, err := getContainerGroupsClient()
	if err != nil {
		return c, fmt.Errorf("cannot get container group client: %v", err)
	}

	c, err = containerGroupsClient.Get(resourceGroupName, containerGroupName)
	if err != nil {
		return c, fmt.Errorf("cannot get container group %v from resource group %v: %v", containerGroupName, resourceGroupName, err)
	}

	return c, nil
}
