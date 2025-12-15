package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://busoptima_user:busoptima_pass@localhost:5432/busoptima?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("Connected to database successfully")

	// Regenerate user password hashes
	if err := regenerateUserHashes(db); err != nil {
		log.Fatal("Failed to regenerate user hashes:", err)
	}

	// Regenerate device token hashes
	if err := regenerateDeviceHashes(db); err != nil {
		log.Fatal("Failed to regenerate device hashes:", err)
	}

	fmt.Println("All hashes regenerated successfully!")
}

func regenerateUserHashes(db *sql.DB) error {
	fmt.Println("Regenerating user password hashes...")

	// Default password for all test users
	defaultPassword := "password123"

	// Generate new hash
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to generate password hash: %w", err)
	}

	// Update all users with new hash
	query := `UPDATE users SET password_hash = $1`
	result, err := db.Exec(query, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("failed to update user passwords: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("Updated %d user password hashes\n", rowsAffected)

	return nil
}

func regenerateDeviceHashes(db *sql.DB) error {
	fmt.Println("Regenerating device token hashes...")

	// Default token for all test devices
	defaultToken := "device_token_123"

	// Generate new hash
	hashedToken, err := bcrypt.GenerateFromPassword([]byte(defaultToken), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to generate token hash: %w", err)
	}

	// Update all devices with new hash
	query := `UPDATE devices SET auth_token_hash = $1`
	result, err := db.Exec(query, string(hashedToken))
	if err != nil {
		return fmt.Errorf("failed to update device tokens: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("Updated %d device token hashes\n", rowsAffected)

	return nil
}
