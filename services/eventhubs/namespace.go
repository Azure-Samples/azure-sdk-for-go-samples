// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package eventhubs

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/internal/iam"
	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	"github.com/Azure/go-autorest/autorest/to"
)

func getNamespacesClient() eventhub.NamespacesClient {
	nsClient := eventhub.NewNamespacesClient(config.SubscriptionID())
	auth, _ := iam.GetResourceManagementAuthorizer()
	nsClient.Authorizer = auth
	_ = nsClient.AddToUserAgent(config.UserAgent())
	return nsClient
}

// CreateNamespace creates an Event Hubs namespace
func CreateNamespace(ctx context.Context, nsName string) (*eventhub.EHNamespace, error) {
	nsClient := getNamespacesClient()
	future, err := nsClient.CreateOrUpdate(
		ctx,
		config.GroupName(),
		nsName,
		eventhub.EHNamespace{
			Location: to.StringPtr(config.Location()),
		},
	)
	if err != nil {
		return nil, err
	}

	err = future.WaitForCompletionRef(ctx, nsClient.Client)
	if err != nil {
		return nil, err
	}

	result, err := future.Result(nsClient)
	return &result, err
}
