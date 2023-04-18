// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"log"
	"os"
)

var (
	subscriptionID    string
	location          = "eastus"
	resourceGroupName = "sample-resource-group"
	serverName        = "sample2server"
)

var (
	resourcesClientFactory  *armresources.ClientFactory
	postgresqlClientFactory *armpostgresql.ClientFactory
)

var (
	resourceGroupClient         *armresources.ResourceGroupsClient
	checkNameAvailabilityClient *armpostgresql.CheckNameAvailabilityClient
	serversClient               *armpostgresql.ServersClient
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

	resourcesClientFactory, err = armresources.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	resourceGroupClient = resourcesClientFactory.NewResourceGroupsClient()

	postgresqlClientFactory, err = armpostgresql.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	checkNameAvailabilityClient = postgresqlClientFactory.NewCheckNameAvailabilityClient()
	serversClient = postgresqlClientFactory.NewServersClient()

	resourceGroup, err := createResourceGroup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	availability, err := checkNameAvailability(ctx, serverName)
	if err != nil {
		log.Println(err)
	}
	log.Println("check name availability:", *availability.NameAvailable)

	server, err := createServer(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("postgresql server:", *server.ID)

	server, err = updateServer(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("postgresql update server:", *server.ID)

	err = restartServer(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("postgresql restart server")

	server, err = getServer(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get postgresql server:", *server.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func checkNameAvailability(ctx context.Context, checkName string) (*armpostgresql.NameAvailability, error) {

	resp, err := checkNameAvailabilityClient.Execute(
		ctx,
		armpostgresql.NameAvailabilityRequest{
			Name: to.Ptr(checkName),
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.NameAvailability, nil
}

func createServer(ctx context.Context) (*armpostgresql.Server, error) {

	pollerResp, err := serversClient.BeginCreate(
		ctx,
		resourceGroupName,
		serverName,
		armpostgresql.ServerForCreate{
			Location: to.Ptr(location),
			Properties: &armpostgresql.ServerPropertiesForDefaultCreate{
				AdministratorLogin:         to.Ptr("dummylogin"),
				AdministratorLoginPassword: to.Ptr("QWE123!@#"),
			},
			SKU: &armpostgresql.SKU{
				Name: to.Ptr("B_Gen5_1"),
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
	return &resp.Server, nil
}

func updateServer(ctx context.Context) (*armpostgresql.Server, error) {

	pollerResp, err := serversClient.BeginUpdate(
		ctx,
		resourceGroupName,
		serverName,
		armpostgresql.ServerUpdateParameters{
			Properties: &armpostgresql.ServerUpdateParametersProperties{
				AdministratorLoginPassword: to.Ptr("QWE123!@#"),
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
	return &resp.Server, nil
}

func restartServer(ctx context.Context) error {

	pollerResp, err := serversClient.BeginRestart(ctx, resourceGroupName, serverName, nil)
	if err != nil {
		return err
	}
	_, err = pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}
	return nil
}

func getServer(ctx context.Context) (*armpostgresql.Server, error) {

	resp, err := serversClient.Get(ctx, resourceGroupName, serverName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Server, nil
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
