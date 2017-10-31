package main

import (
  "log"
  "github.com/joshgav/az-go/util"
  "github.com/joshgav/az-go/resources"
  "github.com/joshgav/az-go/mssql"
)

func main() {
  var err error
  var errC <-chan error

  group, err := resources.CreateGroup()
  util.OnErrorFail(err, "failed to create group")
  log.Printf("group: %+v\n", group)

  server, errC := resources.CreateServer()
  util.OnErrorFail(<-errC, "failed to create server")
  log.Printf("new server created: %+v\n", <-server)

  db, errC := resources.CreateDb()
  util.OnErrorFail(<-errC, "failed to create database")
  log.Printf("new database created: %+v\n", <-db)

  fwRule, err := resources.OpenDbPort()
  util.OnErrorFail(err, "failed to open db port")
  log.Printf("db fw rule set: %+v\n", fwRule)

  mssql.TestDb()

  // comment out the following stanzas to retain your db
  _, err = resources.DeleteDb()
  util.OnErrorFail(err, "failed to delete database")
  log.Printf("database deleted\n")

  err = resources.DeleteGroup()
  util.OnErrorFail(err, "failed to delete group")
  log.Printf("group deleted\n")
}
