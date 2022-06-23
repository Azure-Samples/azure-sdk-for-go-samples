// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package resources

import (
	"context"
	"net/http"
	"net/url"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/internal/iam"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
	"github.com/Azure/go-autorest/autorest"
)

func getResourcesClient() resources.Client {
	resourcesClient := resources.NewClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	resourcesClient.Authorizer = a
	_ = resourcesClient.AddToUserAgent(config.UserAgent())
	return resourcesClient
}

// WithAPIVersion returns a prepare decorator that changes the request's query for api-version
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
		config.GroupName(),
		resourceProvider,
		"",
		resourceType,
		resourceName,
		apiVersion,
	)
}
