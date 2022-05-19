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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	serviceName       = "sample-api-service"
	userID            = "sampleuserid"
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

	apiManagementService, err := createApiManagementService(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("api management service:", *apiManagementService.ID)

	user, err := createUser(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("user:", *user.ID)

	entityTag, err := getEntityTag(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("entity tag:", *entityTag.ETag, entityTag.Success)

	sharedAccessToken, err := getSharedAccessToken(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("shared access token:", *sharedAccessToken.Value)

	generateSsoUrl, err := generateSsoURL(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("generate Sso URL:", *generateSsoUrl.Value)

	users, err := listUsers(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	for _, u := range users {
		log.Printf("user name:%s,user id:%s\n", *u.Name, *u.ID)
	}

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createApiManagementService(ctx context.Context, cred azcore.TokenCredential) (*armapimanagement.ServiceResource, error) {
	apiManagementServiceClient, err := armapimanagement.NewServiceClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := apiManagementServiceClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serviceName,
		armapimanagement.ServiceResource{
			Location: to.Ptr(location),
			Properties: &armapimanagement.ServiceProperties{
				PublisherName:  to.Ptr("sample"),
				PublisherEmail: to.Ptr("xxx@wircesoft.com"),
			},
			SKU: &armapimanagement.ServiceSKUProperties{
				Name:     to.Ptr(armapimanagement.SKUTypeStandard),
				Capacity: to.Ptr[int32](2),
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
	return &resp.ServiceResource, nil
}

func createUser(ctx context.Context, cred azcore.TokenCredential) (*armapimanagement.UserContract, error) {
	userClient, err := armapimanagement.NewUserClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := userClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		serviceName,
		userID,
		armapimanagement.UserCreateParameters{
			Properties: &armapimanagement.UserCreateParameterProperties{
				FirstName: to.Ptr("foo"),
				LastName:  to.Ptr("bar"),
				Email:     to.Ptr("foobar@wricesoft.com"),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.UserContract, nil
}

func getEntityTag(ctx context.Context, cred azcore.TokenCredential) (*armapimanagement.UserClientGetEntityTagResponse, error) {
	userClient, err := armapimanagement.NewUserClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := userClient.GetEntityTag(ctx, resourceGroupName, serviceName, userID, nil)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func getSharedAccessToken(ctx context.Context, cred azcore.TokenCredential) (*armapimanagement.UserTokenResult, error) {
	userClient, err := armapimanagement.NewUserClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := userClient.GetSharedAccessToken(
		ctx,
		resourceGroupName,
		serviceName, userID,
		armapimanagement.UserTokenParameters{
			Properties: &armapimanagement.UserTokenParameterProperties{
				Expiry:  to.Ptr(time.Now().AddDate(0, 0, 7)),
				KeyType: to.Ptr(armapimanagement.KeyTypePrimary),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.UserTokenResult, nil
}

func generateSsoURL(ctx context.Context, cred azcore.TokenCredential) (*armapimanagement.GenerateSsoURLResult, error) {
	userClient, err := armapimanagement.NewUserClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := userClient.GenerateSsoURL(ctx, resourceGroupName, serviceName, userID, nil)
	if err != nil {
		return nil, err
	}
	return &resp.GenerateSsoURLResult, nil
}

func listUsers(ctx context.Context, cred azcore.TokenCredential) ([]*armapimanagement.UserContract, error) {
	userClient, err := armapimanagement.NewUserClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pager := userClient.NewListByServicePager(resourceGroupName, serviceName, nil)

	users := make([]*armapimanagement.UserContract, 0)
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		users = append(users, resp.Value...)
	}

	return users, nil
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

	_, err = pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}
	return nil
}
