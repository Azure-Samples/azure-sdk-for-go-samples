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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresqlflexibleservers"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	serverName        = "sample2server"
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

	availability, err := checkNameAvailability(ctx, conn, serverName)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("check name availability:", *availability.NameAvailable)

	server, err := createServer(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("postgresql server:", *server.ID)

	server, err = updateServer(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("postgresql update server:", *server.ID)

	resp, err := restartServer(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("postgresql restart server:", resp)

	server, err = getServer(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get postgresql server:", *server.ID)

	//keepResource := os.Getenv("KEEP_RESOURCE")
	//if len(keepResource) == 0 {
	//	_, err := cleanup(ctx, conn)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	log.Println("cleaned up successfully.")
	//}
}

func checkNameAvailability(ctx context.Context, conn *arm.Connection, checkName string) (*armpostgresqlflexibleservers.NameAvailability, error) {
	checkNameAvailabilityClient := armpostgresqlflexibleservers.NewCheckNameAvailabilityClient(conn, subscriptionID)

	resp, err := checkNameAvailabilityClient.Execute(
		ctx,
		armpostgresqlflexibleservers.NameAvailabilityRequest{
			Name: to.StringPtr(checkName),
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.NameAvailability, nil
}

func createServer(ctx context.Context, conn *arm.Connection) (*armpostgresqlflexibleservers.Server, error) {
	serversClient := armpostgresqlflexibleservers.NewServersClient(conn, subscriptionID)

	pollerResp, err := serversClient.BeginCreate(
		ctx,
		resourceGroupName,
		serverName,
		armpostgresqlflexibleservers.Server{
			TrackedResource: armpostgresqlflexibleservers.TrackedResource{
				Location: to.StringPtr(location),
			},
			Properties: &armpostgresqlflexibleservers.ServerProperties{
				AdministratorLogin:         to.StringPtr("dummylogin"),
				AdministratorLoginPassword: to.StringPtr("QWE123!@#"),
			},
			SKU: &armpostgresqlflexibleservers.SKU{
				Name: to.StringPtr("B_Gen5_1"),
				Tier: armpostgresqlflexibleservers.SKUTierGeneralPurpose.ToPtr(),
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

func updateServer(ctx context.Context, conn *arm.Connection) (*armpostgresqlflexibleservers.Server, error) {
	serversClient := armpostgresqlflexibleservers.NewServersClient(conn, subscriptionID)

	pollerResp, err := serversClient.BeginUpdate(
		ctx,
		resourceGroupName,
		serverName,
		armpostgresqlflexibleservers.ServerForUpdate{
			Properties: &armpostgresqlflexibleservers.ServerPropertiesForUpdate{
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

func restartServer(ctx context.Context, conn *arm.Connection) (*http.Response, error) {
	serversClient := armpostgresqlflexibleservers.NewServersClient(conn, subscriptionID)

	pollerResp, err := serversClient.BeginRestart(ctx, resourceGroupName, serverName, nil)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp.RawResponse, nil
}

func getServer(ctx context.Context, conn *arm.Connection) (*armpostgresqlflexibleservers.Server, error) {
	serversClient := armpostgresqlflexibleservers.NewServersClient(conn, subscriptionID)

	resp, err := serversClient.Get(ctx, resourceGroupName, serverName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Server, nil
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
