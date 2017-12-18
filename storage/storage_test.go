package storage

import (
	"flag"
	"os"
	"strings"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
	"github.com/subosito/gotenv"
)

var (
	accountName   = "azuresamplesgo" + helpers.GetRandomLetterSequence(10)
	containerName = "container1"
	blobName      = "blob1"
)

func init() {
	gotenv.Load()
	name := os.Getenv("AZ_STORAGE_ACCOUNT_NAME")
	if len(name) > 0 {
		accountName = name
	}

	flag.StringVar(&accountName, "storageAccoutName", accountName, "Provide a name for the storage account to be created")
	flag.StringVar(&containerName, "containerName", containerName, "Provide a name for the container.")
	flag.StringVar(&blobName, "blobName", blobName, "Provide a name for the blob.")
	helpers.ParseArgs()
}

// Example creates a resource group and a storage account. Then it adds a container and a blob in that account.
// Finally it removes the blob, container, account, and group.
// more examples available at https://github.com/Azure/azure-storage-blob-go/2016-05-31/azblob/zt_examples_test.go
func ExampleUploadBlockBlob() {
	defer resources.Cleanup()

	_, err := resources.CreateGroup(helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("resource group created")

	accountName = strings.ToLower(accountName)
	containerName = strings.ToLower(containerName)

	_, errC := CreateStorageAccount(accountName)
	err = <-errC
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created storage account")

	_, err = CreateContainer(accountName, containerName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created container")

	_, err = CreateBlockBlob(accountName, containerName, blobName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created blob")

	// Output:
	// resource group created
	// created storage account
	// created container
	// created blob
}
