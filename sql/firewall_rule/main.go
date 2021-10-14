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
	firewallRuleName  = "sample-firewall-rule"
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

	firewallRule, err := createFirewallRule(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("firewall rule:", *firewallRule.ID)

	firewallRule, err = getFirewallRule(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get firewall rule:", *firewallRule.ID)

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

func createFirewallRule(ctx context.Context, conn *arm.Connection) (*armsql.FirewallRule, error) {
	firewallRulesClient := armsql.NewFirewallRulesClient(conn, subscriptionID)

	resp, err := firewallRulesClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		firewallRuleName,
		armsql.FirewallRule{
			Properties: &armsql.ServerFirewallRuleProperties{
				StartIPAddress: to.StringPtr("0.0.0.3"),
				EndIPAddress:   to.StringPtr("0.0.0.3"),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.FirewallRule, nil
}

func getFirewallRule(ctx context.Context, conn *arm.Connection) (*armsql.FirewallRule, error) {
	firewallRulesClient := armsql.NewFirewallRulesClient(conn, subscriptionID)

	resp, err := firewallRulesClient.Get(ctx, resourceGroupName, serverName, firewallRuleName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.FirewallRule, nil
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
