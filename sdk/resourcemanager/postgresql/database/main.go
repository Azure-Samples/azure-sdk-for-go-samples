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
	serverName        = "sampleXserver"
	databaseName      = "sample-postgresql-database"
)

var (
	resourcesClientFactory  *armresources.ClientFactory
	postgresqlClientFactory *armpostgresql.ClientFactory
)

var (
	resourceGroupClient *armresources.ResourceGroupsClient
	serversClient       *armpostgresql.ServersClient
	databasesClient     *armpostgresql.DatabasesClient
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
	serversClient = postgresqlClientFactory.NewServersClient()
	databasesClient = postgresqlClientFactory.NewDatabasesClient()

	resourceGroup, err := createResourceGroup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	server, err := createServer(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("postgresql server:", *server.ID)

	database, err := createDatabase(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("postgresql database:", *database.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createServer(ctx context.Context) (*armpostgresql.Server, error) {

	pollerResp, err := serversClient.BeginCreate(
		ctx,
		resourceGroupName,
		serverName,
		armpostgresql.ServerForCreate{
			Location: to.Ptr(location),
			Properties: &armpostgresql.ServerPropertiesForDefaultCreate{
				CreateMode:                 to.Ptr(armpostgresql.CreateModeDefault),
				InfrastructureEncryption:   to.Ptr(armpostgresql.InfrastructureEncryptionDisabled),
				PublicNetworkAccess:        to.Ptr(armpostgresql.PublicNetworkAccessEnumEnabled),
				Version:                    to.Ptr(armpostgresql.ServerVersionEleven),
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

func createDatabase(ctx context.Context) (*armpostgresql.Database, error) {

	pollerResp, err := databasesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		databaseName,
		armpostgresql.Database{
			Properties: &armpostgresql.DatabaseProperties{
				Charset:   to.Ptr("UTF8"),
				Collation: to.Ptr("English_United States.1252"),
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
	return &resp.Database, nil
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
