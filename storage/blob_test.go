// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package storage

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-06-01/storage"
	"github.com/Azure/go-autorest/autorest/to"
)

func Example_blobSetServiceProperties() {
	// retrieves the current blob services settings and modifies them
	blobClient := getBlobClient()
	props, err := blobClient.GetServiceProperties(context.Background(), testAccountGroupName, testAccountName)
	if err != nil {
		util.LogAndPanic(err)
	}
	// enable blob versioning
	props.IsVersioningEnabled = to.BoolPtr(true)
	_, err = blobClient.SetServiceProperties(context.Background(), testAccountGroupName, testAccountName, props)
	if err != nil {
		util.LogAndPanic(err)
	}
}

func Example_blobObjectReplicationPolicy() {
	// create two object replication policies on a blob storage account.
	// each rule applies to separate source/destination containers.
	objRepClient := getObjRepClient()
	policy, err := objRepClient.CreateOrUpdate(context.Background(), testAccountGroupName, testAccountName, "default", storage.ObjectReplicationPolicy{
		ObjectReplicationPolicyProperties: &storage.ObjectReplicationPolicyProperties{
			SourceAccount:      to.StringPtr("source-account"),
			DestinationAccount: to.StringPtr("destination-account"),
			Rules: &[]storage.ObjectReplicationPolicyRule{
				{
					RuleID:               to.StringPtr("prefix-match-rule"),
					SourceContainer:      to.StringPtr("some source container"),
					DestinationContainer: to.StringPtr("some destination container"),
					Filters: &storage.ObjectReplicationPolicyFilter{
						// only replicate blobs with the prefix "foo"
						PrefixMatch: &[]string{"foo"},
					},
				},
				{
					RuleID:               to.StringPtr("creation-time-rule"),
					SourceContainer:      to.StringPtr("another source container"),
					DestinationContainer: to.StringPtr("another destination container"),
					Filters: &storage.ObjectReplicationPolicyFilter{
						// only replicate blobs created after this time
						MinCreationTime: to.StringPtr("2021-03-01T13:30:00Z"),
					},
				},
			},
		},
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	// display the ID of the policy that was created
	util.PrintAndLog(*policy.PolicyID)
}
