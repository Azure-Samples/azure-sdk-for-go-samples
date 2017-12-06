// tests basic functionality for an existing mssql db
package mssql

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"database/sql"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	_ "github.com/denisenkom/go-mssqldb"
)

const (
	connectionTimeout int = 300
	port              int = 1433
)

var db *sql.DB

func TestDb() {
	log.Printf("available drivers: %v", sql.Drivers())

	err := open()
	if err != nil {
		log.Fatal("open connection failed:", err.Error())
	}

	err = createTable()
	if err != nil {
		log.Fatal("create table failed:", err.Error())
	}

	err = insert()
	if err != nil {
		log.Fatal("insert failed:", err.Error())
	}

	err = query()
	if err != nil {
		log.Fatal("query failed:", err.Error())
	}
}

func open() error {
	query := url.Values{}
	query.Add("connection timeout", fmt.Sprintf("%d", connectionTimeout))
	query.Add("database", dbName)

	var _serverName string
	if !strings.ContainsRune(serverName, '.') {
		_serverName = serverName + ".database.windows.net"
	} else {
		_serverName = serverName
	}

	u := &url.URL{
		Scheme: "sqlserver",
		User:   url.UserPassword(dbLogin, dbPassword),
		Host:   fmt.Sprintf("%s:%d", _serverName, port),
		// Path:  instance, // if connecting to an instance instead of a port
		RawQuery: query.Encode(),
	}

	connectionString := u.String()

	log.Printf("using connString %s\n", connectionString)

	_db, err := sql.Open("sqlserver", connectionString)

	if err != nil {
		log.Fatal("open connection failed:", err.Error())
	}
	db = _db

	log.Printf("opened conn to %+v\n", db)
	return nil
}

func createTable() error {
	const createTableStatement string = `
    CREATE TABLE customers (
      id int NOT NULL PRIMARY KEY,
      name nvarchar(max)
    )`
	result, err := db.Exec(createTableStatement)
	helpers.OnErrorFail(err, "failed to create table")
	rows, err := result.RowsAffected()
	log.Printf("table created, rows affected: %d\n", rows)
	return err
}

func insert() error {
	const insertStmt string = `
    INSERT INTO customers VALUES (1, 'Josh')`
	result, err := db.Exec(insertStmt)
	helpers.OnErrorFail(err, "failed to insert record")
	rows, err := result.RowsAffected()
	log.Printf("rows inserted: %d\n", rows)
	return err
}

func query() error {
	// assert(db != null)
	const queryString string = "SELECT id,name FROM customers"
	log.Printf("using query %s\n", queryString)

	rows, err := db.Query(queryString)
	if err != nil {
		log.Fatal("query failed:", err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		err := rows.Scan(&id, &name)
		if err != nil {
			log.Print("query failed:", err.Error())
		}

		log.Printf("  id: %d\n  name: %s\n", id, name)
	}

	return rows.Err()
}
