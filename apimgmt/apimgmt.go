// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package apimgmt

import (
	"fmt"
	"strings"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
	api "github.com/Azure/azure-sdk-for-go/services/apimanagement/mgmt/2019-01-01/apimanagement"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

// returns a new instance of an API Svc client
func getAPISvcClient() api.ServiceClient {
	serviceClient := api.NewServiceClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	serviceClient.Authorizer = a
	serviceClient.AddToUserAgent(config.UserAgent())
	return serviceClient
}

// returns a valid instance of an API client
func getAPIClient() api.APIClient {
	apiClient := api.NewAPIClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	apiClient.Authorizer = a
	apiClient.AddToUserAgent(config.UserAgent())
	return apiClient
}

// CreateAPIMgmtSvc creates an instance of an API Management service
// wraps: https://godoc.org/github.com/Azure/azure-sdk-for-go/services/apimanagement/mgmt/2019-01-01/apimanagement#ServiceClient.CreateOrUpdate
func CreateAPIMgmtSvc(apimgmtsvc ServiceInfo) (service api.ServiceResource, err error) {
	serviceClient := getAPISvcClient()
	svcProp := api.ServiceProperties{
		PublisherEmail: &apimgmtsvc.Email,
		PublisherName:  &apimgmtsvc.Name,
	}
	sku := api.ServiceSkuProperties{
		Name: api.SkuTypeBasic,
	}
	future, err := serviceClient.CreateOrUpdate(
		apimgmtsvc.Ctx,
		apimgmtsvc.ResourceGroupName,
		apimgmtsvc.ServiceName,
		api.ServiceResource{
			Location:          to.StringPtr(config.Location()),
			ServiceProperties: &svcProp,
			Sku:               &sku,
		},
	)
	if err != nil {
		return service, fmt.Errorf("cannot create api management service: %v", err)
	}

	err = future.WaitForCompletionRef(apimgmtsvc.Ctx, serviceClient.Client)
	if err != nil {
		return service, fmt.Errorf("cannot get the api management service future response: %v", err)
	}

	return future.Result(serviceClient)
}

// CreateOrUpdateAPI creates an API endpoint on an API Management service
// wraps: https://godoc.org/github.com/Azure/azure-sdk-for-go/services/apimanagement/mgmt/2019-01-01/apimanagement#APIClient.CreateOrUpdate
func CreateOrUpdateAPI(apimgmtsvc ServiceInfo, properties api.APICreateOrUpdateParameter, apiid, ifMatch string) (apiContract api.APIContract, err error) {
	apiClient := getAPIClient()
	future, err := apiClient.CreateOrUpdate(
		apimgmtsvc.Ctx,
		apimgmtsvc.ResourceGroupName,
		apimgmtsvc.ServiceName,
		apiid,
		properties,
		ifMatch,
	)
	if err != nil {
		return apiContract, fmt.Errorf("cannot create api endpoint: %v", err)
	}

	err = future.WaitForCompletionRef(apimgmtsvc.Ctx, apiClient.Client)
	if err != nil {
		return apiContract, fmt.Errorf("cannot get the api endpoint future response: %v", err)
	}

	return future.Result(apiClient)
}

// DeleteAPI deletes an API endpoint on an API Management service
// wraps: https://godoc.org/github.com/Azure/azure-sdk-for-go/services/apimanagement/mgmt/2019-01-01/apimanagement#APIClient.Delete
func DeleteAPI(apimgmtsvc ServiceInfo, apiid, ifMatch string) (response autorest.Response, err error) {
	apiClient := getAPIClient()
	response, err = apiClient.Delete(
		apimgmtsvc.Ctx,
		apimgmtsvc.ResourceGroupName,
		apimgmtsvc.ServiceName,
		apiid,
		ifMatch,
		to.BoolPtr(true),
	)
	if err != nil {
		return response, fmt.Errorf("cannot delete api endpoint: %v", err)
	}

	return response, nil
}

// DeleteAPIMgmtSvc deletes an instance of an API Management service
// wraps: https://godoc.org/github.com/Azure/azure-sdk-for-go/services/apimanagement/mgmt/2019-01-01/apimanagement#ServiceClient.Delete
func DeleteAPIMgmtSvc(apimgmtsvc ServiceInfo) (service api.ServiceResource, err error) {
	serviceClient := getAPISvcClient()
	future, err := serviceClient.Delete(
		apimgmtsvc.Ctx,
		apimgmtsvc.ResourceGroupName,
		apimgmtsvc.ServiceName,
	)
	if err != nil {
		return service, fmt.Errorf("cannot delete api management service: %v", err)
	}

	err = future.WaitForCompletionRef(apimgmtsvc.Ctx, serviceClient.Client)
	if err != nil {
		return service, fmt.Errorf("cannot get the api management service future response: %v", err)
	}

	return future.Result(serviceClient)
}

// IsAPIMgmtSvcActivated check to see if the API Mgmt Svc has been activated, returns "true" if it has been activated.
func IsAPIMgmtSvcActivated(apimgmtsvc ServiceInfo) (activated bool, err error) {
	serviceClient := getAPISvcClient()

	resource, err := serviceClient.Get(
		apimgmtsvc.Ctx,
		apimgmtsvc.ResourceGroupName,
		apimgmtsvc.ServiceName,
	)
	if err != nil {
		return false, fmt.Errorf("cannot check the api management service: %v", err)
	}

	activated = false
	if strings.Compare(*resource.ServiceProperties.ProvisioningState, "Activating") != 0 {
		activated = true
	}

	return activated, err
}
