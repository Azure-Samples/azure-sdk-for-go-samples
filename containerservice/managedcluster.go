package containerservice

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2017-09-30/containerservice"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

// fakepubkey is used if a key isn't available at the specified path in CreateAKS(...)
var fakepubkey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC7laRyN4B3YZmVrDEZLZoIuUA72pQ0DpGuZBZWykCofIfCPrFZAJgFvonKGgKJl6FGKIunkZL9Us/mV4ZPkZhBlE7uX83AAf5i9Q8FmKpotzmaxN10/1mcnEE7pFvLoSkwqrQSkrrgSm8zaJ3g91giXSbtqvSIj/vk2f05stYmLfhAwNo3Oh27ugCakCoVeuCrZkvHMaJgcYrIGCuFo6q0Pfk9rsZyriIqEa9AtiUOtViInVYdby7y71wcbl0AbbCZsTSqnSoVxm2tRkOsXV6+8X4SnwcmZbao3H+zfO1GBhQOLxJ4NQbzAa8IJh810rYARNLptgmsd4cYXVOSosTX azureuser"

func getAKSClient() (containerservice.ManagedClustersClient, error) {
	token, err := iam.GetResourceManagementToken(iam.AuthGrantType())
	if err != nil {
		return containerservice.ManagedClustersClient{}, fmt.Errorf("cannot get token: %v", err)
	}

	aksClient := containerservice.NewManagedClustersClient(internal.SubscriptionID())
	aksClient.Authorizer = autorest.NewBearerAuthorizer(token)
	aksClient.AddToUserAgent(internal.UserAgent())
	return aksClient, nil
}

// CreateAKS creates a new managed Kubernetes cluster
func CreateAKS(ctx context.Context, resourceName, location, resourceGroupName, username, sshPublicKeyPath, clientID, clientSecret string, agentPoolCount int32) (c containerservice.ManagedCluster, err error) {
	var sshKeyData string
	if _, err = os.Stat(sshPublicKeyPath); err == nil {
		sshBytes, err := ioutil.ReadFile(sshPublicKeyPath)
		if err != nil {
			log.Fatalf("failed to read SSH key data: %v", err)
		}
		sshKeyData = string(sshBytes)
	} else {
		sshKeyData = fakepubkey
	}

	aksClient, err := getAKSClient()
	if err != nil {
		return c, fmt.Errorf("cannot get AKS client: %v", err)
	}

	future, err := aksClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		resourceName,
		containerservice.ManagedCluster{
			Name:     &resourceName,
			Location: &location,
			ManagedClusterProperties: &containerservice.ManagedClusterProperties{
				DNSPrefix: &resourceName,
				LinuxProfile: &containerservice.LinuxProfile{
					AdminUsername: to.StringPtr(username),
					SSH: &containerservice.SSHConfiguration{
						PublicKeys: &[]containerservice.SSHPublicKey{
							{
								KeyData: to.StringPtr(sshKeyData),
							},
						},
					},
				},
				AgentPoolProfiles: &[]containerservice.AgentPoolProfile{
					{
						Count:  to.Int32Ptr(agentPoolCount),
						Name:   to.StringPtr("agentpool1"),
						VMSize: containerservice.StandardD2V2,
					},
				},
				ServicePrincipalProfile: &containerservice.ServicePrincipalProfile{
					ClientID: to.StringPtr(clientID),
					Secret:   to.StringPtr(clientSecret),
				},
			},
		},
	)
	if err != nil {
		return c, fmt.Errorf("cannot create AKS cluster: %v", err)
	}

	err = future.WaitForCompletion(ctx, aksClient.Client)
	if err != nil {
		return c, fmt.Errorf("cannot get the AKS cluster create or update future response: %v", err)
	}

	return future.Result(aksClient)
}

// GetAKS returns an existing AKS cluster given a resource group name and resource name
func GetAKS(ctx context.Context, resourceGroupName, resourceName string) (c containerservice.ManagedCluster, err error) {
	aksClient, err := getAKSClient()
	if err != nil {
		return c, fmt.Errorf("cannot get AKS client: %v", err)
	}

	c, err = aksClient.Get(ctx, resourceGroupName, resourceName)
	if err != nil {
		return c, fmt.Errorf("cannot get AKS managed cluster %v from resource group %v: %v", resourceName, resourceGroupName, err)
	}

	return c, nil
}

// DeleteAKS deletes an existing AKS cluster
func DeleteAKS(ctx context.Context, resourceGroupName, resourceName string) (c containerservice.ManagedClustersDeleteFuture, err error) {
	aksClient, err := getAKSClient()
	if err != nil {
		return c, fmt.Errorf("cannot get AKS client: %v", err)
	}

	return aksClient.Delete(ctx, resourceGroupName, resourceName)
}
