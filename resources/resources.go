package resources

import (
	"log"
	"os"
)

// Cleanup deletes the current resource group if env var AZURE_KEEP_SAMPLE_RESOURCES is unset
func Cleanup() error {
	if os.Getenv("AZURE_KEEP_SAMPLE_RESOURCES") == "1" {
		log.Printf("retaining resources because env var is set\n")
		os.Exit(0)
	}

	_, errC := DeleteGroup()
	err := <-errC
	return err
}
