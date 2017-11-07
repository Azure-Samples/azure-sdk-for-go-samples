package main

import (
	"log"
	"os"
	"github.com/joshgav/az-go/common"
	"github.com/joshgav/az-go/management"
//"github.com/joshgav/az-go/storage"
  "github.com/Azure/go-autorest/autorest"
)

func main() {
	var err error
	var errC <-chan error

	group, err := management.CreateGroup()
	common.OnErrorFail(err, "failed to create group")
	log.Printf("group: %+v\n", group)

  sa, errC := management.CreateStorageAccount()
  common.OnErrorFail(<-errC, "failed to create storage account")
  log.Printf("storage account: %+v\n", <-sa)

	if os.Getenv("AZURE_KEEP_SAMPLE_RESOURCES") == "1" {
    log.Printf("retaining resources because env var is set\n")
		os.Exit(0)
	}

//storage.TestStorage()

	log.Printf("going to delete all resources\n")

  var res autorest.Response
  var resC <-chan autorest.Response

  res, err = management.DeleteStorageAccount()
  common.OnErrorFail(err, "failed to delete storage account")
  log.Printf("storage account deleted: %+v\n", res)

	resC, errC = management.DeleteGroup()
	common.OnErrorFail(<-errC, "failed to delete group")
	log.Printf("group deleted: %+v\n", <-resC)
}
