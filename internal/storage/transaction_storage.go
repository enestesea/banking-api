package storage

import (
	"fmt"
	"log"

	"bankapp/internal/models"
	"bankapp/pkg/utils"
)

// AddTransaction Добавляет новую транзакцию в базу данных
func AddTransaction(tx models.Transaction) error {
	query := `
		INSERT INTO transactions (id, from_account_id, to_account_id, amount, timestamp, transaction_type, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := db.DB.Exec(query,
		tx.ID,
		tx.FromAccountID,
		tx.ToAccountID,
		tx.Amount,
		tx.Timestamp,
		tx.TransactionType,
		tx.Description)

	if err != nil {
		return fmt.Errorf("ошибка при добавлении транзакции: %w", err)
	}

	log.Printf("Транзакция %s добавлена. Тип: %s, Сумма: %s", tx.ID, tx.TransactionType, tx.Amount.String())
	return nil
}

// GetAccountTransactions Получает все транзакции для счета
// Возвращает срез транзакций, где счет является либо источником, либо получателем
func GetAccountTransactions(accountID string) []models.Transaction {
	query := `
		SELECT id, from_account_id, to_account_id, amount, timestamp, transaction_type, description
		FROM transactions
		WHERE from_account_id = $1 OR to_account_id = $1
		ORDER BY timestamp DESC
	`
	rows, err := db.DB.Query(query, accountID)
	if err != nil {
		log.Printf("Ошибка при получении транзакций для счета: %v", err)
		return []models.Transaction{}
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		err := rows.Scan(
			&tx.ID,
			&tx.FromAccountID,
			&tx.ToAccountID,
			&tx.Amount,
			&tx.Timestamp,
			&tx.TransactionType,
			&tx.Description,
		)
		if err != nil {
			log.Printf("Ошибка при сканировании данных транзакции: %v", err)
			continue
		}
		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Ошибка при итерации по результатам запроса: %v", err)
	}

	return transactions
}

// GetAllTransactions Получает все транзакции из базы данных
// Возвращает срез всех транзакций
func GetAllTransactions() ([]models.Transaction, error) {
	query := `
		SELECT id, from_account_id, to_account_id, amount, timestamp, transaction_type, description
		FROM transactions
		ORDER BY timestamp DESC
	`
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении всех транзакций: %w", err)
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		err := rows.Scan(
			&tx.ID,
			&tx.FromAccountID,
			&tx.ToAccountID,
			&tx.Amount,
			&tx.Timestamp,
			&tx.TransactionType,
			&tx.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании данных транзакции: %w", err)
		}
		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по результатам запроса: %w", err)
	}

	return transactions, nil
}

// GenerateTransactionID генерирует уникальный ID для транзакции
// Использует функцию CreateUniqueIdentifier из пакета utils
func GenerateTransactionID() string {
	return utils.CreateUniqueIdentifier()
}
