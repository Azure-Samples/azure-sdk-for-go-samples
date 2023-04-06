// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package azstart

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

// ExampleUsingARMClients shows how to construct & use an ARM Client to invoke service methods
func ExampleUsingARMClients() {
	// Construct a credential type from the azidentity package
	// or the module defining the client you wish to use
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		panic(err)
	}

	// Construct an ARM client factory passing subscription ID, credential, & optional options
	// which could be used to create any client in one ARM module
	clientFactory, err := armresources.NewClientFactory("<SubscriptionId>", credential, nil)

	// This example creates a ResourceGroupsClient, but you can create any ARM client
	client := clientFactory.NewResourceGroupsClient()

	// You can now call client methods to invoke service operations
	// This example calls CreateOrUpdate, but you can call any client method
	response, err := client.CreateOrUpdate(context.TODO(), "<ResourceGroupName>",
		armresources.ResourceGroup{
			Location: to.Ptr("<ResouceGroupLocation>"), // to.Ptr converts this string to a *string
		}, nil)
	if err != nil {
		panic(err)
	}

	// Use the service's response as your application desires
	fmt.Printf("Resource group ID: %s\n", *response.ResourceGroup.ID)
}

// ExampleUsingDPClients shows how to construct & use a data-plane Client to invoke service methods
func ExampleUsingDataPlaneClients() {
	// Construct a credential type from the azidentity package
	// or the module defining the client you wish to use
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		panic(err)
	}

	// Construct a DP client passing endpointURL, credential, & optional options
	client, err := azsecrets.NewClient("https://<KeyVaultName>.vault.azure.net/", credential, nil)
	if err != nil {
		panic(err)
	}

	// You can now call client methods to invoke service operations
	response, err := client.SetSecret(context.TODO(), "<SecretName>",
		azsecrets.SetSecretParameters{Value: to.Ptr("<SecretValue>")}, nil)
	if err != nil {
		panic(err)
	}

	// Use the service's response as your application desires
	fmt.Printf("Name: %s, Value: %s\n", *response.ID, *response.Value)
}

// PagingOverACollection shows how to page over a collection's items
func ExamplePagingOverACollection() {
	// Construct a credential type from the azidentity
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		panic(err)
	}

	client, err := armresources.NewResourceGroupsClient("<SubscriptionId>", credential, nil)
	if err != nil {
		panic(err)
	}

	// Call a client method that creates a XxxPager; this does NOT invoke a service operation
	for pager := client.NewListPager(nil); pager.More(); {
		// While pages are getable, request a page of items from the service
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			panic(err)
		}

		// Process the page's items.
		// NOTE: The service desides how many items to return on a page.
		// If a page has 0 items, go get the next page.
		// Other clients may be adding/deleting items from the collection while
		// this code is paging; some items may be skipped or returned multiple times.
		for _, item := range page.Value {
			_ = item // Here's where your code processes the item as you desire
		}
		// Looping around will request the next page of items from the service
	}
}

// LongRunningOperation shows how to invoke a long-running operation and poll for its completion
func ExampleLongRunningOperation() {
	// Construct a credential type from the azidentity
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		panic(err)
	}

	client, err := armresources.NewResourceGroupsClient("<SubscriptionId>", credential, nil)
	if err != nil {
		panic(err)
	}

	// Initiating a long-Running Operation causes the method to return a Poller[T]
	poller, err := client.BeginDelete(context.TODO(), "<ResourceGroupName>", nil)
	if err != nil {
		panic(err)
	}

	// PollUntilDone causes your goroutine to periodically ask the service the status of the LRO
	// It ultimately returns when the operation succeeds, fails, or was canceled.
	// If the operation succeeds, err == nil and lroResult has result (if any); else err != nil
	lroResult, err := poller.PollUntilDone(context.TODO(), nil)
	if err != nil {
		panic(err)
	}
	_ = lroResult // Examine sucessful result (if any)
}
