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
	// Підключення до бази даних
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://busoptima_user:busoptima_pass@localhost:5432/busoptima?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Тестування підключення
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("Connected to database successfully")

	// Регенерація хешів паролів користувачів
	if err := regenerateUserHashes(db); err != nil {
		log.Fatal("Failed to regenerate user hashes:", err)
	}

	// Регенерація хешів токенів пристроїв
	if err := regenerateDeviceHashes(db); err != nil {
		log.Fatal("Failed to regenerate device hashes:", err)
	}

	fmt.Println("All hashes regenerated successfully!")
}

func regenerateUserHashes(db *sql.DB) error {
	fmt.Println("Regenerating user password hashes...")

	// Пароль за замовчуванням для всіх тестових користувачів
	defaultPassword := "password123"

	// Генерація нового хешу
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to generate password hash: %w", err)
	}

	// Оновлення всіх користувачів новим хешем
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

	// Токен за замовчуванням для всіх тестових пристроїв
	defaultToken := "device_token_123"

	// Генерація нового хешу
	hashedToken, err := bcrypt.GenerateFromPassword([]byte(defaultToken), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to generate token hash: %w", err)
	}

	// Оновлення всіх пристроїв новим хешем
	query := `UPDATE devices SET auth_token_hash = $1`
	result, err := db.Exec(query, string(hashedToken))
	if err != nil {
		return fmt.Errorf("failed to update device tokens: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("Updated %d device token hashes\n", rowsAffected)

	return nil
}
