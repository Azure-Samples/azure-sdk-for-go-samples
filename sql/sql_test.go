package mssql

import (
	"fmt"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

// Example creates a SQL server and database, then creates a table and inserts a record.
func Example() {
	var err error
	var errC <-chan error

	defer resources.Cleanup()

	group, err := resources.CreateGroup()
	helpers.OnErrorFail(err, "failed to create group")
	log.Printf("group: %+v\n", group)

	server, errC := CreateServer()
	helpers.OnErrorFail(<-errC, "failed to create server")
	log.Printf("new server created: %+v\n", <-server)

	db, errC := CreateDb()
	helpers.OnErrorFail(<-errC, "failed to create database")
	log.Printf("new database created: %+v\n", <-db)

	err = OpenDbPort()
	helpers.OnErrorFail(err, "failed to open db port")
	log.Printf("db fw rules set\n")

	TestDb()
	fmt.Println("Success")
	// Output: Success
}
