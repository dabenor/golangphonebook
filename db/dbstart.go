package db

import (
	"fmt"
	"golangphonebook/internal"
	"golangphonebook/pkg/contacts"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func DBInit() (*gorm.DB, error) {
	// Set up PostgreSQL connection
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})

	if err != nil {
		internal.Logger.Error(fmt.Sprintf("Failed to connect to DB with error: %v", err))
		return nil, err
	}
	db.Logger.LogMode(logger.Info)
	err = db.AutoMigrate(&contacts.Contact{})
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("Error migrating schema: %v\n", err))
	}

	internal.Logger.Info("Database schema created successfully.")

	return db, nil

}
