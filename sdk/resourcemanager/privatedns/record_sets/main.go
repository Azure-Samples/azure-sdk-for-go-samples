// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"log"
	"os"
)

var (
	subscriptionID        string
	location              = "westus"
	resourceGroupName     = "sample-resource-group"
	privateZoneName       = "sample-private-zone"
	relativeRecordSetName = "sample-relative-record-set"
)

var (
	resourcesClientFactory  *armresources.ClientFactory
	privatednsClientFactory *armprivatedns.ClientFactory
)

var (
	resourceGroupClient *armresources.ResourceGroupsClient
	privateZonesClient  *armprivatedns.PrivateZonesClient
	recordSetsClient    *armprivatedns.RecordSetsClient
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

	privatednsClientFactory, err = armprivatedns.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	privateZonesClient = privatednsClientFactory.NewPrivateZonesClient()
	recordSetsClient = privatednsClientFactory.NewRecordSetsClient()

	resourceGroup, err := createResourceGroup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	privateZone, err := createPrivateZone(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("private zone:", *privateZone.ID)

	recordSets, err := createRecordSets(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("record sets:", *recordSets.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createPrivateZone(ctx context.Context) (*armprivatedns.PrivateZone, error) {

	pollersResp, err := privateZonesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		privateZoneName,
		armprivatedns.PrivateZone{
			Location: to.Ptr(location),
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollersResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.PrivateZone, nil
}

func createRecordSets(ctx context.Context) (*armprivatedns.RecordSet, error) {

	resp, err := recordSetsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		privateZoneName,
		armprivatedns.RecordTypeA,
		relativeRecordSetName,
		armprivatedns.RecordSet{
			Properties: &armprivatedns.RecordSetProperties{
				ARecords: []*armprivatedns.ARecord{},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.RecordSet, nil
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
