package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/armcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/to"
	"log"
	"net/http"
	"os"
	"time"
)

var subscriptionID string
var location = "westus"
var resourceGroupName = "sample-resource-group"
var publicIPAddressName = "sample-public-ip"

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

	resourceGroup,err := createResourceGroup(ctx,conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resource group:",*resourceGroup.ID)

	publicIP,err := createPublicIP(ctx,conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("public IP:",*publicIP.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_,err := cleanup(ctx,conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createPublicIP(ctx context.Context, conn *armcore.Connection) (*armnetwork.PublicIPAddress, error) {
	publicIPClient := armnetwork.NewPublicIPAddressesClient(conn, subscriptionID)

	pollerResp,err := publicIPClient.BeginCreateOrUpdate(
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

	resp,err := pollerResp.PollUntilDone(ctx, 10 * time.Second)
	if err != nil {
		return nil, err
	}
	return resp.PublicIPAddress,nil
}

func createResourceGroup(ctx context.Context,conn *armcore.Connection) (*armresources.ResourceGroup,error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn,subscriptionID)

	resourceGroupResp,err := resourceGroupClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		armresources.ResourceGroup{
			Location: to.StringPtr(location),
		},
		nil)
	if err != nil {
		return nil,err
	}
	return resourceGroupResp.ResourceGroup,nil
}

func cleanup(ctx context.Context, conn *armcore.Connection) (*http.Response,error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn,subscriptionID)

	pollerResp,err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return nil,err
	}

	resp,err := pollerResp.PollUntilDone(ctx, 10 * time.Second)
	if err != nil {
		return nil, err
	}
	return resp,nil
}
