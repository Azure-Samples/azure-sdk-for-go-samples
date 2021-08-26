package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/armcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/to"
)

var (
	subscriptionID      string
	location            = "westus"
	resourceGroupName   = "sample-resource-group"
	publicIPAddressName = "sample-public-ip"
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

	conn := armcore.NewDefaultConnection(cred, &armcore.ConnectionOptions{
		Logging: azcore.LogOptions{
			IncludeBody: true,
		},
	})
	ctx := context.Background()

	resourceGroup, err := createResourceGroup(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	publicIP, err := createPublicIP(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("public IP:", *publicIP.ID)

	exist, err := checkExistResource(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources is exist:", exist)

	resources, err := createResource(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("created resources:", *resources.ID)

	genericResource, err := getResource(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get resource:", *genericResource.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

var resourceProviderNamespace = "Microsoft.Network"
var resourceType = "publicIPAddresses"
var apiVersion = "2021-02-01"

func checkExistResource(ctx context.Context, conn *armcore.Connection) (bool, error) {
	resourceClient := armresources.NewResourcesClient(conn, subscriptionID)

	boolResp, err := resourceClient.CheckExistence(
		ctx,
		resourceGroupName,
		resourceProviderNamespace,
		"",
		resourceType,
		publicIPAddressName,
		apiVersion,
		nil)
	if err != nil {
		return false, err
	}

	return boolResp.Success, nil
}

func createResource(ctx context.Context, conn *armcore.Connection) (*armresources.GenericResource, error) {
	resourceClient := armresources.NewResourcesClient(conn, subscriptionID)

	pollerResp, err := resourceClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		resourceProviderNamespace,
		resourceGroupName,
		resourceType,
		publicIPAddressName,
		apiVersion,
		armresources.GenericResource{},
		nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}

	return resp.GenericResource, nil
}

func getResource(ctx context.Context, conn *armcore.Connection) (*armresources.GenericResource, error) {
	resourceClient := armresources.NewResourcesClient(conn, subscriptionID)

	resp, err := resourceClient.Get(
		ctx,
		resourceGroupName,
		resourceProviderNamespace,
		"/",
		resourceType,
		publicIPAddressName,
		apiVersion,
		nil)
	if err != nil {
		return nil, err
	}

	return resp.GenericResource, nil
}

func createPublicIP(ctx context.Context, conn *armcore.Connection) (*armnetwork.PublicIPAddress, error) {
	publicIPClient := armnetwork.NewPublicIPAddressesClient(conn, subscriptionID)

	pollerResp, err := publicIPClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		publicIPAddressName,
		armnetwork.PublicIPAddress{
			Resource: armnetwork.Resource{
				Name:     to.StringPtr(publicIPAddressName),
				Location: to.StringPtr(location),
			},
			Properties: &armnetwork.PublicIPAddressPropertiesFormat{
				PublicIPAddressVersion:   armnetwork.IPVersionIPv4.ToPtr(),
				PublicIPAllocationMethod: armnetwork.IPAllocationMethodStatic.ToPtr(),
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
	return resp.PublicIPAddress, nil
}

func createResourceGroup(ctx context.Context, conn *armcore.Connection) (*armresources.ResourceGroup, error) {
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
	return resourceGroupResp.ResourceGroup, nil
}

func cleanup(ctx context.Context, conn *armcore.Connection) (*http.Response, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn, subscriptionID)

	pollerResp, err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
