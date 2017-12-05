package storage

import (
	"context"
	"fmt"
	"net/url"
	"os"

	blob "github.com/Azure/azure-storage-blob-go/2016-05-31/azblob"
	"github.com/subosito/gotenv"
)

var (
	containerName    string
	blobFormatString = `https://%s.blob.core.windows.net`
)

func init() {
	gotenv.Load()
	containerName = os.Getenv("AZURE_STORAGE_CONTAINERNAME")
	// use a default name if not specified
	if len(containerName) < 1 {
		containerName = "mycontainer"
	}
}

func getContainerURL(name string) blob.ContainerURL {
	loadKey()
	c := blob.NewSharedKeyCredential(accountName, accountKey)
	p := blob.NewPipeline(c, blob.PipelineOptions{})
	u, _ := url.Parse(fmt.Sprintf(blobFormatString, accountName))
	service := blob.NewServiceURL(*u, p)
	container := service.NewContainerURL(name)
	return container
}

// CreateContainer creates a new container with the specified name
// in the Storage Account specified by env var
func CreateContainer(name string) (blob.ContainerURL, error) {
	c := getContainerURL(name)

	_, err := c.Create(
		context.Background(),
		blob.Metadata{},
		blob.PublicAccessContainer)
	return c, err
}

// GetContainer gets info about an existing container.
func GetContainer(name string) (blob.ContainerURL, error) {
	c := getContainerURL(name)

	_, err := c.GetPropertiesAndMetadata(context.Background(), blob.LeaseAccessConditions{})
	return c, err
}

// DeleteContainer deletes the named container.
func DeleteContainer(name string) error {
	c := getContainerURL(name)

	_, err := c.Delete(context.Background(), blob.ContainerAccessConditions{})
	return err
}
