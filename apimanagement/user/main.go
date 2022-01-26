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

	users := listUsers(ctx, cred)
	for _, u := range users {
		log.Printf("user name:%s,user id:%s\n", *u.Name, *u.ID)
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

func createApiManagementService(ctx context.Context, cred azcore.TokenCredential) (*armapimanagement.ServiceResource, error) {
	apiManagementServiceClient := armapimanagement.NewServiceClient(subscriptionID, cred, nil)

	pollerResp, err := apiManagementServiceClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serviceName,
		armapimanagement.ServiceResource{
			Location: to.StringPtr(location),
			Properties: &armapimanagement.ServiceProperties{
				PublisherName:  to.StringPtr("sample"),
				PublisherEmail: to.StringPtr("xxx@wircesoft.com"),
			},
			SKU: &armapimanagement.ServiceSKUProperties{
				Name:     armapimanagement.SKUTypeStandard.ToPtr(),
				Capacity: to.Int32Ptr(2),
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
	return &resp.ServiceResource, nil
}

func createUser(ctx context.Context, cred azcore.TokenCredential) (*armapimanagement.UserContract, error) {
	userClient := armapimanagement.NewUserClient(subscriptionID, cred, nil)

	resp, err := userClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		serviceName,
		userID,
		armapimanagement.UserCreateParameters{
			Properties: &armapimanagement.UserCreateParameterProperties{
				FirstName: to.StringPtr("foo"),
				LastName:  to.StringPtr("bar"),
				Email:     to.StringPtr("foobar@wricesoft.com"),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.UserContract, nil
}

func getEntityTag(ctx context.Context, cred azcore.TokenCredential) (*armapimanagement.UserClientGetEntityTagResult, error) {
	userClient := armapimanagement.NewUserClient(subscriptionID, cred, nil)

	resp, err := userClient.GetEntityTag(ctx, resourceGroupName, serviceName, userID, nil)
	if err != nil {
		return nil, err
	}
	return &resp.UserClientGetEntityTagResult, nil
}

func getSharedAccessToken(ctx context.Context, cred azcore.TokenCredential) (*armapimanagement.UserTokenResult, error) {
	userClient := armapimanagement.NewUserClient(subscriptionID, cred, nil)

	resp, err := userClient.GetSharedAccessToken(
		ctx,
		resourceGroupName,
		serviceName, userID,
		armapimanagement.UserTokenParameters{
			Properties: &armapimanagement.UserTokenParameterProperties{
				Expiry:  to.TimePtr(time.Now().AddDate(0, 0, 7)),
				KeyType: armapimanagement.KeyTypePrimary.ToPtr(),
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
	userClient := armapimanagement.NewUserClient(subscriptionID, cred, nil)

	resp, err := userClient.GenerateSsoURL(ctx, resourceGroupName, serviceName, userID, nil)
	if err != nil {
		return nil, err
	}
	return &resp.GenerateSsoURLResult, nil
}

func listUsers(ctx context.Context, cred azcore.TokenCredential) []*armapimanagement.UserContract {
	userClient := armapimanagement.NewUserClient(subscriptionID, cred, nil)

	pager := userClient.ListByService(resourceGroupName, serviceName, nil)

	users := make([]*armapimanagement.UserContract, 0)
	for pager.NextPage(ctx) {
		resp := pager.PageResponse()
		users = append(users, resp.Value...)
	}

	return users
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
