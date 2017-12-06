package storage

import (
	"fmt"
	"log"

	_ "github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

// Example creates a resource group and a storage account. Then it adds a container and a blob in that account.
// Finally it removes the blob, container, account, and group.
// more examples available at https://github.com/Azure/azure-storage-blob-go/2016-05-31/azblob/zt_examples_test.go
func Example() {
	var err error
	var errC <-chan error

	defer resources.Cleanup()

	group, err := resources.CreateGroup()
	if err != nil {
		log.Fatalf("failed to get create group: %v", err)
	}
	log.Printf("created group: %v\n", group)

	account, errC := CreateStorageAccount()
	err = <-errC // wait on error channel
	if err != nil {
		log.Fatalf("failed to create storage account: %v", err)
	}
	log.Printf("created storage account: %v\n", <-account)

	c, err := CreateContainer(containerName)
	if err != nil {
		log.Fatalf("failed to create container: %v", err)
	}
	log.Printf("created container: %v", c)

	b, err := CreateBlockBlob(blobName)
	if err != nil {
		log.Fatalf("failed to create blob: %v", err)
	}
	log.Printf("created blob: %v", b)

	fmt.Println("Success")
	// Output: Success
}
