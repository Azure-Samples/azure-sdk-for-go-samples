package cdn

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/services/cdn/mgmt/2017-10-12/cdn"
	"github.com/Azure/go-autorest/autorest/to"
)

func getCDNClient() cdn.BaseClient {
	cdnClient := cdn.New(helpers.SubscriptionID())
	auth, _ := iam.GetResourceManagementAuthorizer(iam.AuthGrantType())
	cdnClient.Authorizer = auth
	cdnClient.AddToUserAgent(helpers.UserAgent())
	return cdnClient
}

// CheckNameAvailability use thes CDN package to determine whether or not a given name is appropriate.
func CheckNameAvailability(ctx context.Context, name, resourceType string) (bool, error) {
	client := getCDNClient()
	resp, err := client.CheckNameAvailability(ctx, cdn.CheckNameAvailabilityInput{
		Name: to.StringPtr(name),
		Type: to.StringPtr(resourceType),
	})

	return *resp.NameAvailable, err
}
