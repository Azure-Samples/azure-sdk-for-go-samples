package storage

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/marstr/randname"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	hybridresources "github.com/Azure-Samples/azure-sdk-for-go-samples/resources/hybrid"
)

var (
	accountName = randname.Prefixed{Prefix: "storageaccount", Len: 10, Acceptable: randname.LowercaseAlphabet}.Generate()
)

func TestMain(m *testing.M) {
	if err := config.ParseEnvironment(); err != nil {
		log.Fatalf("failed to parse env: %+v", err)
	}
	if err := config.AddFlags(); err != nil {
		log.Fatalf("failed to add flags: %+v", err)
	}
	flag.Parse()

	os.Exit(m.Run())
}

func ExampleCreateStorageAccount() {
	ctx := context.Background()
	defer hybridresources.Cleanup(ctx)

	_, err := hybridresources.CreateGroup(ctx)
	if err != nil {
		util.LogAndPanic(err)
	}
	_, err = CreateStorageAccount(context.Background(), accountName)
	if err != nil {
		util.LogAndPanic(fmt.Errorf("cannot create storage account. Error details: %+v", err))
	}
	fmt.Println("Storage account created")

	// Output:
	// Storage account created
}
