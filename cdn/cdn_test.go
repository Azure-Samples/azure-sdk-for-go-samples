package cdn

import (
	"context"
	"fmt"
	"log"
)

func ExampleCheckNameAvailability() {
	ctx := context.Background()

	available, err := CheckNameAvailability(ctx, "gocdnname", "Microsoft.Cdn/Profiles/Endpoints")
	if err != nil {
		log.Fatalf("cannot check availability: %v", err)
	}
	fmt.Println(available)
	// Output: true
}
