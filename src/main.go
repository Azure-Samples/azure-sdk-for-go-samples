package main

import (
  "fmt"
  "github.com/joshgav/az-go/src/util"
  "github.com/joshgav/az-go/src/resources"
  "github.com/joshgav/az-go/src/mssql"
)

func main() {
  group, err := resources.CreateGroup()
  util.OnErrorFail(err, "failed to create group")
  fmt.Printf("group: %+v\n", group)

  server, err2 := resources.CreateServer()
  util.OnErrorFail(<-err2, "failed to create server")
  fmt.Printf("new server created: %+v\n", <-server)

  db, err3 := resources.CreateDb()
  util.OnErrorFail(<-err3, "failed to create database")
  fmt.Printf("new database created: %+v\n", <-db)

  fwRule, err4 := resources.OpenDbPort()
  util.OnErrorFail(err4, "failed to open db port")
  fmt.Printf("db fw rule set: %+v\n", fwRule)

  mssql.TestDb()

  // comment out the following stanzas to retain your db
  _, err5 := resources.DeleteDb()
  util.OnErrorFail(err5, "failed to delete database")
  fmt.Println("database deleted")

  err6 := resources.DeleteGroup()
  util.OnErrorFail(err6, "failed to delete group")
  fmt.Println("group deleted")
}
