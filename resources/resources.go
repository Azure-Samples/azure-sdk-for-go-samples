// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package resources

import (
	"context"
	"net/http"
	"net/url"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/Azure/go-autorest/autorest"
)

func getResourcesClient() resources.Client {
	token, _ := iam.GetResourceManagementToken(iam.AuthGrantType())
	resourcesClient := resources.NewClient(helpers.SubscriptionID())
	resourcesClient.Authorizer = autorest.NewBearerAuthorizer(token)
	resourcesClient.AddToUserAgent(helpers.UserAgent())
	return resourcesClient
}

func GetResource(ctx context.Context, provider, resourceType, resourceName, apiVersion string) (resources.GenericResource, error) {
	resourcesClient := getResourcesClient()
	requestInspector := func(p autorest.Preparer) autorest.Preparer {
		return autorest.PreparerFunc(func(r *http.Request) (*http.Request, error) {
			v := r.URL.Query()
			d, err := url.QueryUnescape(apiVersion)
			if err != nil {
				return r, err
			}
			v.Set("api-version", d)
			r.URL.RawQuery = v.Encode()
			return r, nil
		})
	}
	resourcesClient.RequestInspector = requestInspector

	return resourcesClient.Get(
		ctx,
		helpers.ResourceGroupName(),
		provider,
		"",
		resourceType,
		resourceName,
	)
}
