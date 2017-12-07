package keyvault

import (
	"fmt"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

// Example creates a resource group and a storage account. Then it adds a container and a blob in that account.
// Finally it removes the blob, container, account, and group
// more examples available at https://github.com/Azure/azure-storage-blob-go/2016-05-31/azblob/zt_examples_test.go
func Example() {
	var err error

	defer resources.Cleanup()

	group, err := resources.CreateGroup()
	if err != nil {
		log.Fatalf("failed to get create group: %v", err)
	}
	log.Printf("created group: %v\n", group)

	v, err := CreateVault()
	if err != nil {
		log.Fatalf("failed to create vault: %v", err)
	}
	log.Printf("created vault: %v\n", v)

	v, err = SetVaultPermissions()
	if err != nil {
		log.Fatalf("failed to set vault permissions: %v", err)
	}
	log.Printf("set vault permissions: %v\n", v)

	fmt.Println("Success")
	// Output: Success
}
