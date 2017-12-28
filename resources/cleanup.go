package resources

import (
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
)

func Cleanup() error {
	if helpers.KeepResources() {
		log.Println("keeping resources")
		return nil
	}
	log.Println("deleting resources")
	_, error := DeleteGroup(helpers.ResourceGroupName())
	return error
}
