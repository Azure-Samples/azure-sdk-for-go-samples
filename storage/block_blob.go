package storage

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	blob "github.com/Azure/azure-storage-blob-go/2016-05-31/azblob"
	"github.com/subosito/gotenv"
)

var (
	blobPath string
	blobName string
)

func init() {
	gotenv.Load()
	blobName = os.Getenv("AZURE_STORAGE_BLOBNAME")
	if len(blobName) < 1 {
		blobName = "myblob"
	}

	blobPath = os.Getenv("AZURE_STORAGE_BLOBPATH")
	if len(blobPath) < 1 {
		blobPath = "./_testdata/blob.bin"
	}
}

func getBlockBlobURL(containerName string, blobName string) blob.BlockBlobURL {
	loadKey()
	c := blob.NewSharedKeyCredential(accountName, accountKey)
	p := blob.NewPipeline(c, blob.PipelineOptions{})
	u, _ := url.Parse(fmt.Sprintf(blobFormatString, accountName))
	service := blob.NewServiceURL(*u, p)
	container := service.NewContainerURL(containerName)
	blob := container.NewBlockBlobURL(blobName)
	return blob
}

// CreateBlockBlob creates a new test blob in the container specified by env var
func CreateBlockBlob(name string) (blob.BlockBlobURL, error) {
	if len(name) > 0 {
		blobName = name
	}
	b := getBlockBlobURL(containerName, blobName)
	data := "blob created by Azure-Samples, okay to delete!"

	_, err := b.PutBlob(
		context.Background(),
		strings.NewReader(data),
		blob.BlobHTTPHeaders{
			ContentType: "text/plain",
		},
		blob.Metadata{},
		blob.BlobAccessConditions{},
	)

	return b, err
}
