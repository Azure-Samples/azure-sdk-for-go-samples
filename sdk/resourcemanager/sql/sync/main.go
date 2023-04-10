// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
	"log"
	"os"
)

var (
	subscriptionID    string
	location          = "eastus"
	resourceGroupName = "sample-resource-group"
	serverName        = "sample2server"
	databaseName      = "sample-database"
	syncDatabaseName  = "sample-sync-database"
	syncAgentName     = "sample-sync-agent2"
	syncGroupName     = "sample-sync-group"
	syncMemberName    = "sample-sync-member"
)

var (
	resourcesClientFactory *armresources.ClientFactory
	sqlClientFactory       *armsql.ClientFactory
)

var (
	resourceGroupClient *armresources.ResourceGroupsClient
	serversClient       *armsql.ServersClient
	databasesClient     *armsql.DatabasesClient
	syncAgentsClient    *armsql.SyncAgentsClient
	syncGroupsClient    *armsql.SyncGroupsClient
	syncMembersClient   *armsql.SyncMembersClient
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

	sqlClientFactory, err = armsql.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	serversClient = sqlClientFactory.NewServersClient()
	databasesClient = sqlClientFactory.NewDatabasesClient()
	syncAgentsClient = sqlClientFactory.NewSyncAgentsClient()
	syncGroupsClient = sqlClientFactory.NewSyncGroupsClient()
	syncMembersClient = sqlClientFactory.NewSyncMembersClient()

	resourceGroup, err := createResourceGroup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	server, err := createServer(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("server:", *server.ID)

	database, err := createDatabase(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("database:", *database.ID)

	syncDatabase, err := createSyncDatabase(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("sync database:", *syncDatabase.ID)

	syncAgent, err := createSyncAgent(ctx, *syncDatabase.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("sync agent:", *syncAgent.ID)

	syncGroup, err := createSyncGroup(ctx, *syncDatabase.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("sync group:", *syncGroup.ID)

	syncMember, err := createSyncMember(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("sync member:", *syncMember.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createServer(ctx context.Context) (*armsql.Server, error) {

	pollerResp, err := serversClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		armsql.Server{
			Location: to.Ptr(location),
			Properties: &armsql.ServerProperties{
				AdministratorLogin:         to.Ptr("samplelogin"),
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

func createDatabase(ctx context.Context) (*armsql.Database, error) {

	pollerResp, err := databasesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		databaseName,
		armsql.Database{
			Location: to.Ptr(location),
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

func createSyncDatabase(ctx context.Context) (*armsql.Database, error) {

	pollerResp, err := databasesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		syncDatabaseName,
		armsql.Database{
			Location: to.Ptr(location),
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

func createSyncAgent(ctx context.Context, syncDatabaseID string) (*armsql.SyncAgent, error) {

	pollerResp, err := syncAgentsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		syncAgentName,
		armsql.SyncAgent{
			Properties: &armsql.SyncAgentProperties{
				SyncDatabaseID: to.Ptr(syncDatabaseID),
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
	return &resp.SyncAgent, nil
}

func createSyncGroup(ctx context.Context, syncDatabaseID string) (*armsql.SyncGroup, error) {

	pollerResp, err := syncGroupsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		databaseName,
		syncGroupName,
		armsql.SyncGroup{
			Properties: &armsql.SyncGroupProperties{
				Interval:                 to.Ptr[int32](-1),
				ConflictResolutionPolicy: to.Ptr(armsql.SyncConflictResolutionPolicyHubWin),
				SyncDatabaseID:           to.Ptr(syncDatabaseID),
				HubDatabaseUserName:      to.Ptr("hubUser"),
				UsePrivateLinkConnection: to.Ptr(false),
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
	return &resp.SyncGroup, nil
}

func createSyncMember(ctx context.Context) (*armsql.SyncMember, error) {

	pollerResp, err := syncMembersClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		databaseName,
		syncGroupName,
		syncMemberName,
		armsql.SyncMember{
			Properties: &armsql.SyncMemberProperties{
				DatabaseType:             to.Ptr(armsql.SyncMemberDbTypeAzureSQLDatabase),
				ServerName:               to.Ptr(serverName),
				DatabaseName:             to.Ptr(databaseName),
				UserName:                 to.Ptr("dummylogin"),
				Password:                 to.Ptr("QWE123!@#"),
				SyncDirection:            to.Ptr(armsql.SyncDirectionBidirectional),
				UsePrivateLinkConnection: to.Ptr(false),
				SyncState:                to.Ptr(armsql.SyncMemberStateUnProvisioned),
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
	return &resp.SyncMember, nil
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
