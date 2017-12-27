package resources

import (
	"context"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
)

func Cleanup(ctx context.Context) error {
	if helpers.KeepResources() {
		log.Println("keeping resources")
		return nil
	}
	log.Println("deleting resources")
	_, err := DeleteGroup(ctx, helpers.ResourceGroupName())
	return err
}
