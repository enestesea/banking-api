package config

import (
	"fmt"
	"os"
)

// DBConfig holds the database connection configuration
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// GetDBConfig returns the database configuration from environment variables
// or default values if environment variables are not set
func GetDBConfig() DBConfig {
	return DBConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "admin"),
		DBName:   getEnv("DB_NAME", "banking"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

// ConnectionString returns the PostgreSQL connection string
func (c DBConfig) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// Helper function to get environment variable or default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
