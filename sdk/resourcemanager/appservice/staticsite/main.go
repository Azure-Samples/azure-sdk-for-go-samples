// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"log"
	"os"
)

var (
	subscriptionID    string
	location          = "eastus2"
	resourceGroupName = "sample-resource-group"
	staticSiteName    = "sample-static-site"
)

var (
	resourcesClientFactory  *armresources.ClientFactory
	appserviceClientFactory *armappservice.ClientFactory
)

var (
	resourceGroupClient *armresources.ResourceGroupsClient
	staticSitesClient   *armappservice.StaticSitesClient
)

// replace your repo information
var repoURL = "" // https://github.com/<github-name>/azure-rest-api-specs
var repoToken = "" // github token https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token
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

	resourcesClientFactory, err = armresources.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	resourceGroupClient = resourcesClientFactory.NewResourceGroupsClient()

	appserviceClientFactory, err = armappservice.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	staticSitesClient = appserviceClientFactory.NewStaticSitesClient()

	resourceGroup, err := createResourceGroup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	staticSite, err := createStaticSite(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("static site:", *staticSite.ID)

	staticSite, err = getStaticSite(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get static site:", *staticSite.ID)

	listFunctions, err := listStaticSiteFunctions(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("list static site functions:", len(listFunctions))

	list, err := listStaticSite(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("list static site:", len(list))

	listCustimDomain, err := listStaticSiteCustomDomain(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("list static site custom domain:", len(listCustimDomain))

	err = resetStaticSiteApiKey(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("reset static site api key")

	err = detachStaticSite(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("detached static site")

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createStaticSite(ctx context.Context) (*armappservice.StaticSiteARMResource, error) {

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
	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.StaticSiteARMResource, nil
}

func listStaticSiteFunctions(ctx context.Context) ([]*armappservice.StaticSiteFunctionOverviewARMResource, error) {

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

func listStaticSite(ctx context.Context) ([]*armappservice.StaticSiteARMResource, error) {

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

func listStaticSiteCustomDomain(ctx context.Context) ([]*armappservice.StaticSiteCustomDomainOverviewARMResource, error) {

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

func getStaticSite(ctx context.Context) (*armappservice.StaticSiteARMResource, error) {

	resp, err := staticSitesClient.GetStaticSite(ctx, resourceGroupName, staticSiteName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.StaticSiteARMResource, nil
}

func resetStaticSiteApiKey(ctx context.Context) error {

	_, err := staticSitesClient.ResetStaticSiteAPIKey(
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

func detachStaticSite(ctx context.Context) error {

	pollerResp, err := staticSitesClient.BeginDetachStaticSite(ctx, resourceGroupName, staticSiteName, nil)
	if err != nil {
		return err
	}
	_, err = pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}
	return nil
}

func createResourceGroup(ctx context.Context) (*armresources.ResourceGroup, error) {

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

func cleanup(ctx context.Context) error {

	pollerResp, err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return err
	}

	_, err = pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}
	return nil
}
