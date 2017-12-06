// tests basic functionality for an existing mssql db
package sql

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
)

const (
	connectionTimeout int = 300
	port              int = 1433
)

func DbOperations(server, database, username, password string) error {
	log.Printf("available drivers: %v", sql.Drivers())

	db, err := Open(server, database, username, password)
	if err != nil {
		return err
	}

	err = CreateTable(db)
	if err != nil {
		return err
	}

	err = Insert(db)
	if err != nil {
		return err
	}

	return Query(db)
}

func Open(server, database, username, password string) (*sql.DB, error) {
	query := url.Values{}
	query.Add("connection timeout", fmt.Sprintf("%d", connectionTimeout))
	query.Add("database", database)

	u := &url.URL{
		Scheme: "sqlserver",
		User:   url.UserPassword(username, password),
		Host:   fmt.Sprintf("%s.database.windows.net:%d", server, port),
		// Path:  instance, // if connecting to an instance instead of a port
		RawQuery: query.Encode(),
	}

	connectionString := u.String()

	log.Printf("using connString %s\n", connectionString)

	db, err := sql.Open("sqlserver", connectionString)
	if err != nil {
		fmt.Errorf("open connection failed: %v", err)
	}

	log.Printf("opened conn to %+v\n", db)
	return db, nil
}

func CreateTable(db *sql.DB) error {
	const createTableStatement string = `
    CREATE TABLE customers (
      id int NOT NULL PRIMARY KEY,
      name nvarchar(max)
    )`
	result, err := db.Exec(createTableStatement)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}
	rows, err := result.RowsAffected()
	log.Printf("table created, rows affected: %d\n", rows)
	return err
}

func Insert(db *sql.DB) error {
	const insertStmt string = `
    INSERT INTO customers VALUES (1, 'Josh')`
	result, err := db.Exec(insertStmt)
	if err != nil {
		return fmt.Errorf("failed to insert record: %v", err)
	}
	rows, err := result.RowsAffected()
	log.Printf("rows inserted: %d\n", rows)
	return err
}

func Query(db *sql.DB) error {
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
			fmt.Errorf("query failed:", err.Error())
		}

		log.Printf("  id: %d\n  name: %s\n", id, name)
	}

	return rows.Err()
}
