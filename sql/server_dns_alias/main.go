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
	"github.com/Azure/azure-sdk-for-go/sdk/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/sql/armsql"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	serverName        = "sample2server"
	dnsAliasName      = "sample-dns-alias"
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

	server, err := createServer(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("server:", *server.ID)

	serverDNSAlias, err := createServerDNSAlias(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("server dns alias:", *serverDNSAlias.ID)

	serverDNSAlias, err = getServerDNSAlias(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get server dns alias:", *serverDNSAlias.ID)

	//keepResource := os.Getenv("KEEP_RESOURCE")
	//if len(keepResource) == 0 {
	//	_, err := cleanup(ctx, conn)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	log.Println("cleaned up successfully.")
	//}
}

func createServer(ctx context.Context, conn *arm.Connection) (*armsql.Server, error) {
	serversClient := armsql.NewServersClient(conn, subscriptionID)

	pollerResp, err := serversClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		armsql.Server{
			TrackedResource: armsql.TrackedResource{
				Location: to.StringPtr(location),
			},
			Properties: &armsql.ServerProperties{
				AdministratorLogin:         to.StringPtr("dummylogin"),
				AdministratorLoginPassword: to.StringPtr("QWE123!@#"),
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
	return &resp.Server, nil
}

func createServerDNSAlias(ctx context.Context, conn *arm.Connection) (*armsql.ServerDNSAlias, error) {
	serverDNSAliasesClient := armsql.NewServerDNSAliasesClient(conn, subscriptionID)

	pollerResp, err := serverDNSAliasesClient.BeginCreateOrUpdate(ctx, resourceGroupName, serverName, dnsAliasName, nil)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.ServerDNSAlias, nil
}

func getServerDNSAlias(ctx context.Context, conn *arm.Connection) (*armsql.ServerDNSAlias, error) {
	serverDNSAliasesClient := armsql.NewServerDNSAliasesClient(conn, subscriptionID)

	resp, err := serverDNSAliasesClient.Get(ctx, resourceGroupName, serverName, dnsAliasName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ServerDNSAlias, nil
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
