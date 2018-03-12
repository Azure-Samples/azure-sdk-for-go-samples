package cdn

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
)

func TestMain(m *testing.M) {
	err := helpers.ParseArgs()
	if err != nil {
		log.Fatalln("failed to parse args")
	}
	os.Exit(m.Run())
}

func ExampleCheckNameAvailability() {
	ctx := context.Background()

	available, err := CheckNameAvailability(ctx, "gocdnname"+helpers.GetRandomLetterSequence(6), "Microsoft.Cdn/Profiles/Endpoints")
	if err != nil {
		log.Fatalf("cannot check availability: %v", err)
	}
	fmt.Println(available)
	// Output: true
}
