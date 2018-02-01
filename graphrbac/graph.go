package graphrbac

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

func getServicePrincipalClient() graphrbac.ServicePrincipalsClient {
	token, _ := iam.GetGraphToken(iam.AuthGrantType())
	spClient := graphrbac.NewServicePrincipalsClient(iam.TenantID())
	spClient.Authorizer = autorest.NewBearerAuthorizer(token)
	spClient.AddToUserAgent(helpers.UserAgent())
	return spClient
}

func getApplicationsClient() graphrbac.ApplicationsClient {
	token, _ := iam.GetGraphToken(iam.AuthGrantType())
	appClient := graphrbac.NewApplicationsClient(iam.TenantID())
	appClient.Authorizer = autorest.NewBearerAuthorizer(token)
	appClient.AddToUserAgent(helpers.UserAgent())
	return appClient
}

func CreateServicePrincipal(ctx context.Context, appID string) (graphrbac.ServicePrincipal, error) {
	spClient := getServicePrincipalClient()
	return spClient.Create(ctx, graphrbac.ServicePrincipalCreateParameters{
		AppID:          to.StringPtr(appID),
		AccountEnabled: to.BoolPtr(true),
	})
}

func CreateADApplication(ctx context.Context) (graphrbac.Application, error) {
	appClient := getApplicationsClient()
	return appClient.Create(ctx, graphrbac.ApplicationCreateParameters{
		AvailableToOtherTenants: to.BoolPtr(false),
		DisplayName:             to.StringPtr("go SDK samples"),
		Homepage:                to.StringPtr("http://gosdksamples"),
		IdentifierUris:          &[]string{"http://gosdksamples" + helpers.GetRandomLetterSequence(10)},
	})
}

func DeleteADApplication(ctx context.Context, appObjID string) (autorest.Response, error) {
	appClient := getApplicationsClient()
	return appClient.Delete(ctx, appObjID)
}
