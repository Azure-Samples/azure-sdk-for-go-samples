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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
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
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createServer(ctx context.Context, cred azcore.TokenCredential) (*armsql.Server, error) {
	serversClient := armsql.NewServersClient(subscriptionID, cred, nil)

	pollerResp, err := serversClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		armsql.Server{
			Location: to.StringPtr(location),
			Properties: &armsql.ServerProperties{
				AdministratorLogin:         to.StringPtr("samplelogin"),
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

func createDatabase(ctx context.Context, cred azcore.TokenCredential) (*armsql.Database, error) {
	databasesClient := armsql.NewDatabasesClient(subscriptionID, cred, nil)

	pollerResp, err := databasesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		databaseName,
		armsql.Database{
			Location: to.StringPtr(location),
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
	return &resp.Database, nil
}

func createSyncDatabase(ctx context.Context, cred azcore.TokenCredential) (*armsql.Database, error) {
	databasesClient := armsql.NewDatabasesClient(subscriptionID, cred, nil)

	pollerResp, err := databasesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		syncDatabaseName,
		armsql.Database{
			Location: to.StringPtr(location),
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
	return &resp.Database, nil
}

func createSyncAgent(ctx context.Context, cred azcore.TokenCredential, syncDatabaseID string) (*armsql.SyncAgent, error) {
	syncAgentsClient := armsql.NewSyncAgentsClient(subscriptionID, cred, nil)

	pollerResp, err := syncAgentsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		syncAgentName,
		armsql.SyncAgent{
			Properties: &armsql.SyncAgentProperties{
				SyncDatabaseID: to.StringPtr(syncDatabaseID),
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
	return &resp.SyncAgent, nil
}

func createSyncGroup(ctx context.Context, cred azcore.TokenCredential, syncDatabaseID string) (*armsql.SyncGroup, error) {
	syncGroupsClient := armsql.NewSyncGroupsClient(subscriptionID, cred, nil)

	pollerResp, err := syncGroupsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		databaseName,
		syncGroupName,
		armsql.SyncGroup{
			Properties: &armsql.SyncGroupProperties{
				Interval:                 to.Int32Ptr(-1),
				ConflictResolutionPolicy: armsql.SyncConflictResolutionPolicyHubWin.ToPtr(),
				SyncDatabaseID:           to.StringPtr(syncDatabaseID),
				HubDatabaseUserName:      to.StringPtr("hubUser"),
				UsePrivateLinkConnection: to.BoolPtr(false),
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
	return &resp.SyncGroup, nil
}

func createSyncMember(ctx context.Context, cred azcore.TokenCredential) (*armsql.SyncMember, error) {
	syncMembersClient := armsql.NewSyncMembersClient(subscriptionID, cred, nil)

	pollerResp, err := syncMembersClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		databaseName,
		syncGroupName,
		syncMemberName,
		armsql.SyncMember{
			Properties: &armsql.SyncMemberProperties{
				DatabaseType:             armsql.SyncMemberDbTypeAzureSQLDatabase.ToPtr(),
				ServerName:               to.StringPtr(serverName),
				DatabaseName:             to.StringPtr(databaseName),
				UserName:                 to.StringPtr("dummylogin"),
				Password:                 to.StringPtr("QWE123!@#"),
				SyncDirection:            armsql.SyncDirectionBidirectional.ToPtr(),
				UsePrivateLinkConnection: to.BoolPtr(false),
				SyncState:                armsql.SyncMemberStateUnProvisioned.ToPtr(),
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
	return &resp.SyncMember, nil
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
