package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "../yoga.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

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
            name TEXT NOT NULL,
            email TEXT NOT NULL,
            phone TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (workshop_id) REFERENCES workshops(id)
        );
    `)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("✓ Database tables created!")

	// Add a sample workshop
	_, err = db.Exec(`
        INSERT INTO workshops (title, description, date, location, max_capacity) 
        VALUES (?, ?, ?, ?, ?)
    `,
		"Sound Healing & Restorative Yoga",
		"Join us for a transformative evening of sound healing and gentle yoga. Experience deep relaxation through the resonant tones of crystal singing bowls, gongs, and guided meditation.",
		"Saturday, February 15, 2025 at 6:00 PM",
		"Peaceful Studio, Downtown",
		20,
	)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("✓ Sample workshop added!")
	fmt.Println("\nYou can now run: go run . (from the parent directory)")
}
