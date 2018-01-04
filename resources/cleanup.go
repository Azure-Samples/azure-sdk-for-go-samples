package resources

import (
	"context"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
)

// Cleanup deletes the rescource group created for the sample
func Cleanup(ctx context.Context) error {
	if helpers.KeepResources() {
		log.Println("keeping resources")
		return nil
	}
	log.Println("deleting resources")
	_, err := DeleteGroup(ctx, helpers.ResourceGroupName())
	return err
}
