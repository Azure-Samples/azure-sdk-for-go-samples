package postgresqlsamples

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"

	// Postgresql driver.
	_ "github.com/lib/pq"
)

const (
	connectionTimeout int = 300
	port              int = 5432
)

// Open opens a connection to the MySQL server
func Open(server, database, username, password string) (*sql.DB, error) {
	query := url.Values{}
	query.Add("connection timeout", fmt.Sprintf("%d", connectionTimeout))
	query.Add("allowNativePasswords", "true")
	query.Add("database", database)

	u := &url.URL{
		Scheme: "postgresql",
		User:   url.UserPassword(username, password),
		Host:   fmt.Sprintf("%s.postgresql.database.azure.com:%d", server, port),
		// Path:  instance, // if connecting to an instance instead of a port
		RawQuery: query.Encode(),
	}

	connectionString := u.String()

	log.Printf("using connString %s\n", connectionString)

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return db, fmt.Errorf("open connection failed: %v", err)
	}

	log.Printf("opened conn to %+v\n", db)
	return db, nil
}

// CreateTable creates an SQL table
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

// Insert adds a row to the SQL datablase
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

// Query queries the SQL database
func Query(db *sql.DB) error {
	// assert(db != null)
	const queryString string = "SELECT id,name FROM customers"
	log.Printf("using query %s\n", queryString)

	rows, err := db.Query(queryString)
	if err != nil {
		log.Fatalf("query failed: %+v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		err := rows.Scan(&id, &name)
		if err != nil {
			return fmt.Errorf("query failed: %+v", err)
		}

		log.Printf("  id: %d\n  name: %s\n", id, name)
	}

	return rows.Err()
}
