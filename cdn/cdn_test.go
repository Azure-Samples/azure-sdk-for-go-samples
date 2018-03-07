package cdn

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal"
)

func ExampleCheckNameAvailability() {
	ctx := context.Background()

	available, err := CheckNameAvailability(ctx, "gocdnname"+internal.GetRandomLetterSequence(6), "Microsoft.Cdn/Profiles/Endpoints")
	if err != nil {
		log.Fatalf("cannot check availability: %v", err)
	}
	fmt.Println(available)
	// Output: true
}
