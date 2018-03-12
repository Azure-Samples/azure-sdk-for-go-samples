package graphrbac

import (
	"context"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/Azure/go-autorest/autorest/to"
)

func getServicePrincipalClient() graphrbac.ServicePrincipalsClient {
	token, _ := iam.GetGraphToken(iam.AuthGrantType())
	spClient := graphrbac.NewServicePrincipalsClient(iam.TenantID())
	spClient.Authorizer = autorest.NewBearerAuthorizer(token)
	spClient.AddToUserAgent(internal.UserAgent())
	return spClient
}

func getApplicationsClient() graphrbac.ApplicationsClient {
	token, _ := iam.GetGraphToken(iam.AuthGrantType())
	appClient := graphrbac.NewApplicationsClient(iam.TenantID())
	appClient.Authorizer = autorest.NewBearerAuthorizer(token)
	appClient.AddToUserAgent(internal.UserAgent())
	return appClient
}

// CreateServicePrincipal creates a service principal on the specified Azure Active Directory application.
func CreateServicePrincipal(ctx context.Context, appID string) (graphrbac.ServicePrincipal, error) {
	spClient := getServicePrincipalClient()
	return spClient.Create(ctx, graphrbac.ServicePrincipalCreateParameters{
		AppID:          to.StringPtr(appID),
		AccountEnabled: to.BoolPtr(true),
	})
}

// CreateADApplication creates an Azure Active Directory (AAD) application
func CreateADApplication(ctx context.Context) (graphrbac.Application, error) {
	appClient := getApplicationsClient()
	return appClient.Create(ctx, graphrbac.ApplicationCreateParameters{
		AvailableToOtherTenants: to.BoolPtr(false),
		DisplayName:             to.StringPtr("go SDK samples"),
		Homepage:                to.StringPtr("http://gosdksamples"),
		IdentifierUris:          &[]string{"http://gosdksamples" + internal.GetRandomLetterSequence(10)},
	})
}

// DeleteADApplication deletes the specified AAD application
func DeleteADApplication(ctx context.Context, appObjID string) (autorest.Response, error) {
	appClient := getApplicationsClient()
	return appClient.Delete(ctx, appObjID)
}

// AddClientSecret adds a client secret (aka password credential) to the specified AAD app
func AddClientSecret(ctx context.Context, objID string) (autorest.Response, error) {
	appClient := getApplicationsClient()
	return appClient.UpdatePasswordCredentials(ctx, objID, graphrbac.PasswordCredentialsUpdateParameters{
		Value: &[]graphrbac.PasswordCredential{
			{
				StartDate: &date.Time{time.Now()},
				EndDate:   &date.Time{time.Date(2018, time.December, 20, 22, 0, 0, 0, time.UTC)},
				Value:     to.StringPtr("052265a2-bdc8-49aa-81bd-ecf7e9fe0c42"), // this will become the client secret! Record this value, there is no way to get it back
				KeyID:     to.StringPtr("08023993-9209-4580-9d4a-e060b44a64b8"),
			},
		},
	})
}

func getObjectsClient() graphrbac.ObjectsClient {
	token, _ := iam.GetGraphToken(iam.AuthGrantType())
	objClient := graphrbac.NewObjectsClient(iam.TenantID())
	objClient.Authorizer = autorest.NewBearerAuthorizer(token)
	objClient.AddToUserAgent(internal.UserAgent())
	return objClient
}

// GetCurrentUser gets the Azure Active Directory object of the current user
func GetCurrentUser(ctx context.Context) (graphrbac.AADObject, error) {
	objClient := getObjectsClient()
	return objClient.GetCurrentUser(ctx)
}
