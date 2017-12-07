package storage

import (
	"flag"
	"strings"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/common"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/management"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
	"github.com/marstr/randname"
)

var (
	accountName   string
	containerName string
	blobName      string
)

func init() {
	management.GetStartParams()
	flag.StringVar(&accountName, "storageAccName", "acc"+randname.AdjNoun{}.Generate(), "Provide a name for the storage account to be created")
	flag.StringVar(&containerName, "containerName", "cnt"+randname.AdjNoun{}.Generate(), "Provide a name for the storage account to be created")
	flag.StringVar(&blobName, "blobName", "blob"+randname.AdjNoun{}.Generate(), "Provide a name for the storage account to be created")
	flag.Parse()
}

// Example creates a resource group and a storage account. Then it adds a container and a blob in that account.
// Finally it removes the blob, container, account, and group.
// more examples available at https://github.com/Azure/azure-storage-blob-go/2016-05-31/azblob/zt_examples_test.go
func ExampleUploadBlockBlob() {
	defer resources.Cleanup()

	_, err := resources.CreateGroup()
	if err != nil {
		common.PrintAndLog(err.Error())
	}
	common.PrintAndLog("resource group created")

	accountName = strings.ToLower(accountName)
	containerName = strings.ToLower(containerName)

	_, errC := CreateStorageAccount(accountName)
	err = <-errC
	if err != nil {
		common.PrintAndLog(err.Error())
	}
	common.PrintAndLog("created storage account")

	_, err = CreateContainer(accountName, containerName)
	if err != nil {
		common.PrintAndLog(err.Error())
	}
	common.PrintAndLog("created container")

	_, err = CreateBlockBlob(accountName, containerName, blobName)
	if err != nil {
		common.PrintAndLog(err.Error())
	}
	common.PrintAndLog("created blob")

	// Output:
	// resource group created
	// created storage account
	// created container
	// created blob
}
