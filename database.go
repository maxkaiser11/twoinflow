package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func initDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./yoga.db")
	if err != nil {
		log.Fatal(err)
	}

	// Create tables
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS workshops (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            title TEXT NOT NULL,
            description TEXT,
            date TEXT NOT NULL,
            location TEXT,
            max_capacity INTEGER DEFAULT 20
        );

        CREATE TABLE IF NOT EXISTS signups (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            workshop_id INTEGER,
            first_name TEXT NOT NULL,
            last_name TEXT NOT NULL,
            email TEXT NOT NULL,
            phone TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (workshop_id) REFERENCES workshops(id)
        );

        CREATE TABLE IF NOT EXISTS admin_users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username TEXT UNIQUE NOT NULL,
            password_hash TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );
    `)
	if err != nil {
		log.Fatal(err)
	}

	// Check if default admin exists, if not create one
	var count int
	db.QueryRow("SELECT COUNT(*) FROM admin_users").Scan(&count)
	if count == 0 {
		// Get default credentials from environment or use fallback
		defaultUsername := os.Getenv("DEFAULT_ADMIN_USERNAME")
		defaultPassword := os.Getenv("DEFAULT_ADMIN_PASSWORD")

		if defaultUsername == "" {
			defaultUsername = "admin"
		}
		if defaultPassword == "" {
			defaultPassword = "yoga2025"
			log.Println("⚠️  WARNING: Using default password. Set DEFAULT_ADMIN_PASSWORD in production!")
		}

		// Create default admin
		hash, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal(err)
		}

		_, err = db.Exec("INSERT INTO admin_users (username, password_hash) VALUES (?, ?)",
			defaultUsername, string(hash))
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("✓ Default admin user created (username: %s)\n", defaultUsername)
	}

	return db
}
