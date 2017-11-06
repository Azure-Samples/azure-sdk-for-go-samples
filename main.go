package main

import (
	"github.com/joshgav/az-go/common"
	"github.com/joshgav/az-go/mssql"
	"github.com/joshgav/az-go/resources"
	"log"
)

func main() {
	var err error
	var errC <-chan error

	group, err := resources.CreateGroup()
	common.OnErrorFail(err, "failed to create group")
	log.Printf("group: %+v\n", group)

	server, errC := resources.CreateServer()
	common.OnErrorFail(<-errC, "failed to create server")
	log.Printf("new server created: %+v\n", <-server)

	db, errC := resources.CreateDb()
	common.OnErrorFail(<-errC, "failed to create database")
	log.Printf("new database created: %+v\n", <-db)

	fwRule, err := resources.OpenDbPort()
	common.OnErrorFail(err, "failed to open db port")
	log.Printf("db fw rule set: %+v\n", fwRule)

	mssql.TestDb()

	// comment out the following stanzas to retain your db
	_, err = resources.DeleteDb()
	common.OnErrorFail(err, "failed to delete database")
	log.Printf("database deleted\n")

	err = resources.DeleteGroup()
	common.OnErrorFail(err, "failed to delete group")
	log.Printf("group deleted\n")
}
