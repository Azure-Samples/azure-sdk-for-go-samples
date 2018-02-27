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
)

func getGroupsClient() resources.GroupsClient {
	token, err := iam.GetResourceManagementTokenHybrid(helpers.ActiveDirectoryEndpoint(), helpers.TenantID(), helpers.ClientID(), helpers.ClientSecret(), helpers.ActiveDirectoryResourceID())
	if err != nil {
		log.Fatal(fmt.Sprintf("Cannot generate token. Error details: %s.", err.Error()))
	}
	groupsClient := resources.NewGroupsClientWithBaseURI(helpers.ArmEndpoint(), helpers.SubscriptionID())
	groupsClient.Authorizer = autorest.NewBearerAuthorizer(token)

	return groupsClient
}

// CreateGroup creates a new resource group named by env var
func CreateGroup(cntx context.Context, groupName string) (resources.Group, error) {
	err := helpers.ParseArgs()
	if err != nil {
		log.Fatal(err.Error())
	}
	groupClient := getGroupsClient()
	location := "local"
	return groupClient.CreateOrUpdate(cntx, groupName, resources.Group{Location: &location})
}

// DeleteGroup removes the resource group named by env var
func DeleteGroup(ctx context.Context, groupName string) (result resources.GroupsDeleteFuture, err error) {
	groupsClient := getGroupsClient()
	return groupsClient.Delete(ctx, groupName)
}
