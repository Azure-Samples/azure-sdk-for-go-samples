package main

import (
	"flag"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/common"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/management"
	"github.com/Azure/go-autorest/autorest"
)

var (
	accountName string
)

func init() {
	management.GetStartParams()
	flag.StringVar(&accountName, "storageAccName", "storageaccname", "Provide a name for the storage account to be created")
	flag.Parse()
}

func main() {
	var err error
	var errC <-chan error

	group, err := management.CreateGroup()
	common.OnErrorFail(err, "failed to create group")
	log.Printf("group: %+v\n", group)

	sa, errC := management.CreateStorageAccount(accountName)
	common.OnErrorFail(<-errC, "failed to create storage account")
	log.Printf("storage account: %+v\n", <-sa)

	management.KeepResourcesAndExit()
	log.Printf("going to delete all resources\n")

	var res autorest.Response
	var resC <-chan autorest.Response

	res, err = management.DeleteStorageAccount(accountName)
	common.OnErrorFail(err, "failed to delete storage account")
	log.Printf("storage account deleted: %+v\n", res)

	resC, errC = management.DeleteGroup()
	common.OnErrorFail(<-errC, "failed to delete group")
	log.Printf("group deleted: %+v\n", <-resC)
}
