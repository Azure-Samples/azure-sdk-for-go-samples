// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"log"
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

	listFunctions, err := listStaticSiteFunctions(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("list static site functions:", len(listFunctions))

	list, err := listStaticSite(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("list static site:", len(list))

	listCustimDomain, err := listStaticSiteCustomDomain(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("list static site custom domain:", len(listCustimDomain))

	err = resetStaticSiteApiKey(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("reset static site api key")

	err = detachStaticSite(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("detached static site")

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createStaticSite(ctx context.Context, cred azcore.TokenCredential) (*armappservice.StaticSiteARMResource, error) {
	staticSitesClient, err := armappservice.NewStaticSitesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := staticSitesClient.BeginCreateOrUpdateStaticSite(
		ctx,
		resourceGroupName,
		staticSiteName,
		armappservice.StaticSiteARMResource{
			Location: to.Ptr(location),
			SKU: &armappservice.SKUDescription{
				Name: to.Ptr("Free"),
			},
			Properties: &armappservice.StaticSite{
				RepositoryURL:   to.Ptr(repoURL),
				Branch:          to.Ptr("master"),
				RepositoryToken: to.Ptr(repoToken),
				BuildProperties: &armappservice.StaticSiteBuildProperties{
					AppLocation: to.Ptr("app"),
					APILocation: to.Ptr("api"),
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

func listStaticSiteFunctions(ctx context.Context, cred azcore.TokenCredential) ([]*armappservice.StaticSiteFunctionOverviewARMResource, error) {
	staticSitesClient, err := armappservice.NewStaticSitesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	result := make([]*armappservice.StaticSiteFunctionOverviewARMResource, 0)
	listPager := staticSitesClient.NewListStaticSiteFunctionsPager(resourceGroupName, staticSiteName, nil)
	for listPager.More() {
		resp, err := listPager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		result = append(result, resp.Value...)
	}
	return result, nil
}

func listStaticSite(ctx context.Context, cred azcore.TokenCredential) ([]*armappservice.StaticSiteARMResource, error) {
	staticSitesClient, err := armappservice.NewStaticSitesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	result := make([]*armappservice.StaticSiteARMResource, 0)
	listPager := staticSitesClient.NewListPager(nil)
	for listPager.More() {
		resp, err := listPager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		result = append(result, resp.Value...)
	}
	return result, nil
}

func listStaticSiteCustomDomain(ctx context.Context, cred azcore.TokenCredential) ([]*armappservice.StaticSiteCustomDomainOverviewARMResource, error) {
	staticSitesClient, err := armappservice.NewStaticSitesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	result := make([]*armappservice.StaticSiteCustomDomainOverviewARMResource, 0)
	listPager := staticSitesClient.NewListStaticSiteCustomDomainsPager(resourceGroupName, staticSiteName, nil)
	for listPager.More() {
		resp, err := listPager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		result = append(result, resp.Value...)
	}
	return result, nil
}

func getStaticSite(ctx context.Context, cred azcore.TokenCredential) (*armappservice.StaticSiteARMResource, error) {
	staticSitesClient, err := armappservice.NewStaticSitesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	resp, err := staticSitesClient.GetStaticSite(ctx, resourceGroupName, staticSiteName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.StaticSiteARMResource, nil
}

func resetStaticSiteApiKey(ctx context.Context, cred azcore.TokenCredential) error {
	staticSitesClient, err := armappservice.NewStaticSitesClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}
	_, err = staticSitesClient.ResetStaticSiteAPIKey(
		ctx,
		resourceGroupName,
		staticSiteName,
		armappservice.StaticSiteResetPropertiesARMResource{
			Properties: &armappservice.StaticSiteResetPropertiesARMResourceProperties{
				ShouldUpdateRepository: to.Ptr(true),
				RepositoryToken:        to.Ptr(repoToken),
			},
		},
		nil,
	)
	if err != nil {
		return err
	}
	return nil
}

func detachStaticSite(ctx context.Context, cred azcore.TokenCredential) error {
	staticSitesClient, err := armappservice.NewStaticSitesClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}
	pollerResp, err := staticSitesClient.BeginDetachStaticSite(ctx, resourceGroupName, staticSiteName, nil)
	if err != nil {
		return err
	}
	_, err = pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return err
	}
	return nil
}

func createResourceGroup(ctx context.Context, cred azcore.TokenCredential) (*armresources.ResourceGroup, error) {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resourceGroupResp, err := resourceGroupClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		armresources.ResourceGroup{
			Location: to.Ptr(location),
		},
		nil)
	if err != nil {
		return nil, err
	}
	return &resourceGroupResp.ResourceGroup, nil
}

func cleanup(ctx context.Context, cred azcore.TokenCredential) error {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	pollerResp, err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return err
	}

	_, err = pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return err
	}
	return nil
}
