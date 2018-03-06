package cdn

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/services/cdn/mgmt/2017-10-12/cdn"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

func getCDNClient() (cdn.BaseClient, error) {
	token, err := iam.GetResourceManagementToken(iam.AuthGrantType())
	if err != nil {
		return cdn.BaseClient{}, err
	}

	cdnClient := cdn.New(helpers.SubscriptionID())
	cdnClient.Authorizer = autorest.NewBearerAuthorizer(token)
	cdnClient.AddToUserAgent(helpers.UserAgent())
	return cdnClient, nil
}

// CheckNameAvailability use thes CDN package to determine whether or not a given name is appropriate.
func CheckNameAvailability(ctx context.Context, name, resourceType string) (bool, error) {
	client, err := getCDNClient()
	if err != nil {
		return false, err
	}

	resp, err := client.CheckNameAvailability(ctx, cdn.CheckNameAvailabilityInput{
		Name: to.StringPtr(name),
		Type: to.StringPtr(resourceType),
	})

	return *resp.NameAvailable, err
}
