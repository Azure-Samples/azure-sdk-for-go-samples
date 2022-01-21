package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
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
	ctx := context.Background()

	resourceGroup, err := createResourceGroup(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	domain, err := createDomain(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("domain:", *domain.ID)

	domain, err = getDomain(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get domain:", *domain.ID)

	keys, err := regenerateKey(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("regenerate key:", *keys.Key1, *keys.Key2)

	domains := listDomain(ctx, cred)
	for _, d := range domains {
		log.Println(*d.Name, *d.ID)
	}

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createDomain(ctx context.Context, cred azcore.TokenCredential) (*armeventgrid.Domain, error) {
	domainsClient := armeventgrid.NewDomainsClient(subscriptionID, cred, nil)

	pollerResp, err := domainsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		domainName,
		armeventgrid.Domain{
			Location: to.StringPtr(location),
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

func getDomain(ctx context.Context, cred azcore.TokenCredential) (*armeventgrid.Domain, error) {
	domainsClient := armeventgrid.NewDomainsClient(subscriptionID, cred, nil)

	resp, err := domainsClient.Get(ctx, resourceGroupName, domainName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Domain, nil
}

func regenerateKey(ctx context.Context, cred azcore.TokenCredential) (*armeventgrid.DomainSharedAccessKeys, error) {
	domainsClient := armeventgrid.NewDomainsClient(subscriptionID, cred, nil)

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

func listDomain(ctx context.Context, cred azcore.TokenCredential) []*armeventgrid.Domain {
	domainsClient := armeventgrid.NewDomainsClient(subscriptionID, cred, nil)

	pager := domainsClient.ListByResourceGroup(resourceGroupName, nil)

	domains := make([]*armeventgrid.Domain, 0)
	for pager.NextPage(ctx) {
		resp := pager.PageResponse()
		domains = append(domains, resp.Value...)
	}
	return domains
}

func createResourceGroup(ctx context.Context, cred azcore.TokenCredential) (*armresources.ResourceGroup, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)

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

func cleanup(ctx context.Context, cred azcore.TokenCredential) (*http.Response, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)

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
