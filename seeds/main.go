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
		dailySwipeCount := rand.Intn(11) // Random daily swipe count between 0 and 10

		// Insert user into the "users" table
		_, err = db.Exec(`
			INSERT INTO users (id, username, email, password, is_premium, verified, daily_swipe_count)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			uuid.New(), username, email, hashedPassword, isPremium, isVerified, dailySwipeCount)

		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Inserted user: %s\n", username)
	}

	// Seed data for profiles
	for i := 1; i <= 50; i++ {
		profileUsername := fmt.Sprintf("profile%d", i)
		description := fmt.Sprintf("This is the description for profile %d.", i)
		imageURL := fmt.Sprintf("https://example.com/images/profile%d.jpg", i)

		// Insert profile into the "profiles" table
		_, err = db.Exec(`
			INSERT INTO profiles (id, username, description, image_url)
			VALUES ($1, $2, $3, $4)`,
			uuid.New(), profileUsername, description, imageURL)

		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Inserted profile: %s\n", profileUsername)
	}

	fmt.Println("Seed data insertion complete.")
}
