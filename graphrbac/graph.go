package graphrbac

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/marstr/randname"
)

func getServicePrincipalsClient() graphrbac.ServicePrincipalsClient {
	spClient := graphrbac.NewServicePrincipalsClient(config.TenantID())
	a, _ := iam.GetGraphAuthorizer()
	spClient.Authorizer = a
	spClient.AddToUserAgent(config.UserAgent())
	return spClient
}

func getApplicationsClient() graphrbac.ApplicationsClient {
	appClient := graphrbac.NewApplicationsClient(config.TenantID())
	a, _ := iam.GetGraphAuthorizer()
	appClient.Authorizer = a
	appClient.AddToUserAgent(config.UserAgent())
	return appClient
}

// getADGroupsClient retrieves a GroupsClient to assist with creating and managing Active Directory groups
func getADGroupsClient() graphrbac.GroupsClient {
	groupsClient := graphrbac.NewGroupsClient(config.TenantID())
	a, _ := iam.GetGraphAuthorizer()
	groupsClient.Authorizer = a
	groupsClient.AddToUserAgent(config.UserAgent())
	return groupsClient
}

// CreateServicePrincipal creates a service principal associated with the specified application.
func CreateServicePrincipal(ctx context.Context, appID string) (graphrbac.ServicePrincipal, error) {
	spClient := getServicePrincipalsClient()
	return spClient.Create(ctx,
		graphrbac.ServicePrincipalCreateParameters{
			AppID:          to.StringPtr(appID),
			AccountEnabled: to.BoolPtr(true),
		})
}

// CreateADApplication creates an Azure Active Directory (AAD) application
func CreateADApplication(ctx context.Context) (graphrbac.Application, error) {
	appClient := getApplicationsClient()
	return appClient.Create(ctx, graphrbac.ApplicationCreateParameters{
		AvailableToOtherTenants: to.BoolPtr(false),
		DisplayName:             to.StringPtr("Go SDK Samples"),
		Homepage:                to.StringPtr("https://azure.com"),
		IdentifierUris:          &[]string{randname.GenerateWithPrefix("https://gosdksamples", 10)},
	})
}

// DeleteADApplication deletes the specified AAD application
func DeleteADApplication(ctx context.Context, appObjID string) (autorest.Response, error) {
	appClient := getApplicationsClient()
	return appClient.Delete(ctx, appObjID)
}

// AddClientSecret adds a secret to the specified AAD app
func AddClientSecret(ctx context.Context, objID string) (autorest.Response, error) {
	appClient := getApplicationsClient()
	return appClient.UpdatePasswordCredentials(
		ctx,
		objID,
		graphrbac.PasswordCredentialsUpdateParameters{
			Value: &[]graphrbac.PasswordCredential{
				{
					StartDate: &date.Time{Time: time.Now()},
					EndDate:   &date.Time{Time: time.Date(2018, time.December, 20, 22, 0, 0, 0, time.UTC)},
					Value:     to.StringPtr("052265a2-bdc8-49aa-81bd-ecf7e9fe0c42"), // this will become the client secret! Record this value, there is no way to get it back
					KeyID:     to.StringPtr("08023993-9209-4580-9d4a-e060b44a64b8"),
				},
			},
		})
}

func getSignedInUserClient() graphrbac.SignedInUserClient {
	signedInUserClient := graphrbac.NewSignedInUserClient(config.TenantID())
	a, _ := iam.GetGraphAuthorizer()
	signedInUserClient.Authorizer = a
	signedInUserClient.AddToUserAgent(config.UserAgent())
	return signedInUserClient
}

// GetCurrentUser gets the Azure Active Directory object of the current signed in user
func GetCurrentUser(ctx context.Context) (graphrbac.User, error) {
	signedInUserClient := getSignedInUserClient()
	return signedInUserClient.Get(ctx)
}

// CreateADGroup creates an Active Directory group
func CreateADGroup(ctx context.Context) (graphrbac.ADGroup, error) {
	groupClient := getADGroupsClient()
	return groupClient.Create(ctx, graphrbac.GroupCreateParameters{
		DisplayName:     to.StringPtr("GoSDKSamples"),
		MailEnabled:     to.BoolPtr(false),
		MailNickname:    to.StringPtr("GoSDKMN"),
		SecurityEnabled: to.BoolPtr(true),
	})
}

// DeleteADGroup deletes the specified Active Directory group
func DeleteADGroup(ctx context.Context, groupObjID string) (autorest.Response, error) {
	groupClient := getADGroupsClient()
	return groupClient.Delete(ctx, groupObjID)
}

// GetServicePrincipalObjectID returns the service principal object ID for the specified client ID.
func GetServicePrincipalObjectID(ctx context.Context, clientID string) (string, error) {
	spClient := getServicePrincipalsClient()
	page, err := spClient.List(ctx, fmt.Sprintf("servicePrincipalNames/any(c:c eq '%s')", clientID))
	if err != nil {
		return "", err
	}
	servicePrincipals := page.Values()
	if len(servicePrincipals) == 0 {
		return "", fmt.Errorf("didn't find any service principals for client ID %s", clientID)
	}
	return *servicePrincipals[0].ObjectID, nil
}
