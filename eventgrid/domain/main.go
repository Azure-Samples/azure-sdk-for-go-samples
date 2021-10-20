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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventgrid/armeventgrid"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	domainName        = "sample-domain"
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

	domain, err = getDomain(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get domain:", *domain.ID)

	keys, err := regenerateKey(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("regenerate key:", *keys.Key1, *keys.Key2)

	domains := listDomain(ctx, conn)
	for _, d := range domains {
		log.Println(*d.Name, *d.ID)
	}

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createDomain(ctx context.Context, conn *arm.Connection) (*armeventgrid.Domain, error) {
	domainsClient := armeventgrid.NewDomainsClient(conn, subscriptionID)

	pollerResp, err := domainsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		domainName,
		armeventgrid.Domain{
			TrackedResource: armeventgrid.TrackedResource{
				Location: to.StringPtr(location),
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

func getDomain(ctx context.Context, conn *arm.Connection) (*armeventgrid.Domain, error) {
	domainsClient := armeventgrid.NewDomainsClient(conn, subscriptionID)

	resp, err := domainsClient.Get(ctx, resourceGroupName, domainName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Domain, nil
}

func regenerateKey(ctx context.Context, conn *arm.Connection) (*armeventgrid.DomainSharedAccessKeys, error) {
	domainsClient := armeventgrid.NewDomainsClient(conn, subscriptionID)

	resp, err := domainsClient.RegenerateKey(
		ctx,
		resourceGroupName,
		domainName,
		armeventgrid.DomainRegenerateKeyRequest{
			KeyName: to.StringPtr("key1"),
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.DomainSharedAccessKeys, nil
}

func listDomain(ctx context.Context, conn *arm.Connection) []*armeventgrid.Domain {
	domainsClient := armeventgrid.NewDomainsClient(conn, subscriptionID)

	pager := domainsClient.ListByResourceGroup(resourceGroupName, nil)

	domains := make([]*armeventgrid.Domain, 0)
	for pager.NextPage(ctx) {
		resp := pager.PageResponse()
		domains = append(domains, resp.Value...)
	}
	return domains
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
