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
	resourcesClient := resources.NewClient(helpers.SubscriptionID())
	auth, _ := iam.GetResourceManagementAuthorizer(iam.AuthGrantType())
	resourcesClient.Authorizer = auth
	resourcesClient.AddToUserAgent(helpers.UserAgent())
	return resourcesClient
}

// WithAPIVersion returns a prepare decorator thta changes the request's query for api-version
// This can be set up as a client's RequestInspector.
func WithAPIVersion(apiVersion string) autorest.PrepareDecorator {
	return func(p autorest.Preparer) autorest.Preparer {
		return autorest.PreparerFunc(func(r *http.Request) (*http.Request, error) {
			r, err := p.Prepare(r)
			if err == nil {
				v := r.URL.Query()
				d, err := url.QueryUnescape(apiVersion)
				if err != nil {
					return r, err
				}
				v.Set("api-version", d)
				r.URL.RawQuery = v.Encode()
			}
			return r, err
		})
	}
}

// GetResource gets a resource, the generic way.
// The API version parameter overrides the API version in
// the SDK, this is needed because not all resources are
// supported on all API versions.
func GetResource(ctx context.Context, resourceProvider, resourceType, resourceName, apiVersion string) (resources.GenericResource, error) {
	resourcesClient := getResourcesClient()
	resourcesClient.RequestInspector = WithAPIVersion(apiVersion)

	return resourcesClient.Get(
		ctx,
		helpers.ResourceGroupName(),
		resourceProvider,
		"",
		resourceType,
		resourceName,
	)
}
