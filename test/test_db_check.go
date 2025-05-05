package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	connStr := "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed to connect to DB: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, username FROM users")
	if err != nil {
		log.Fatalf("failed to query users: %v", err)
	}
	defer rows.Close()

	fmt.Println("📄 Users in the database:")
	for rows.Next() {
		var id int
		var username string
		if err := rows.Scan(&id, &username); err != nil {
			log.Printf("error reading row: %v", err)
			continue
		}
		fmt.Printf("👤 ID: %d, Username: %s\n", id, username)
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("row error: %v", err)
	}
}
