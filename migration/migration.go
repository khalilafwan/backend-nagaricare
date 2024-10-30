package migration

import (
	"backend-nagaricare/database"
	"log"
)

// Migrate runs the migration to create the forum_posts table
func Migrate() {
	db := database.DB

	// SQL statement to create the forum_posts table
	query := `
    CREATE TABLE IF NOT EXISTS forum_posts (
        id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
        title VARCHAR(255) NOT NULL,
        content TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    `

	// Execute the migration
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("Failed to run migration: ", err)
	}

	log.Println("Migration completed successfully")
}
