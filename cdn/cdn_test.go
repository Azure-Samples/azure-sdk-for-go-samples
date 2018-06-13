package cdn

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"

	"github.com/marstr/randname"
)

func TestMain(m *testing.M) {
	err := iam.ParseArgs()
	if err != nil {
		log.Fatalln("failed to parse IAM args")
	}

	os.Exit(m.Run())
}

func ExampleCheckNameAvailability() {
	ctx := context.Background()

	available, err := CheckNameAvailability(ctx, randname.GenerateWithPrefix("gocdnname", 6), "Microsoft.Cdn/Profiles/Endpoints")
	if err != nil {
		log.Fatalf("cannot check availability: %v", err)
	}
	fmt.Println(available)
	// Output: true
}
