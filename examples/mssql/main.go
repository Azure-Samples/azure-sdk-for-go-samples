package main

import (
	"flag"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/common"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/dataplane/mssql"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/management"
	"github.com/Azure/go-autorest/autorest"
)

var (
	serverName string
	dbName     string
	dbLogin    string
	dbPassword string
)

func init() {
	management.GetStartParams()
	flag.StringVar(&serverName, "sqlServerName", "sqlservername", "Provide a name for the SQL server name to be created")
	flag.StringVar(&dbName, "sqlDbName", "sqldbname", "Provide a name for the SQL data basename to be created")
	flag.StringVar(&dbLogin, "sqlDbUserName", "sqldbuser", "Provide a name for the SQL database username")
	flag.StringVar(&dbPassword, "sqlDbPassword", "Pa$$w0rd1975", "Provide a name for the SQL database password")
	flag.Parse()
}

func main() {
	var err error
	var errC <-chan error

	group, err := management.CreateGroup()
	common.OnErrorFail(err, "failed to create group")
	log.Printf("group: %+v\n", group)

	server, errC := management.CreateServer(serverName, dbLogin, dbPassword)
	common.OnErrorFail(<-errC, "failed to create server")
	log.Printf("new server created: %+v\n", <-server)

	db, errC := management.CreateDb(serverName, dbName)
	common.OnErrorFail(<-errC, "failed to create database")
	log.Printf("new database created: %+v\n", <-db)

	err = management.OpenDbPort(serverName)
	common.OnErrorFail(err, "failed to open db port")
	log.Printf("db fw rules set\n")

	mssql.DbOperations(serverName, dbName, dbLogin, dbPassword)

	management.KeepResourcesAndExit()
	log.Printf("going to delete all resources\n")

	var res autorest.Response
	var resC <-chan autorest.Response

	res, err = management.DeleteDb(serverName, dbName)
	common.OnErrorFail(err, "failed to delete database")
	log.Printf("database deleted: %+v\n", res)

	resC, errC = management.DeleteGroup()
	common.OnErrorFail(<-errC, "failed to delete group")
	log.Printf("group deleted: %+v\n", <-resC)
}
