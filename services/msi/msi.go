// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package msi

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/internal/iam"
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
	_ = msiClient.AddToUserAgent(config.UserAgent())
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
