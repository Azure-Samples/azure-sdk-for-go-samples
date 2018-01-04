package storage

import (
	"context"
	"fmt"
	"net/url"

	blob "github.com/Azure/azure-storage-blob-go/2016-05-31/azblob"
)

var (
	blobFormatString = `https://%s.blob.core.windows.net`
)

func getContainerURL(ctx context.Context, accountName, containerName string) blob.ContainerURL {
	key := getFirstKey(ctx, accountName)
	c := blob.NewSharedKeyCredential(accountName, key)
	p := blob.NewPipeline(c, blob.PipelineOptions{})
	u, _ := url.Parse(fmt.Sprintf(blobFormatString, accountName))
	service := blob.NewServiceURL(*u, p)
	container := service.NewContainerURL(containerName)
	return container
}

// CreateContainer creates a new container with the specified name in the specified account
func CreateContainer(ctx context.Context, accountName, containerName string) (blob.ContainerURL, error) {
	c := getContainerURL(ctx, accountName, containerName)

	_, err := c.Create(
		context.Background(),
		blob.Metadata{},
		blob.PublicAccessContainer)
	return c, err
}

// GetContainer gets info about an existing container.
func GetContainer(ctx context.Context, accountName, containerName string) (blob.ContainerURL, error) {
	c := getContainerURL(ctx, accountName, containerName)

	_, err := c.GetPropertiesAndMetadata(context.Background(), blob.LeaseAccessConditions{})
	return c, err
}

// DeleteContainer deletes the named container.
func DeleteContainer(ctx context.Context, accountName, containerName string) error {
	c := getContainerURL(ctx, containerName, containerName)

	_, err := c.Delete(context.Background(), blob.ContainerAccessConditions{})
	return err
}
