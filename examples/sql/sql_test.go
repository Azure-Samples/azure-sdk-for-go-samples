package sql

import (
	"flag"
	"log"
	"testing"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/examples/resources"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/management"
	chk "gopkg.in/check.v1"
)

func Test(t *testing.T) { chk.TestingT(t) }

type SQLSuite struct{}

var _ = chk.Suite(&SQLSuite{})

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

// Example creates a SQL server and database, then creates a table and inserts a record.
func (s *SQLSuite) TestDatabaseQueries(c *chk.C) {
	defer resources.Cleanup()

	group, err := resources.CreateGroup()
	c.Check(err, chk.IsNil)
	log.Printf("group: %+v\n", group)

	server, errC := CreateServer(serverName, dbLogin, dbPassword)
	c.Check(<-errC, chk.IsNil)
	log.Printf("new server created: %+v\n", <-server)

	db, errC := CreateDb(serverName, dbName)
	c.Check(<-errC, chk.IsNil)
	log.Printf("new database created: %+v\n", <-db)

	err = CreateFirewallRules(serverName)
	c.Check(err, chk.IsNil)
	log.Printf("db fw rules set\n")

	err = DbOperations(serverName, dbName, dbLogin, dbPassword)
	c.Check(err, chk.IsNil)
	log.Printf("db operations done\n")
}
