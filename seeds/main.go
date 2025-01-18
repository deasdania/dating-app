package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

const (
	USER_ID_TEST = "d5dba9f0-2daf-47d3-9763-f98b1ea25376"
)

func main() {
	// Connect to your PostgreSQL database (make sure you have the correct connection details)
	// connStr := "postgres://username:password@localhost:5432/database_name?sslmode=disable"
	connStr := "postgres://postgres:secret@localhost:5432/dating_app?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Seed data for users
	for i := 1; i <= 50; i++ {
		username := fmt.Sprintf("user%d", i)
		email := fmt.Sprintf("%s@example.com", username)
		password := fmt.Sprintf("password%d", i)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal(err)
		}

		isPremium := rand.Intn(2) == 1
		isVerified := rand.Intn(2) == 1
		// Insert user into the "users" table
		_, err = db.Exec(`
			INSERT INTO users (id, username, email, password, is_premium, verified)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			uuid.New(), username, email, hashedPassword, isPremium, isVerified)

		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Inserted user: %s\n", username)
	}
	// e2e needed
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Password123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`
			INSERT INTO users (id, username, email, password, is_premium, verified)
			VALUES ($1, $2, $3, $4, $5, $6)`,
		USER_ID_TEST, "testuser", "testuser@example.com", hashedPassword, 0, 0)

	fmt.Println("Seed data insertion complete.")
}
