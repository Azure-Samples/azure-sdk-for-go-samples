package storage

import (
	"flag"
	"log"
	"strings"
	"testing"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/examples/resources"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/management"
	"github.com/marstr/randname"
	chk "gopkg.in/check.v1"
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

func Test(t *testing.T) { chk.TestingT(t) }

type StorageSuite struct{}

var _ = chk.Suite(&StorageSuite{})

// Example creates a resource group and a storage account. Then it adds a container and a blob in that account.
// Finally it removes the blob, container, account, and group.
// more examples available at https://github.com/Azure/azure-storage-blob-go/2016-05-31/azblob/zt_examples_test.go
func (s *StorageSuite) TestUploadBlockBlob(c *chk.C) {
	defer resources.Cleanup()

	group, err := resources.CreateGroup()
	c.Check(err, chk.IsNil)
	log.Printf("group: %+v\n", group)

	accountName = strings.ToLower(accountName)
	containerName = strings.ToLower(containerName)

	account, errC := CreateStorageAccount(accountName)
	c.Check(<-errC, chk.IsNil)
	log.Printf("created storage account: %v\n", <-account)

	cnt, err := CreateContainer(accountName, containerName)
	c.Check(err, chk.IsNil)
	log.Printf("created container: %v", cnt)

	b, err := CreateBlockBlob(accountName, containerName, blobName)
	c.Check(err, chk.IsNil)
	log.Printf("created blob: %v", b)
}
