// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package authorization

import (
	"context"
	"fmt"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/internal/iam"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/resources"
	"github.com/Azure/azure-sdk-for-go/services/authorization/mgmt/2015-07-01/authorization"
	"github.com/Azure/go-autorest/autorest/to"
	uuid "github.com/satori/go.uuid"
)

func getRoleDefinitionsClient() (authorization.RoleDefinitionsClient, error) {
	roleDefClient := authorization.NewRoleDefinitionsClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	roleDefClient.Authorizer = a
	_ = roleDefClient.AddToUserAgent(config.UserAgent())
	return roleDefClient, nil
}

func getRoleAssignmentsClient() (authorization.RoleAssignmentsClient, error) {
	roleClient := authorization.NewRoleAssignmentsClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	roleClient.Authorizer = a
	_ = roleClient.AddToUserAgent(config.UserAgent())
	return roleClient, nil
}

// ListRoles gets the role definitions in the used resource group
func ListRoleDefinitions(ctx context.Context, filter string) (list authorization.RoleDefinitionListResultPage, err error) {
	rg, err := resources.GetGroup(ctx)
	if err != nil {
		return
	}

	roleDefClient, _ := getRoleDefinitionsClient()
	return roleDefClient.List(ctx, *rg.ID, filter)
}

// AssignRole assigns a role to the named principal at the scope of the current group.
func AssignRole(ctx context.Context, principalID, roleDefID string) (role authorization.RoleAssignment, err error) {
	rg, err := resources.GetGroup(ctx)
	if err != nil {
		return
	}

	roleAssignmentsClient, _ := getRoleAssignmentsClient()
	return roleAssignmentsClient.Create(
		ctx,
		*rg.ID,
		uuid.NewV1().String(),
		authorization.RoleAssignmentCreateParameters{
			Properties: &authorization.RoleAssignmentProperties{
				PrincipalID:      to.StringPtr(principalID),
				RoleDefinitionID: to.StringPtr(roleDefID),
			},
		})
}

// AssignRoleWithSubscriptionScope assigns a role to the named principal at the
// subscription scope.
func AssignRoleWithSubscriptionScope(ctx context.Context, principalID, roleDefID string) (role authorization.RoleAssignment, err error) {
	scope := fmt.Sprintf("/subscriptions/%s", config.SubscriptionID())

	roleAssignmentsClient, _ := getRoleAssignmentsClient()
	return roleAssignmentsClient.Create(
		ctx,
		scope,
		uuid.NewV1().String(),
		authorization.RoleAssignmentCreateParameters{
			Properties: &authorization.RoleAssignmentProperties{
				PrincipalID:      to.StringPtr(principalID),
				RoleDefinitionID: to.StringPtr(roleDefID),
			},
		})
}

// DeleteRoleAssignment deletes a roleassignment
func DeleteRoleAssignment(ctx context.Context, id string) (authorization.RoleAssignment, error) {
	roleAssignmentsClient, _ := getRoleAssignmentsClient()
	return roleAssignmentsClient.DeleteByID(ctx, id)
}
