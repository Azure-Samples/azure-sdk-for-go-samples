package graphrbac

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/Azure/go-autorest/autorest"
)

func getObjectsClient() graphrbac.ObjectsClient {
	token, err := iam.GetGraphToken(iam.AuthGrantType())
	if err != nil {
		helpers.PrintAndLog("err getting grph token")
		helpers.PrintAndLog(err.Error())
	}
	objClient := graphrbac.NewObjectsClient(iam.TenantID())
	objClient.Authorizer = autorest.NewBearerAuthorizer(token)
	objClient.AddToUserAgent(helpers.UserAgent())
	helpers.PrintAndLog("Created client")
	return objClient
}

// GetCurrentUser gets the Azure Active Directory object of the current user
func GetCurrentUser(ctx context.Context) (graphrbac.AADObject, error) {
	objClient := getObjectsClient()
	return objClient.GetCurrentUser(ctx)
}
