package database

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

// ConnectDB connects to the MySQL database
func ConnectDB() {
	var err error

	DB, err = sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/forum_posts")
	if err != nil {
		log.Fatal("Failed to connect to the database: ", err)
	}

	// Ping the database to ensure connection is established
	err = DB.Ping()
	if err != nil {
		log.Fatal("Failed to ping the database: ", err)
	}

	log.Println("Database connected")
}
