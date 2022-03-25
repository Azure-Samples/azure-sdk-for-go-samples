// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

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
	partnerServerName = "sample2partner2server"
	failoverGroupName = "sample-failover-group"
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

	partnerServer, err := createPartnerServer(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("partner server:", *partnerServer.ID)

	failoverGroup, err := createFailoverGroup(ctx, cred, *partnerServer.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("failover group:", *failoverGroup.ID)

	failoverGroup, err = getFailoverGroup(ctx, cred, *partnerServer.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get failover group:", *failoverGroup.ID)

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

func createPartnerServer(ctx context.Context, cred azcore.TokenCredential) (*armsql.Server, error) {
	serversClient := armsql.NewServersClient(subscriptionID, cred, nil)

	pollerResp, err := serversClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		partnerServerName,
		armsql.Server{
			Location: to.StringPtr("eastus2"),
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

func createFailoverGroup(ctx context.Context, cred azcore.TokenCredential, partnerServerID string) (*armsql.FailoverGroup, error) {
	failoverGroupsClient := armsql.NewFailoverGroupsClient(subscriptionID, cred, nil)

	pollerResp, err := failoverGroupsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		failoverGroupName,
		armsql.FailoverGroup{
			Properties: &armsql.FailoverGroupProperties{
				PartnerServers: []*armsql.PartnerInfo{
					{
						ID: to.StringPtr(partnerServerID),
					},
				},
				ReadWriteEndpoint: &armsql.FailoverGroupReadWriteEndpoint{
					FailoverPolicy:                         armsql.ReadWriteEndpointFailoverPolicyAutomatic.ToPtr(),
					FailoverWithDataLossGracePeriodMinutes: to.Int32Ptr(480),
				},
				Databases: []*string{},
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
	return &resp.FailoverGroup, nil
}

func getFailoverGroup(ctx context.Context, cred azcore.TokenCredential, partnerServerID string) (*armsql.FailoverGroup, error) {
	failoverGroupsClient := armsql.NewFailoverGroupsClient(subscriptionID, cred, nil)

	resp, err := failoverGroupsClient.Get(ctx, resourceGroupName, serverName, failoverGroupName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.FailoverGroup, nil
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
