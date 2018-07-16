// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package eventhubs

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	"github.com/Azure/go-autorest/autorest/to"
)

func getNamespacesClient() eventhub.NamespacesClient {
	nsClient := eventhub.NewNamespacesClient(helpers.SubscriptionID())
	auth, _ := iam.GetResourceManagementAuthorizer(iam.AuthGrantType())
	nsClient.Authorizer = auth
	nsClient.AddToUserAgent(helpers.UserAgent())
	return nsClient
}

// CreateNamespace creates an Event Hubs namespace
func CreateNamespace(ctx context.Context, nsName string) (*eventhub.EHNamespace, error) {
	nsClient := getNamespacesClient()
	future, err := nsClient.CreateOrUpdate(
		ctx,
		helpers.ResourceGroupName(),
		nsName,
		eventhub.EHNamespace{
			Location: to.StringPtr(helpers.Location()),
		},
	)
	if err != nil {
		return nil, err
	}

	err = future.WaitForCompletion(ctx, nsClient.Client)
	if err != nil {
		return nil, err
	}

	result, err := future.Result(nsClient)
	return &result, err
}
