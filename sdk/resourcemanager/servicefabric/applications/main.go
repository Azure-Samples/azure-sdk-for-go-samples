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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicefabric/armservicefabric"
)

var (
	subscriptionID    string
	location          = "eastus"
	resourceGroupName = "sample-resource-group"
	clusterName       = "sample-servicefabric-cluster"
	applicationName   = "sample-servicefabric-application"
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

	application, err := createApplication(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service fabric application:", *application.ID)

	application, err = getApplication(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get service fabric application:", *application.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createCluster(ctx context.Context, cred azcore.TokenCredential) (*armservicefabric.Cluster, error) {
	clustersClient := armservicefabric.NewClustersClient(subscriptionID, cred, nil)
	pollerResp, err := clustersClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		clusterName,
		armservicefabric.Cluster{
			Location: to.StringPtr(location),
			Properties: &armservicefabric.ClusterProperties{
				ManagementEndpoint: to.StringPtr("https://myCluster.eastus.cloudapp.azure.com:19080"),
				NodeTypes: []*armservicefabric.NodeTypeDescription{
					{
						Name:                         to.StringPtr("nt1vm"),
						ClientConnectionEndpointPort: to.Int32Ptr(19000),
						HTTPGatewayEndpointPort:      to.Int32Ptr(19007),
						ApplicationPorts: &armservicefabric.EndpointRangeDescription{
							StartPort: to.Int32Ptr(20000),
							EndPort:   to.Int32Ptr(30000),
						},
						EphemeralPorts: &armservicefabric.EndpointRangeDescription{
							StartPort: to.Int32Ptr(49000),
							EndPort:   to.Int32Ptr(64000),
						},
						IsPrimary:       to.BoolPtr(true),
						VMInstanceCount: to.Int32Ptr(5),
						DurabilityLevel: armservicefabric.DurabilityLevelBronze.ToPtr(),
					},
				},
				FabricSettings: []*armservicefabric.SettingsSectionDescription{
					{
						Name: to.StringPtr("UpgradeService"),
						Parameters: []*armservicefabric.SettingsParameterDescription{
							{
								Name:  to.StringPtr("AppPollIntervalInSeconds"),
								Value: to.StringPtr("60"),
							},
						},
					},
				},
				DiagnosticsStorageAccountConfig: &armservicefabric.DiagnosticsStorageAccountConfig{
					StorageAccountName:      to.StringPtr("diag"),
					ProtectedAccountKeyName: to.StringPtr("StorageAccountKey1"),
					BlobEndpoint:            to.StringPtr("https://diag.blob.core.windows.net/"),
					QueueEndpoint:           to.StringPtr("https://diag.queue.core.windows.net/"),
					TableEndpoint:           to.StringPtr("https://diag.table.core.windows.net/"),
				},
				ReliabilityLevel: armservicefabric.ReliabilityLevelSilver.ToPtr(),
				UpgradeMode:      armservicefabric.UpgradeModeAutomatic.ToPtr(),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, 30*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.Cluster, nil
}

func createApplication(ctx context.Context, cred azcore.TokenCredential) (*armservicefabric.ApplicationResource, error) {
	applicationsClient := armservicefabric.NewApplicationsClient(subscriptionID, cred, nil)
	pollerResp, err := applicationsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		clusterName,
		applicationName,
		armservicefabric.ApplicationResource{
			Location: to.StringPtr(location),
			Properties: &armservicefabric.ApplicationResourceProperties{
				// minimum parameters
				TypeName:                  to.StringPtr("myAppType"),
				TypeVersion:               to.StringPtr("1.0"),
				RemoveApplicationCapacity: to.BoolPtr(false),
				Parameters: map[string]*string{
					"param1": to.StringPtr("value1"),
				},
				// minimum parameters
				UpgradePolicy: &armservicefabric.ApplicationUpgradePolicy{
					ApplicationHealthPolicy: &armservicefabric.ArmApplicationHealthPolicy{
						ConsiderWarningAsError:                  to.BoolPtr(true),
						MaxPercentUnhealthyDeployedApplications: to.Int32Ptr(0),
						DefaultServiceTypeHealthPolicy: &armservicefabric.ArmServiceTypeHealthPolicy{
							MaxPercentUnhealthyServices:             to.Int32Ptr(0),
							MaxPercentUnhealthyPartitionsPerService: to.Int32Ptr(0),
							MaxPercentUnhealthyReplicasPerPartition: to.Int32Ptr(0),
						},
					},
					RollingUpgradeMonitoringPolicy: &armservicefabric.ArmRollingUpgradeMonitoringPolicy{
						FailureAction:             armservicefabric.ArmUpgradeFailureActionRollback.ToPtr(),
						HealthCheckRetryTimeout:   to.StringPtr("00:10:00"),
						HealthCheckWaitDuration:   to.StringPtr("00:02:00"),
						HealthCheckStableDuration: to.StringPtr("00:05:00"),
						UpgradeDomainTimeout:      to.StringPtr("1.06:00:00"),
						UpgradeTimeout:            to.StringPtr("01:00:00"),
					},
					UpgradeReplicaSetCheckTimeout: to.StringPtr("01:00:00"),
					ForceRestart:                  to.BoolPtr(false),
				},
				MaximumNodes: to.Int64Ptr(3),
				MinimumNodes: to.Int64Ptr(1),
				Metrics: []*armservicefabric.ApplicationMetricDescription{
					{
						Name:                     to.StringPtr("metric1"),
						ReservationCapacity:      to.Int64Ptr(1),
						MaximumCapacity:          to.Int64Ptr(3),
						TotalApplicationCapacity: to.Int64Ptr(5),
					},
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, 30*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.ApplicationResource, nil
}

func getApplication(ctx context.Context, cred azcore.TokenCredential) (*armservicefabric.ApplicationResource, error) {
	applicationsClient := armservicefabric.NewApplicationsClient(subscriptionID, cred, nil)
	resp, err := applicationsClient.Get(ctx, resourceGroupName, clusterName, applicationName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ApplicationResource, nil
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
