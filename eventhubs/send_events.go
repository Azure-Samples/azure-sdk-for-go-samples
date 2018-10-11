package eventhubs

import (
	"context"
	"log"

	// "bufio"
	// "os"

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
	defer hub.Close(ctx)
	if err != nil {
		log.Fatalf("failed to get hub: %s\n", err)
	}

	// get info about partitions in hub
	info, err := hub.GetRuntimeInformation(ctx)
	if err != nil {
		log.Fatalf("failed to get runtime info: %s\n", err)
	}
	log.Printf("partition IDs: %s\n", info.PartitionIDs)

	// send messages to hub
	// reader := bufio.NewReader(os.Stdin)
	// for {
	// 	fmt.Printf("Input message to send: ")
	// 	text, _ := reader.ReadString('\n')
	// 	hub.Send(ctx, eventhubs.NewEventFromString(text))
	// }

	// send message to hub.
	// by default the destination partition is selected round-robin by the
	// Event Hubs service
	hub.Send(ctx, eventhubs.NewEventFromString("test-message"))
}
