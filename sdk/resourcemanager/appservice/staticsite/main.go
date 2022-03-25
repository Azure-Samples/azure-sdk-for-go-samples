// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	location          = "eastus2"
	resourceGroupName = "sample-resource-group"
	staticSiteName    = "sample-static-site"
)

// replace your repo information
var repoURL = "https://github.com/804873052/azure-rest-api-specs" // https://github.com/<github-name>/azure-rest-api-specs
var repoToken = "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"        // github token https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token
func main() {
	if repoToken == "" {
		log.Fatal("Please input repo information.")
	}

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

	staticSite, err := createStaticSite(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("static site:", *staticSite.ID)

	staticSite, err = getStaticSite(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get static site:", *staticSite.ID)

	listFunctions := listStaticSiteFunctions(ctx, cred)
	log.Println("list static site functions:", len(listFunctions))

	list := listStaticSite(ctx, cred)
	log.Println("list static site:", len(list))

	listCustimDomain := listStaticSiteCustomDomain(ctx, cred)
	log.Println("list static site custom domain:", len(listCustimDomain))

	reset, err := resetStaticSiteApiKey(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("reset static site api key:", reset.Status)

	detach, err := detachStaticSite(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("detach static site:", detach.Status)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createStaticSite(ctx context.Context, cred azcore.TokenCredential) (*armappservice.StaticSiteARMResource, error) {
	staticSitesClient := armappservice.NewStaticSitesClient(subscriptionID, cred, nil)

	pollerResp, err := staticSitesClient.BeginCreateOrUpdateStaticSite(
		ctx,
		resourceGroupName,
		staticSiteName,
		armappservice.StaticSiteARMResource{
			Location: to.StringPtr(location),
			SKU: &armappservice.SKUDescription{
				Name: to.StringPtr("Free"),
			},
			Properties: &armappservice.StaticSite{
				RepositoryURL:   to.StringPtr(repoURL),
				Branch:          to.StringPtr("master"),
				RepositoryToken: to.StringPtr(repoToken),
				BuildProperties: &armappservice.StaticSiteBuildProperties{
					AppLocation: to.StringPtr("app"),
					APILocation: to.StringPtr("api"),
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
	return &resp.StaticSiteARMResource, nil
}

func listStaticSiteFunctions(ctx context.Context, cred azcore.TokenCredential) []*armappservice.StaticSiteFunctionOverviewARMResource {
	staticSitesClient := armappservice.NewStaticSitesClient(subscriptionID, cred, nil)
	result := make([]*armappservice.StaticSiteFunctionOverviewARMResource, 0)
	listPager := staticSitesClient.ListStaticSiteFunctions(resourceGroupName, staticSiteName, nil)
	for listPager.NextPage(ctx) {
		resp := listPager.PageResponse()
		result = append(result, resp.Value...)
	}
	return result
}

func listStaticSite(ctx context.Context, cred azcore.TokenCredential) []*armappservice.StaticSiteARMResource {
	staticSitesClient := armappservice.NewStaticSitesClient(subscriptionID, cred, nil)
	result := make([]*armappservice.StaticSiteARMResource, 0)
	listPager := staticSitesClient.List(nil)
	for listPager.NextPage(ctx) {
		resp := listPager.PageResponse()
		result = append(result, resp.Value...)
	}
	return result
}

func listStaticSiteCustomDomain(ctx context.Context, cred azcore.TokenCredential) []*armappservice.StaticSiteCustomDomainOverviewARMResource {
	staticSitesClient := armappservice.NewStaticSitesClient(subscriptionID, cred, nil)
	result := make([]*armappservice.StaticSiteCustomDomainOverviewARMResource, 0)
	listPager := staticSitesClient.ListStaticSiteCustomDomains(resourceGroupName, staticSiteName, nil)
	for listPager.NextPage(ctx) {
		resp := listPager.PageResponse()
		result = append(result, resp.Value...)
	}
	return result
}

func getStaticSite(ctx context.Context, cred azcore.TokenCredential) (*armappservice.StaticSiteARMResource, error) {
	staticSitesClient := armappservice.NewStaticSitesClient(subscriptionID, cred, nil)
	resp, err := staticSitesClient.GetStaticSite(ctx, resourceGroupName, staticSiteName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.StaticSiteARMResource, nil
}

func resetStaticSiteApiKey(ctx context.Context, cred azcore.TokenCredential) (*http.Response, error) {
	staticSitesClient := armappservice.NewStaticSitesClient(subscriptionID, cred, nil)
	resp, err := staticSitesClient.ResetStaticSiteAPIKey(
		ctx,
		resourceGroupName,
		staticSiteName,
		armappservice.StaticSiteResetPropertiesARMResource{
			Properties: &armappservice.StaticSiteResetPropertiesARMResourceProperties{
				ShouldUpdateRepository: to.BoolPtr(true),
				RepositoryToken:        to.StringPtr(repoToken),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return resp.RawResponse, nil
}

func detachStaticSite(ctx context.Context, cred azcore.TokenCredential) (*http.Response, error) {
	staticSitesClient := armappservice.NewStaticSitesClient(subscriptionID, cred, nil)
	pollerResp, err := staticSitesClient.BeginDetachStaticSite(ctx, resourceGroupName, staticSiteName, nil)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp.RawResponse, nil
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
