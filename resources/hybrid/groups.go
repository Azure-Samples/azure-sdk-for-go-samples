// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package hybridresources

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/resources/mgmt/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
)

const (
	errorPrefix = "Cannot create resource group, reason: %v"
)

func getGroupsClient(activeDirectoryEndpoint, tokenAudience string) resources.GroupsClient {
	token, err := iam.GetResourceManagementTokenHybrid(activeDirectoryEndpoint, tokenAudience, helpers.TenantID(), helpers.ClientID(), helpers.ClientSecret())
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}
	groupsClient := resources.NewGroupsClientWithBaseURI(helpers.ArmEndpoint(), helpers.SubscriptionID())
	groupsClient.Authorizer = autorest.NewBearerAuthorizer(token)

	return groupsClient
}

// CreateGroup creates a new resource group named by env var
func CreateGroup(cntx context.Context) (resources.Group, error) {
	environment, _ := azure.EnvironmentFromURL(helpers.ArmEndpoint())
	groupClient := getGroupsClient(environment.ActiveDirectoryEndpoint, environment.TokenAudience)
	location := helpers.Location()
	helpers.SetResourceGroupName("hybridResourceGroup")
	return groupClient.CreateOrUpdate(cntx, helpers.ResourceGroupName(), resources.Group{Location: &location})
}

// DeleteGroup removes the resource group named by env var
func DeleteGroup(ctx context.Context) (result resources.GroupsDeleteFuture, err error) {
	environment, _ := azure.EnvironmentFromURL(helpers.ArmEndpoint())
	groupsClient := getGroupsClient(environment.ActiveDirectoryEndpoint, environment.TokenAudience)
	helpers.SetResourceGroupName("hybridResourceGroup")
	return groupsClient.Delete(ctx, helpers.ResourceGroupName())
}
