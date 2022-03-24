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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicefabric/armservicefabric"
)

var (
	subscriptionID    string
	location          = "eastus"
	resourceGroupName = "sample-resource-group"
	clusterName = "sample-servicefabric-cluster"
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

	cluster, err := createCluster(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service fabric cluster:", *cluster.ID)

	cluster, err = getCluster(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get service fabric cluster:", *cluster.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createCluster(ctx context.Context,cred azcore.TokenCredential) (*armservicefabric.Cluster,error) {
	clustersClient := armservicefabric.NewClustersClient(subscriptionID,cred,nil)
	pollerResp,err := clustersClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		clusterName,
		armservicefabric.Cluster{
			Location: to.StringPtr(location),
			Properties: &armservicefabric.ClusterProperties{
				ManagementEndpoint: to.StringPtr("https://myCluster.eastus.cloudapp.azure.com:19080"),
				NodeTypes: []*armservicefabric.NodeTypeDescription{
					{
						Name: to.StringPtr("nt1vm"),
						ClientConnectionEndpointPort: to.Int32Ptr(19000),
						HTTPGatewayEndpointPort: to.Int32Ptr(19007),
						ApplicationPorts: &armservicefabric.EndpointRangeDescription{
							StartPort: to.Int32Ptr(20000),
							EndPort: to.Int32Ptr(30000),
						},
						EphemeralPorts: &armservicefabric.EndpointRangeDescription{
							StartPort: to.Int32Ptr(49000),
							EndPort: to.Int32Ptr(64000),
						},
						IsPrimary: to.BoolPtr(true),
						VMInstanceCount: to.Int32Ptr(5),
						DurabilityLevel: armservicefabric.DurabilityLevelBronze.ToPtr(),
					},
				},
				FabricSettings: []*armservicefabric.SettingsSectionDescription{
					{
						Name: to.StringPtr("UpgradeService"),
						Parameters: []*armservicefabric.SettingsParameterDescription{
							{
								Name: to.StringPtr("AppPollIntervalInSeconds"),
								Value: to.StringPtr("60"),
							},
						},
					},
				},
				DiagnosticsStorageAccountConfig: &armservicefabric.DiagnosticsStorageAccountConfig{
					StorageAccountName: to.StringPtr("diag"),
					ProtectedAccountKeyName: to.StringPtr("StorageAccountKey1"),
					BlobEndpoint: to.StringPtr("https://diag.blob.core.windows.net/"),
					QueueEndpoint: to.StringPtr("https://diag.queue.core.windows.net/"),
					TableEndpoint: to.StringPtr("https://diag.table.core.windows.net/"),
				},
				ReliabilityLevel: armservicefabric.ReliabilityLevelSilver.ToPtr(),
				UpgradeMode: armservicefabric.UpgradeModeAutomatic.ToPtr(),
			},
		},
		nil,
	)
	if err != nil {
		return nil,err
	}
	resp,err := pollerResp.PollUntilDone(ctx,30*time.Second)
	if err != nil {
		return nil,err
	}
	return &resp.Cluster,nil
}

func getCluster(ctx context.Context,cred azcore.TokenCredential) (*armservicefabric.Cluster,error) {
	clustersClient := armservicefabric.NewClustersClient(subscriptionID,cred,nil)
	resp,err := clustersClient.Get(ctx, resourceGroupName, clusterName, nil)
	if err != nil {
		return nil,err
	}
	return &resp.Cluster,nil
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
