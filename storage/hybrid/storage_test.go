package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	hybridresources "github.com/Azure-Samples/azure-sdk-for-go-samples/resources/hybrid"
)

var (
	accountName = strings.ToLower("storageaccount" + helpers.GetRandomLetterSequence(10))
)

func TestMain(m *testing.M) {
	err := helpers.ParseArgs()
	if err != nil {
		log.Fatalln("failed to parse args")
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
