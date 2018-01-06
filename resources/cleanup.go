package resources

import (
	"context"
	"fmt"
	"log"
	"sync"

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

// CleanupAll deletes all rescource groups
func CleanupAll(ctx context.Context, wg *sync.WaitGroup) {
	for list, err := ListGroups(ctx); list.NotDone(); err = list.Next() {
		if err != nil {
			log.Fatalf("got error: %s", err)
		}
		wg.Add(1)
		rgName := *list.Value().Name
		go func(ctx context.Context, rgName string) {
			fmt.Printf("deleting group '%s'\n", rgName)
			future, err := DeleteGroup(ctx, rgName)
			if err != nil {
				log.Fatalf("got error: %s", err)
			}
			err = future.WaitForCompletion(ctx, getGroupsClient().Client)
			if err != nil {
				log.Fatalf("got error: %s", err)
			} else {
				fmt.Printf("finished deleting group '%s'\n", rgName)
			}
			wg.Done()
		}(ctx, rgName)
	}
}
