// Package storage предоставляет функциональность хранения данных для банковского приложения
package storage

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver

	"bankapp/internal/config"
)

// DBStorage представляет хранилище данных в PostgreSQL
type DBStorage struct {
	DB *sqlx.DB // Соединение с базой данных
}

var db *DBStorage

// InitStorage создает и инициализирует соединение с базой данных PostgreSQL
// Должна быть вызвана перед использованием любых других функций хранилища
func InitStorage() error {
	// Получаем конфигурацию базы данных
	dbConfig := config.GetDBConfig()

	// Устанавливаем соединение с базой данных
	dbConn, err := sqlx.Connect("postgres", dbConfig.ConnectionString())
	if err != nil {
		return fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}

	// Проверяем соединение
	if err := dbConn.Ping(); err != nil {
		return fmt.Errorf("не удалось проверить соединение с базой данных: %w", err)
	}

	// Создаем экземпляр хранилища
	db = &DBStorage{
		DB: dbConn,
	}

	// Инициализируем схему базы данных
	if err := initDBSchema(dbConn); err != nil {
		return fmt.Errorf("не удалось инициализировать схему базы данных: %w", err)
	}

	log.Println("Соединение с базой данных PostgreSQL установлено успешно")
	return nil
}

// CloseStorage закрывает соединение с базой данных
func CloseStorage() error {
	if db != nil && db.DB != nil {
		return db.DB.Close()
	}
	return nil
}

// initDBSchema создает необходимые таблицы в базе данных, если они не существуют
func initDBSchema(db *sqlx.DB) error {
	// SQL-запросы для создания таблиц
	schema := `
	-- Таблица пользователей
	CREATE TABLE IF NOT EXISTS users (
		id VARCHAR(36) PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		password_hash VARCHAR(100) NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	-- Таблица счетов
	CREATE TABLE IF NOT EXISTS accounts (
		id VARCHAR(36) PRIMARY KEY,
		user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		number VARCHAR(20) UNIQUE NOT NULL,
		balance DECIMAL(15, 2) NOT NULL DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	-- Таблица карт
	CREATE TABLE IF NOT EXISTS cards (
		id VARCHAR(36) PRIMARY KEY,
		account_id VARCHAR(36) NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
		number VARCHAR(16) NOT NULL,
		encrypted_number TEXT NOT NULL,
		number_hmac TEXT NOT NULL,
		expiry_month INTEGER NOT NULL,
		expiry_year INTEGER NOT NULL,
		cvv_hash TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	-- Таблица транзакций
	CREATE TABLE IF NOT EXISTS transactions (
		id VARCHAR(36) PRIMARY KEY,
		from_account_id VARCHAR(36) REFERENCES accounts(id),
		to_account_id VARCHAR(36) REFERENCES accounts(id),
		amount DECIMAL(15, 2) NOT NULL,
		timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		transaction_type VARCHAR(50) NOT NULL,
		description TEXT
	);

	-- Таблица кредитов
	CREATE TABLE IF NOT EXISTS credits (
		id VARCHAR(36) PRIMARY KEY,
		user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		account_id VARCHAR(36) NOT NULL REFERENCES accounts(id),
		amount DECIMAL(15, 2) NOT NULL,
		interest_rate DECIMAL(5, 2) NOT NULL,
		term_months INTEGER NOT NULL,
		start_date TIMESTAMP NOT NULL,
		remaining_amount DECIMAL(15, 2) NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	-- Таблица графиков платежей
	CREATE TABLE IF NOT EXISTS payment_schedules (
		id SERIAL PRIMARY KEY,
		credit_id VARCHAR(36) NOT NULL REFERENCES credits(id) ON DELETE CASCADE,
		due_date TIMESTAMP NOT NULL,
		amount DECIMAL(15, 2) NOT NULL,
		principal_part DECIMAL(15, 2) NOT NULL,
		interest_part DECIMAL(15, 2) NOT NULL,
		paid BOOLEAN NOT NULL DEFAULT FALSE
	);
	`

	// Выполняем SQL-запросы для создания таблиц
	_, err := db.Exec(schema)
	return err
}
