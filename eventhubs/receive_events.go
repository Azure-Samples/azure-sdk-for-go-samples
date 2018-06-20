package eventhubs

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure/azure-amqp-common-go/aad"
	"github.com/Azure/azure-amqp-common-go/persist"
	eventhubs "github.com/Azure/azure-event-hubs-go"
)

func Receive(ctx context.Context, nsName, hubName string) {
	// create an access token provider using AAD principal defined in environment
	provider, err := aad.NewJWTProvider(aad.JWTProviderWithEnvironmentVars())
	if err != nil {
		log.Fatalf("failed to configure AAD JWT provider: %s\n", err)
	}

	// get an existing hub for dataplane use
	hub, err := eventhubs.NewHub(nsName, hubName, provider)
	if err != nil {
		log.Fatalf("failed to get hub: %s\n", err)
	}

	// get info about the hub, particularly number and IDs of partitions
	info, err := hub.GetRuntimeInformation(ctx)
	if err != nil {
		log.Fatalf("failed to get runtime info: %s\n", err)
	}
	log.Printf("partition IDs: %s\n", info.PartitionIDs)

	// set up wait group to wait for expected message
	eventReceived := make(chan struct{})

	// declare handler for incoming events
	handler := func(ctx context.Context, event *eventhubs.Event) error {
		fmt.Printf("received: %s\n", string(event.Data))
		// notify channel that event was received
		eventReceived <- struct{}{}
		return nil
	}

	for _, partitionID := range info.PartitionIDs {
		_, err := hub.Receive(
			ctx,
			partitionID,
			handler,
			eventhubs.ReceiveWithStartingOffset(persist.StartOfStream),
		)
		if err != nil {
			log.Fatalf("failed to receive for partition ID %s: %s\n", partitionID, err)
		}
	}

	// don't exit till event is received by handler
	select {
	case <-eventReceived:
	case err := <-ctx.Done():
		log.Fatalf("context cancelled before event received: %s\n", err)
	}
}
