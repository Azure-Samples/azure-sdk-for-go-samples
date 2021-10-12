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
	subscriptionID      string
	objectID            string
	clientSecret        string
	location            = "westus"
	resourceGroupName   = "sample-resource-group"
	managedClustersName = "sample-aks-cluster"
	configName          = "sample-aks-maintenance-config"
)

func main() {
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set.")
	}

	objectID = os.Getenv("AZURE_OBJECT_ID")
	if len(objectID) == 0 {
		log.Fatal("AZURE_OBJECT_ID is not set.")
	}

	clientSecret = os.Getenv("AZURE_CLIENT_SECRET")
	if len(clientSecret) == 0 {
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

	maintenanceConfiguration, err := createMaintenanceConfiguration(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("maintenance configuration:", *maintenanceConfiguration.ID)

	maintenanceConfigurations := listMaintenanceConfiguration(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("List By Managed Cluster:", managedClustersName)
	for _, mc := range maintenanceConfigurations {
		log.Printf("\t%s:%s", *mc.Name, *mc.ID)
	}

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
		managedClustersName,
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
					ClientID: to.StringPtr(objectID),
					Secret:   to.StringPtr(clientSecret),
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

func createMaintenanceConfiguration(ctx context.Context, conn *arm.Connection) (*armcontainerservice.MaintenanceConfiguration, error) {
	maintenanceConfigurationsClient := armcontainerservice.NewMaintenanceConfigurationsClient(conn, subscriptionID)
	start, err := time.Parse("2006-01-02 15:04:05 06", "2021-09-25T13:00:00Z")
	end, err := time.Parse("2006-01-02 15:04:05 06", "2021-09-25T14:00:00Z")
	resp, err := maintenanceConfigurationsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		managedClustersName,
		configName,
		armcontainerservice.MaintenanceConfiguration{
			Properties: &armcontainerservice.MaintenanceConfigurationProperties{
				NotAllowedTime: []*armcontainerservice.TimeSpan{
					{
						Start: to.TimePtr(start),
						End:   to.TimePtr(end),
					},
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.MaintenanceConfiguration, nil
}

func listMaintenanceConfiguration(ctx context.Context, conn *arm.Connection) []*armcontainerservice.MaintenanceConfiguration {
	maintenanceConfigurationsClient := armcontainerservice.NewMaintenanceConfigurationsClient(conn, subscriptionID)

	maintenanceConfigurationPager := maintenanceConfigurationsClient.ListByManagedCluster(resourceGroupName, managedClustersName, nil)

	maintenanceConfigurations := make([]*armcontainerservice.MaintenanceConfiguration, 0)
	for maintenanceConfigurationPager.NextPage(ctx) {
		pageResp := maintenanceConfigurationPager.PageResponse()
		maintenanceConfigurations = append(maintenanceConfigurations, pageResp.Value...)
	}
	return maintenanceConfigurations
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
