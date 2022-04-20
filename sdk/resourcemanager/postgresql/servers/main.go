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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	location          = "eastus"
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
	ctx := context.Background()

	resourceGroup, err := createResourceGroup(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	availability, err := checkNameAvailability(ctx, cred, serverName)
	if err != nil {
		log.Println(err)
	}
	log.Println("check name availability:", *availability.NameAvailable)

	server, err := createServer(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("postgresql server:", *server.ID)

	server, err = updateServer(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("postgresql update server:", *server.ID)

	err = restartServer(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("postgresql restart server")

	server, err = getServer(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get postgresql server:", *server.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func checkNameAvailability(ctx context.Context, cred azcore.TokenCredential, checkName string) (*armpostgresql.NameAvailability, error) {
	checkNameAvailabilityClient, err := armpostgresql.NewCheckNameAvailabilityClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

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

func createServer(ctx context.Context, cred azcore.TokenCredential) (*armpostgresql.Server, error) {
	serversClient, err := armpostgresql.NewServersClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

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
	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.Server, nil
}

func updateServer(ctx context.Context, cred azcore.TokenCredential) (*armpostgresql.Server, error) {
	serversClient, err := armpostgresql.NewServersClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

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
	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.Server, nil
}

func restartServer(ctx context.Context, cred azcore.TokenCredential) error {
	serversClient, err := armpostgresql.NewServersClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	pollerResp, err := serversClient.BeginRestart(ctx, resourceGroupName, serverName, nil)
	if err != nil {
		return err
	}
	_, err = pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return err
	}
	return nil
}

func getServer(ctx context.Context, cred azcore.TokenCredential) (*armpostgresql.Server, error) {
	serversClient, err := armpostgresql.NewServersClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := serversClient.Get(ctx, resourceGroupName, serverName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Server, nil
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
