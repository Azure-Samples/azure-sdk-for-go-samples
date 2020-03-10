// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package msi

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
	"github.com/Azure/azure-sdk-for-go/services/msi/mgmt/2018-11-30/msi"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/pkg/errors"
)

func getMSIUserAssignedIDClient() (*msi.UserAssignedIdentitiesClient, error) {
	a, err := iam.GetResourceManagementAuthorizer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get authorizer")
	}
	msiClient := msi.NewUserAssignedIdentitiesClient(config.SubscriptionID())
	msiClient.Authorizer = a
	msiClient.AddToUserAgent(config.UserAgent())
	return &msiClient, nil
}

// CreateUserAssignedIdentity creates a user-assigned identity in the specified resource group.
func CreateUserAssignedIdentity(resourceGroup, identity string) (*msi.Identity, error) {
	msiClient, err := getMSIUserAssignedIDClient()
	if err != nil {
		return nil, err
	}
	id, err := msiClient.CreateOrUpdate(context.Background(), resourceGroup, identity, msi.Identity{
		Location: to.StringPtr(config.Location()),
	})
	return &id, err
}
