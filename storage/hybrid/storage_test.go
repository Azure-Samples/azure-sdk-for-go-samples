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
	err := config.ParseEnvironment()
	if err != nil {
		log.Fatalln("failed to parse env")
	}

	config.AddFlags()
	flag.Parse()
	if err != nil {
		log.Fatalln("failed to parse flags")
	}

	os.Exit(m.Run())
}

func ExampleCreateStorageAccount() {
	ctx := context.Background()
	defer hybridresources.Cleanup(ctx)

	_, err := hybridresources.CreateGroup(ctx)
	if err != nil {
		util.PrintAndLog(err.Error())
	}
	_, err = CreateStorageAccount(context.Background(), accountName)
	if err != nil {
		log.Fatal(fmt.Sprintf("Cannot create storage account. Error details: %s", err.Error()))
	}
	fmt.Println("Storage account created")

	// Output:
	// Storage account created
}
