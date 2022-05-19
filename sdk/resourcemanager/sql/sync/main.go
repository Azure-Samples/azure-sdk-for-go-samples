// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
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

	server, err := createServer(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("server:", *server.ID)

	database, err := createDatabase(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("database:", *database.ID)

	syncDatabase, err := createSyncDatabase(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("sync database:", *syncDatabase.ID)

	syncAgent, err := createSyncAgent(ctx, cred, *syncDatabase.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("sync agent:", *syncAgent.ID)

	syncGroup, err := createSyncGroup(ctx, cred, *syncDatabase.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("sync group:", *syncGroup.ID)

	syncMember, err := createSyncMember(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("sync member:", *syncMember.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createServer(ctx context.Context, cred azcore.TokenCredential) (*armsql.Server, error) {
	serversClient, err := armsql.NewServersClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

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

func createDatabase(ctx context.Context, cred azcore.TokenCredential) (*armsql.Database, error) {
	databasesClient, err := armsql.NewDatabasesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

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

func createSyncDatabase(ctx context.Context, cred azcore.TokenCredential) (*armsql.Database, error) {
	databasesClient, err := armsql.NewDatabasesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

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

func createSyncAgent(ctx context.Context, cred azcore.TokenCredential, syncDatabaseID string) (*armsql.SyncAgent, error) {
	syncAgentsClient, err := armsql.NewSyncAgentsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

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

func createSyncGroup(ctx context.Context, cred azcore.TokenCredential, syncDatabaseID string) (*armsql.SyncGroup, error) {
	syncGroupsClient, err := armsql.NewSyncGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

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

func createSyncMember(ctx context.Context, cred azcore.TokenCredential) (*armsql.SyncMember, error) {
	syncMembersClient, err := armsql.NewSyncMembersClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

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
