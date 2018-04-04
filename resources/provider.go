// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package resources

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
)

func getProviderClient() resources.ProvidersClient {
	providerClient := resources.NewProvidersClient(helpers.SubscriptionID())
	auth, _ := iam.GetResourceManagementAuthorizer(iam.AuthGrantType())
	providerClient.Authorizer = auth
	providerClient.AddToUserAgent(helpers.UserAgent())
	return providerClient
}

// RegisterProvider registers an azure resource provider for the subscription
func RegisterProvider(ctx context.Context, provider string) (resources.Provider, error) {
	providerClient := getProviderClient()
	return providerClient.Register(ctx, provider)
}
