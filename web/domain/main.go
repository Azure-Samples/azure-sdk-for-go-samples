package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/web/armweb"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	domainName        = "go-sample.co.uk"
)

func main() {
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set.")
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}

	conn := arm.NewDefaultConnection(cred, &arm.ConnectionOptions{
		Logging: policy.LogOptions{
			IncludeBody: true,
		},
	})
	ctx := context.Background()

	resourceGroup, err := createResourceGroup(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	domain, err := createDomain(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("domain:", *domain.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createDomain(ctx context.Context, conn *arm.Connection) (*armweb.Domain, error) {
	domainsClient := armweb.NewDomainsClient(conn, subscriptionID)

	pollerResp, err := domainsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		domainName,
		armweb.Domain{
			Resource: armweb.Resource{
				Location: to.StringPtr(location),
			},
			Properties: &armweb.DomainProperties{
				Consent: &armweb.DomainPurchaseConsent{},
				ContactAdmin: &armweb.Contact{
					NameFirst: to.StringPtr("sample"),
					NameLast:  to.StringPtr("admin"),
					Email:     to.StringPtr("xxx@wricesoft.com"),
					Phone:     to.StringPtr("12333333333"),
				},
				ContactBilling: &armweb.Contact{
					NameFirst: to.StringPtr("sample"),
					NameLast:  to.StringPtr("billing"),
					Email:     to.StringPtr("yyy@wricesoft.com"),
					Phone:     to.StringPtr("123333333334"),
				},
				ContactRegistrant: &armweb.Contact{
					NameFirst: to.StringPtr("sample"),
					NameLast:  to.StringPtr("registrant"),
					Email:     to.StringPtr("yyy@wricesoft.com"),
					Phone:     to.StringPtr("123333333334"),
				},
				ContactTech: &armweb.Contact{
					NameFirst: to.StringPtr("sample"),
					NameLast:  to.StringPtr("tech"),
					Email:     to.StringPtr("yyy@wricesoft.com"),
					Phone:     to.StringPtr("123333333334"),
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.Domain, nil
}

func createResourceGroup(ctx context.Context, conn *arm.Connection) (*armresources.ResourceGroup, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn, subscriptionID)

	resourceGroupResp, err := resourceGroupClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		armresources.ResourceGroup{
			Location: to.StringPtr(location),
		},
		nil)
	if err != nil {
		return nil, err
	}
	return &resourceGroupResp.ResourceGroup, nil
}

func cleanup(ctx context.Context, conn *arm.Connection) (*http.Response, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn, subscriptionID)

	pollerResp, err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp.RawResponse, nil
}
