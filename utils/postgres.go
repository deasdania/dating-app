package utils

import (
	"fmt"
	"log"

	"github.com/jinzhu/gorm"
)

type Database struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}

// Initialize the connection to PostgreSQL
func InitDB(db Database) *gorm.DB {
	connStr := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		db.Host, db.Port, db.User, db.Name, db.Password)

	// Connect to PostgreSQL
	dbg, err := gorm.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}
	return dbg
}
