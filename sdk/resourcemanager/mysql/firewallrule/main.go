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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	location          = "eastus"
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
	log.Println("mysql server:", *server.ID)

	firewallRule, err := createFirewallRule(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("mysql firewall rule:", *firewallRule.ID)

	firewallRule, err = getFirewallRule(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get mysql firewall rule:", *firewallRule.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createServer(ctx context.Context, cred azcore.TokenCredential) (*armmysql.Server, error) {
	serversClient, err := armmysql.NewServersClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := serversClient.BeginCreate(
		ctx,
		resourceGroupName,
		serverName,
		armmysql.ServerForCreate{
			Location: to.Ptr(location),
			Properties: &armmysql.ServerPropertiesForCreate{
				CreateMode: to.Ptr(armmysql.CreateModeDefault),
			},
			SKU: &armmysql.SKU{
				Name:     to.Ptr("GP_Gen5_2"),
				Tier:     to.Ptr(armmysql.SKUTierGeneralPurpose),
				Capacity: to.Ptr[int32](2),
				Family:   to.Ptr("Gen5"),
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

func createFirewallRule(ctx context.Context, cred azcore.TokenCredential) (*armmysql.FirewallRule, error) {
	firewallRulesClient, err := armmysql.NewFirewallRulesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := firewallRulesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		firewallRuleName,
		armmysql.FirewallRule{
			Properties: &armmysql.FirewallRuleProperties{
				StartIPAddress: to.Ptr("0.0.0.0"),
				EndIPAddress:   to.Ptr("255.255.255.255"),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, 30*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.FirewallRule, nil
}

func getFirewallRule(ctx context.Context, cred azcore.TokenCredential) (*armmysql.FirewallRule, error) {
	firewallRulesClient, err := armmysql.NewFirewallRulesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := firewallRulesClient.Get(ctx, resourceGroupName, serverName, firewallRuleName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.FirewallRule, nil
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
