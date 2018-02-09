package authorization

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
	"github.com/Azure/azure-sdk-for-go/services/authorization/mgmt/2015-07-01/authorization"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/satori/go.uuid"
)

func getRoleDefClient() (authorization.RoleDefinitionsClient, error) {
	token, _ := iam.GetResourceManagementToken(iam.AuthGrantType())
	roleDefClient := authorization.NewRoleDefinitionsClient(helpers.SubscriptionID())
	roleDefClient.Authorizer = autorest.NewBearerAuthorizer(token)
	roleDefClient.AddToUserAgent(helpers.UserAgent())
	return roleDefClient, nil
}

func getRoleClient() (authorization.RoleAssignmentsClient, error) {
	token, _ := iam.GetResourceManagementToken(iam.AuthGrantType())
	roleClient := authorization.NewRoleAssignmentsClient(helpers.SubscriptionID())
	roleClient.Authorizer = autorest.NewBearerAuthorizer(token)
	roleClient.AddToUserAgent(helpers.UserAgent())
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
