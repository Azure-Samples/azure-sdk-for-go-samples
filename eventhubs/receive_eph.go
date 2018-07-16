package eventhubs

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure/azure-amqp-common-go/aad"
	eventhubs "github.com/Azure/azure-event-hubs-go"
	"github.com/Azure/azure-event-hubs-go/eph"
	eventhubsstorage "github.com/Azure/azure-event-hubs-go/storage"
	"github.com/Azure/go-autorest/autorest/azure"

	// imports within this repo
	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/storage"
)

// ReceiveViaEPH sets up an Event Processor Host (EPH), a small framework to
// receive events from several partitions.
func ReceiveViaEPH(ctx context.Context, nsName, hubName, storageAccountName, storageContainerName string) {
	// create an access token provider using AAD principal defined in environment
	tokenProvider, err := aad.NewJWTProvider(aad.JWTProviderWithEnvironmentVars())
	if err != nil {
		log.Fatalf("failed to configure AAD JWT provider: %s\n", err)
	}

	// create a storage account and container to maintain dictionary of leases
	// and checkpoints
	_, err = storage.CreateStorageAccount(ctx, storageAccountName)
	if err != nil {
		log.Fatalf("could not create storage account: %s\n", err)
	}
	log.Printf("creating storage container\n")
	_, err = storage.CreateContainer(ctx, storageAccountName, storageContainerName)
	if err != nil {
		log.Fatalf("could not create storage container: %s\n", err)
	}

	// use helper method to exchange AAD credentials for SAS token
	cred, err := eventhubsstorage.NewAADSASCredential(
		helpers.SubscriptionID(),
		helpers.ResourceGroupName(),
		storageAccountName,
		storageContainerName,
		eventhubsstorage.AADSASCredentialWithEnvironmentVars())
	if err != nil {
		log.Fatalf("could not prepare a storage credential: %s\n", err)
	}

	// create a leaser and checkpointer backed by a storage container
	leaserCheckpointer, err := eventhubsstorage.NewStorageLeaserCheckpointer(
		cred,
		storageAccountName,
		storageContainerName,
		azure.PublicCloud)
	if err != nil {
		log.Fatalf("could not prepare a storage leaserCheckpointer: %s\n", err)
	}

	// use all of the above to create an Event Processor
	p, err := eph.New(
		ctx,
		nsName,
		hubName,
		tokenProvider,
		leaserCheckpointer,
		leaserCheckpointer,
		eph.WithNoBanner())
	if err != nil {
		log.Fatalf("failed to create EPH: %s\n", err)
	}

	// set up a handler for the Event Processor which will receive a single
	// message, print it to the console, and allow the process to exit
	// set up a channel to notify when event is successfully received
	eventReceived := make(chan struct{})

	handler := func(ctx context.Context, event *eventhubs.Event) error {
		fmt.Printf("received: %s\n", string(event.Data))
		// notify channel that event was received
		eventReceived <- struct{}{}
		return nil
	}

	// register the handler with the Event Processor
	// discard the HandlerID cause we don't need it
	_, err = p.RegisterHandler(ctx, handler)
	if err != nil {
		log.Fatalf("failed to set up handler: %s\n", err)
	}

	// finally, start the Event Processor with a timeout
	err = p.StartNonBlocking(ctx)
	if err != nil {
		log.Fatalf("failed to start EPH: %s\n", err)
	}

	// don't exit till event is received by handler
	select {
	case <-eventReceived:
	case err := <-ctx.Done():
		log.Fatalf("context cancelled before event received: %s\n", err)
	}

	p.Close(ctx)
}
