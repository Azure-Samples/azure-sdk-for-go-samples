// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package hdinsight

import (
	"context"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/internal/iam"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/internal/util"
	"github.com/Azure/azure-sdk-for-go/services/preview/hdinsight/mgmt/2015-03-01-preview/hdinsight"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/pkg/errors"
)

func getClustersClient() (*hdinsight.ClustersClient, error) {
	a, err := iam.GetResourceManagementAuthorizer()
	if err != nil {
		return nil, err
	}
	client := hdinsight.NewClustersClient(config.SubscriptionID())
	client.Authorizer = a
	client.AddToUserAgent(config.UserAgent())
	return &client, nil
}

// StorageAccountInfo describes the storage account used for the cluster's file system.
type StorageAccountInfo struct {
	Name      string
	Container string
	Key       string
}

// CreateHadoopCluster creats a simple hadoop 3.6 cluster
func CreateHadoopCluster(resourceGroup, clusterName string, info StorageAccountInfo) (*hdinsight.Cluster, error) {
	client, err := getClustersClient()
	if err != nil {
		return nil, err
	}
	// the default duration is 15 minutes which is just a tad too short
	client.PollingDuration = 20 * time.Minute
	util.PrintAndLog("creating hadoop cluster")
	future, err := client.Create(context.Background(), resourceGroup, clusterName, hdinsight.ClusterCreateParametersExtended{
		Location: to.StringPtr(config.Location()),
		Properties: &hdinsight.ClusterCreateProperties{
			ClusterVersion: to.StringPtr("3.6"),
			OsType:         hdinsight.Linux,
			Tier:           hdinsight.Standard,
			ClusterDefinition: &hdinsight.ClusterDefinition{
				Kind: to.StringPtr("hadoop"),
				Configurations: map[string]map[string]interface{}{
					"gateway": {
						"restAuthCredential.isEnabled": true,
						"restAuthCredential.username":  "admin",
						"restAuthCredential.password":  "Thisisalamepasswordthatwillberemoved2.",
					},
				},
			},
			ComputeProfile: &hdinsight.ComputeProfile{
				Roles: &[]hdinsight.Role{
					hdinsight.Role{
						Name:                to.StringPtr("headnode"),
						TargetInstanceCount: to.Int32Ptr(2),
						HardwareProfile: &hdinsight.HardwareProfile{
							VMSize: to.StringPtr("Large"),
						},
						OsProfile: &hdinsight.OsProfile{
							LinuxOperatingSystemProfile: &hdinsight.LinuxOperatingSystemProfile{
								Username: to.StringPtr("clusteruser"),
								Password: to.StringPtr("Thisisalamepasswordthatwillberemoved1."),
							},
						},
					},
					hdinsight.Role{
						Name:                to.StringPtr("workernode"),
						TargetInstanceCount: to.Int32Ptr(1),
						HardwareProfile: &hdinsight.HardwareProfile{
							VMSize: to.StringPtr("Large"),
						},
						OsProfile: &hdinsight.OsProfile{
							LinuxOperatingSystemProfile: &hdinsight.LinuxOperatingSystemProfile{
								Username: to.StringPtr("clusteruser"),
								Password: to.StringPtr("Thisisalamepasswordthatwillberemoved1."),
							},
						},
					},
					hdinsight.Role{
						Name:                to.StringPtr("zookeepernode"),
						TargetInstanceCount: to.Int32Ptr(3),
						HardwareProfile: &hdinsight.HardwareProfile{
							VMSize: to.StringPtr("Small"),
						},
						OsProfile: &hdinsight.OsProfile{
							LinuxOperatingSystemProfile: &hdinsight.LinuxOperatingSystemProfile{
								Username: to.StringPtr("clusteruser"),
								Password: to.StringPtr("Thisisalamepasswordthatwillberemoved1."),
							},
						},
					},
				},
			},
			StorageProfile: &hdinsight.StorageProfile{
				Storageaccounts: &[]hdinsight.StorageAccount{
					hdinsight.StorageAccount{
						Name:      &info.Name,
						Container: &info.Container,
						IsDefault: to.BoolPtr(true),
						Key:       &info.Key,
					},
				},
			},
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cluster")
	}
	util.PrintAndLog("waiting for hadoop cluster to finish deploying, this will take a while...")
	err = future.WaitForCompletionRef(context.Background(), client.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed waiting for cluster creation")
	}
	c, err := future.Result(*client)
	return &c, err
}
