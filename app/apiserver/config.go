package apiserver

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config ...
type Config struct {
	BindAddr    string `json:"bind_addr"`
	LogLevel    string `json:"log_level"`
	DatabaseURL string `json:"database_url"`
}

// NewConfig ...
func NewConfig() *Config {
	return &Config{
		BindAddr:    ":8080",
		LogLevel:    "debug",
		DatabaseURL: DatabaseURL(),
	}
}

// DatabaseURL ...
func DatabaseURL() string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	return fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable", dbUser, dbPassword, dbName, dbHost, dbPort)
}
