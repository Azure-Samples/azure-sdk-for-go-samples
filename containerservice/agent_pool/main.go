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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	ObjectID          string
	ClientSecret      string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	agentPool         = "sample-aks"
	agentPoolName     = "sample-aks-agent-pool"
)

func main() {
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set.")
	}

	ObjectID = os.Getenv("AZURE_OBJECT_ID")
	if len(ObjectID) == 0 {
		log.Fatal("AZURE_OBJECT_ID is not set.")
	}

	ClientSecret = os.Getenv("AZURE_CLIENT_SECRET")
	if len(ObjectID) == 0 {
		log.Fatal("AZURE_CLIENT_SECRET is not set.")
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

	managedCluster, err := createManagedCluster(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("managed cluster:", *managedCluster.ID)

	agentPool, err := createAgentPool(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("agent pool:", *agentPool.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createManagedCluster(ctx context.Context, conn *arm.Connection) (*armcontainerservice.ManagedCluster, error) {
	managedClustersClient := armcontainerservice.NewManagedClustersClient(conn, subscriptionID)

	pollerResp, err := managedClustersClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		agentPool,
		armcontainerservice.ManagedCluster{
			Resource: armcontainerservice.Resource{
				Location: to.StringPtr(location),
			},
			Properties: &armcontainerservice.ManagedClusterProperties{
				DNSPrefix: to.StringPtr("aksgosdk"),
				AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
					{
						Name: to.StringPtr("askagent"),
						ManagedClusterAgentPoolProfileProperties: armcontainerservice.ManagedClusterAgentPoolProfileProperties{
							Count:             to.Int32Ptr(1),
							VMSize:            to.StringPtr("Standard_DS2_v2"),
							MaxPods:           to.Int32Ptr(110),
							MinCount:          to.Int32Ptr(1),
							MaxCount:          to.Int32Ptr(100),
							OSType:            armcontainerservice.OSTypeLinux.ToPtr(),
							Type:              armcontainerservice.AgentPoolTypeVirtualMachineScaleSets.ToPtr(),
							EnableAutoScaling: to.BoolPtr(true),
							Mode:              armcontainerservice.AgentPoolModeSystem.ToPtr(),
						},
					},
				},
				ServicePrincipalProfile: &armcontainerservice.ManagedClusterServicePrincipalProfile{
					ClientID: to.StringPtr(ObjectID),
					Secret:   to.StringPtr(ClientSecret),
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
	return &resp.ManagedCluster, nil
}

func createAgentPool(ctx context.Context, conn *arm.Connection) (*armcontainerservice.AgentPool, error) {
	agentPoolsClient := armcontainerservice.NewAgentPoolsClient(conn, subscriptionID)

	pollerResp, err := agentPoolsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		agentPool,
		agentPoolName,
		armcontainerservice.AgentPool{
			Properties: &armcontainerservice.ManagedClusterAgentPoolProfileProperties{
				OrchestratorVersion: to.StringPtr(""),
				Count:               to.Int32Ptr(3),
				VMSize:              to.StringPtr("Standard_DS2_v2"),
				OSType:              armcontainerservice.OSTypeLinux.ToPtr(),
				Mode:                armcontainerservice.AgentPoolModeSystem.ToPtr(),
				AvailabilityZones: []*string{
					to.StringPtr("1"),
					to.StringPtr("2"),
					to.StringPtr("3"),
				},
				NodeTaints: []*string{},
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
	return &resp.AgentPool, nil
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
