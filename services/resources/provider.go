// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package resources

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/internal/iam"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
)

func getProviderClient() resources.ProvidersClient {
	providerClient := resources.NewProvidersClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	providerClient.Authorizer = a
	providerClient.AddToUserAgent(config.UserAgent())
	return providerClient
}

// RegisterProvider registers an azure resource provider for the subscription
func RegisterProvider(ctx context.Context, provider string) (resources.Provider, error) {
	providerClient := getProviderClient()
	return providerClient.Register(ctx, provider)
}
