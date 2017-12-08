package storage

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	blob "github.com/Azure/azure-storage-blob-go/2016-05-31/azblob"
)

func getBlockBlobURL(accountName, containerName, blobName string) blob.BlockBlobURL {
	key := getFirstKey(accountName)
	c := blob.NewSharedKeyCredential(accountName, key)
	p := blob.NewPipeline(c, blob.PipelineOptions{})
	u, _ := url.Parse(fmt.Sprintf(blobFormatString, accountName))
	service := blob.NewServiceURL(*u, p)
	container := service.NewContainerURL(containerName)
	blob := container.NewBlockBlobURL(blobName)
	return blob
}

// CreateBlockBlob creates a new test blob in the container specified by env var
func CreateBlockBlob(accountName, containerName, blobName string) (blob.BlockBlobURL, error) {
	b := getBlockBlobURL(accountName, containerName, blobName)
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
