package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/marstr/randname"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	hybridresources "github.com/Azure-Samples/azure-sdk-for-go-samples/resources/hybrid"
)

var (
	accountName = randname.Prefixed{Prefix: "storageaccount", Len: 10, Acceptable: randname.LowercaseAlphabet}.Generate()
)

func TestMain(m *testing.M) {
	err := iam.ParseArgs()
	if err != nil {
		log.Fatalln("failed to parse IAM args")
	}

	os.Exit(m.Run())
}

func ExampleCreateStorageAccount() {
	ctx := context.Background()
	defer hybridresources.Cleanup(ctx)

	_, err := hybridresources.CreateGroup(ctx)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	_, err = CreateStorageAccount(context.Background(), accountName)
	if err != nil {
		log.Fatal(fmt.Sprintf("Cannot create storage account. Error details: %s", err.Error()))
	}
	fmt.Println("Storage account created")

	// Output:
	// Storage account created
}
