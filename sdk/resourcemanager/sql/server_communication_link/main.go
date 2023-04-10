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
	subscriptionID        string
	location              = "eastus"
	resourceGroupName     = "sample-resource-group"
	serverName            = "sample2server"
	partnerServerName     = "sample2partner2server"
	communicationLinkName = "sample2communication2link"
)

var (
	resourcesClientFactory *armresources.ClientFactory
	sqlClientFactory       *armsql.ClientFactory
)

var (
	resourceGroupClient            *armresources.ResourceGroupsClient
	serversClient                  *armsql.ServersClient
	serverCommunicationLinksClient *armsql.ServerCommunicationLinksClient
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
	serverCommunicationLinksClient = sqlClientFactory.NewServerCommunicationLinksClient()

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

	partnerServer, err := createPartnerServer(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("partner server:", *partnerServer.ID)

	serverCommunicationLink, err := createServerCommunicationLink(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("server communication link:", *serverCommunicationLink.ID)

	serverCommunicationLink, err = getServerCommunicationLink(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get server communication link:", *serverCommunicationLink.ID)

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
				AdministratorLogin:         to.Ptr("dummylogin"),
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

func createPartnerServer(ctx context.Context) (*armsql.Server, error) {

	pollerResp, err := serversClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		partnerServerName,
		armsql.Server{
			Location: to.Ptr(location),
			Properties: &armsql.ServerProperties{
				AdministratorLogin:         to.Ptr("dummylogin"),
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

func createServerCommunicationLink(ctx context.Context) (*armsql.ServerCommunicationLink, error) {

	pollerResp, err := serverCommunicationLinksClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		communicationLinkName,
		armsql.ServerCommunicationLink{
			Properties: &armsql.ServerCommunicationLinkProperties{
				PartnerServer: to.Ptr(partnerServerName),
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
	return &resp.ServerCommunicationLink, nil
}

func getServerCommunicationLink(ctx context.Context) (*armsql.ServerCommunicationLink, error) {

	resp, err := serverCommunicationLinksClient.Get(ctx, resourceGroupName, serverName, communicationLinkName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ServerCommunicationLink, nil
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
