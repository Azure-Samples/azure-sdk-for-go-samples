package main

import (
	"log"
	"os"
	"github.com/joshgav/az-go/common"
	"github.com/joshgav/az-go/management"
	"github.com/joshgav/az-go/mssql"
)

func main() {
	var err error
	var errC <-chan error

	group, err := management.CreateGroup()
	common.OnErrorFail(err, "failed to create group")
	log.Printf("group: %+v\n", group)

	server, errC := management.CreateServer()
	common.OnErrorFail(<-errC, "failed to create server")
	log.Printf("new server created: %+v\n", <-server)

	db, errC := management.CreateDb()
	common.OnErrorFail(<-errC, "failed to create database")
	log.Printf("new database created: %+v\n", <-db)

	err = management.OpenDbPort()
	common.OnErrorFail(err, "failed to open db port")
	log.Printf("db fw rules set\n")

	mssql.TestDb()

	if os.Getenv("AZURE_KEEP_SAMPLE_RESOURCES") == "1" {
    log.Printf("retaining resources because env var is set\n")
		os.Exit(0)
	}

	log.Printf("going to delete all resources\n")

	_, err = management.DeleteDb()
	common.OnErrorFail(err, "failed to delete database")
	log.Printf("database deleted\n")

	err = management.DeleteGroup()
	common.OnErrorFail(err, "failed to delete group")
	log.Printf("group deleted\n")
}
