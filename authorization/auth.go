package authorization

import (
	"context"
	"fmt"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
	"github.com/Azure/azure-sdk-for-go/services/authorization/mgmt/2015-07-01/authorization"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/satori/go.uuid"
)

func getRoleDefClient() (authorization.RoleDefinitionsClient, error) {
	token, _ := iam.GetResourceManagementToken(iam.AuthGrantType())
	roleDefClient := authorization.NewRoleDefinitionsClient(internal.SubscriptionID())
	roleDefClient.Authorizer = autorest.NewBearerAuthorizer(token)
	roleDefClient.AddToUserAgent(internal.UserAgent())
	return roleDefClient, nil
}

func getRoleClient() (authorization.RoleAssignmentsClient, error) {
	token, _ := iam.GetResourceManagementToken(iam.AuthGrantType())
	roleClient := authorization.NewRoleAssignmentsClient(internal.SubscriptionID())
	roleClient.Authorizer = autorest.NewBearerAuthorizer(token)
	roleClient.AddToUserAgent(internal.UserAgent())
	return roleClient, nil
}

// ListRoles gets the role definitions in the used resource group
func ListRoles(ctx context.Context, filter string) (list authorization.RoleDefinitionListResultPage, err error) {
	rg, err := resources.GetGroup(ctx)
	if err != nil {
		return
	}

	roleDefClient, _ := getRoleDefClient()
	return roleDefClient.List(ctx, *rg.ID, filter)
}

// AssignRole assigns a role, with a resource group scope
func AssignRole(ctx context.Context, principalID, roleDefID string) (role authorization.RoleAssignment, err error) {
	rg, err := resources.GetGroup(ctx)
	if err != nil {
		return
	}

	roleClient, _ := getRoleClient()
	return roleClient.Create(ctx, *rg.ID, uuid.NewV1().String(), authorization.RoleAssignmentCreateParameters{
		Properties: &authorization.RoleAssignmentProperties{
			PrincipalID:      to.StringPtr(principalID),
			RoleDefinitionID: to.StringPtr(roleDefID),
		},
	})
}

// AssignRoleWithSubscriptionScope assigns a role, with a subscription scope
func AssignRoleWithSubscriptionScope(ctx context.Context, principalID, roleDefID string) (role authorization.RoleAssignment, err error) {
	scope := fmt.Sprintf("/subscriptions/%s", internal.SubscriptionID())

	roleClient, _ := getRoleClient()
	return roleClient.Create(ctx, scope, uuid.NewV1().String(), authorization.RoleAssignmentCreateParameters{
		Properties: &authorization.RoleAssignmentProperties{
			PrincipalID:      to.StringPtr(principalID),
			RoleDefinitionID: to.StringPtr(roleDefID),
		},
	})
}

// DeleteRoleAssignment deletes a roleassignment
func DeleteRoleAssignment(ctx context.Context, id string) (authorization.RoleAssignment, error) {
	roleClient, _ := getRoleClient()
	return roleClient.DeleteByID(ctx, id)
}
