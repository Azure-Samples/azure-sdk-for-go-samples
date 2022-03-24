package eventhubs

import (
	"context"
	"log"

	"github.com/Azure/azure-amqp-common-go/aad"
	eventhubs "github.com/Azure/azure-event-hubs-go"
)

func Send(ctx context.Context, nsName, hubName string) {
	// create an access token provider using an AAD principal
	provider, err := aad.NewJWTProvider(aad.JWTProviderWithEnvironmentVars())
	if err != nil {
		log.Fatalf("failed to configure AAD JWT provider: %s\n", err)
	}

	// get an existing hub
	hub, err := eventhubs.NewHub(nsName, hubName, provider)
	if err != nil {
		log.Fatalf("failed to get hub: %s\n", err)
	}
	defer func() {
		if err := hub.Close(ctx); err != nil {
			log.Fatalf("failed to close event hub: %+v", err)
		}
	}()

	// get info about partitions in hub
	info, err := hub.GetRuntimeInformation(ctx)
	if err != nil {
		log.Fatalf("failed to get runtime info: %+v", err)
	}
	log.Printf("partition IDs: %s\n", info.PartitionIDs)

	// send message to hub.
	// by default the destination partition is selected round-robin by the
	// Event Hubs service
	err = hub.Send(ctx, eventhubs.NewEventFromString("test-message"))
	if err != nil {
		log.Fatalf("failed to send messages: %+v", err)
	}
}
