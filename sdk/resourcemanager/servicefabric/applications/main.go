// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicefabric/armservicefabric"
	"log"
	"os"
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
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createCluster(ctx context.Context, cred azcore.TokenCredential) (*armservicefabric.Cluster, error) {
	clustersClient, err := armservicefabric.NewClustersClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := clustersClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		clusterName,
		armservicefabric.Cluster{
			Location: to.Ptr(location),
			Properties: &armservicefabric.ClusterProperties{
				ManagementEndpoint: to.Ptr("https://myCluster.eastus.cloudapp.azure.com:19080"),
				NodeTypes: []*armservicefabric.NodeTypeDescription{
					{
						Name:                         to.Ptr("nt1vm"),
						ClientConnectionEndpointPort: to.Ptr[int32](19000),
						HTTPGatewayEndpointPort:      to.Ptr[int32](19007),
						ApplicationPorts: &armservicefabric.EndpointRangeDescription{
							StartPort: to.Ptr[int32](20000),
							EndPort:   to.Ptr[int32](30000),
						},
						EphemeralPorts: &armservicefabric.EndpointRangeDescription{
							StartPort: to.Ptr[int32](49000),
							EndPort:   to.Ptr[int32](64000),
						},
						IsPrimary:       to.Ptr(true),
						VMInstanceCount: to.Ptr[int32](5),
						DurabilityLevel: to.Ptr(armservicefabric.DurabilityLevelBronze),
					},
				},
				FabricSettings: []*armservicefabric.SettingsSectionDescription{
					{
						Name: to.Ptr("UpgradeService"),
						Parameters: []*armservicefabric.SettingsParameterDescription{
							{
								Name:  to.Ptr("AppPollIntervalInSeconds"),
								Value: to.Ptr("60"),
							},
						},
					},
				},
				DiagnosticsStorageAccountConfig: &armservicefabric.DiagnosticsStorageAccountConfig{
					StorageAccountName:      to.Ptr("diag"),
					ProtectedAccountKeyName: to.Ptr("StorageAccountKey1"),
					BlobEndpoint:            to.Ptr("https://diag.blob.core.windows.net/"),
					QueueEndpoint:           to.Ptr("https://diag.queue.core.windows.net/"),
					TableEndpoint:           to.Ptr("https://diag.table.core.windows.net/"),
				},
				ReliabilityLevel: to.Ptr(armservicefabric.ReliabilityLevelSilver),
				UpgradeMode:      to.Ptr(armservicefabric.UpgradeModeAutomatic),
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
	return &resp.Cluster, nil
}

func createApplication(ctx context.Context, cred azcore.TokenCredential) (*armservicefabric.ApplicationResource, error) {
	applicationsClient, err := armservicefabric.NewApplicationsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := applicationsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		clusterName,
		applicationName,
		armservicefabric.ApplicationResource{
			Location: to.Ptr(location),
			Properties: &armservicefabric.ApplicationResourceProperties{
				// minimum parameters
				TypeName:                  to.Ptr("myAppType"),
				TypeVersion:               to.Ptr("1.0"),
				RemoveApplicationCapacity: to.Ptr(false),
				Parameters: map[string]*string{
					"param1": to.Ptr("value1"),
				},
				// minimum parameters
				UpgradePolicy: &armservicefabric.ApplicationUpgradePolicy{
					ApplicationHealthPolicy: &armservicefabric.ArmApplicationHealthPolicy{
						ConsiderWarningAsError:                  to.Ptr(true),
						MaxPercentUnhealthyDeployedApplications: to.Ptr[int32](0),
						DefaultServiceTypeHealthPolicy: &armservicefabric.ArmServiceTypeHealthPolicy{
							MaxPercentUnhealthyServices:             to.Ptr[int32](0),
							MaxPercentUnhealthyPartitionsPerService: to.Ptr[int32](0),
							MaxPercentUnhealthyReplicasPerPartition: to.Ptr[int32](0),
						},
					},
					RollingUpgradeMonitoringPolicy: &armservicefabric.ArmRollingUpgradeMonitoringPolicy{
						FailureAction:             to.Ptr(armservicefabric.ArmUpgradeFailureActionRollback),
						HealthCheckRetryTimeout:   to.Ptr("00:10:00"),
						HealthCheckWaitDuration:   to.Ptr("00:02:00"),
						HealthCheckStableDuration: to.Ptr("00:05:00"),
						UpgradeDomainTimeout:      to.Ptr("1.06:00:00"),
						UpgradeTimeout:            to.Ptr("01:00:00"),
					},
					UpgradeReplicaSetCheckTimeout: to.Ptr("01:00:00"),
					ForceRestart:                  to.Ptr(false),
				},
				MaximumNodes: to.Ptr[int64](3),
				MinimumNodes: to.Ptr[int64](1),
				Metrics: []*armservicefabric.ApplicationMetricDescription{
					{
						Name:                     to.Ptr("metric1"),
						ReservationCapacity:      to.Ptr[int64](1),
						MaximumCapacity:          to.Ptr[int64](3),
						TotalApplicationCapacity: to.Ptr[int64](5),
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
	return &resp.ApplicationResource, nil
}

func getApplication(ctx context.Context, cred azcore.TokenCredential) (*armservicefabric.ApplicationResource, error) {
	applicationsClient, err := armservicefabric.NewApplicationsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := applicationsClient.Get(ctx, resourceGroupName, clusterName, applicationName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ApplicationResource, nil
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
