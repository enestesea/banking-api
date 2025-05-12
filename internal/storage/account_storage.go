package storage

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/shopspring/decimal"

	"bankapp/internal/models"
)

// CreateBankAccount создает новый банковский счет для пользователя
// Проверяет существование пользователя и добавляет счет в базу данных
// Возвращает ошибку, если пользователь не найден
func CreateBankAccount(account models.Account) error {
	// Проверяем, существует ли пользователь
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", account.UserID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("ошибка при проверке существования пользователя: %w", err)
	}
	if !exists {
		return fmt.Errorf("user with ID %s not found", account.UserID)
	}

	// Сохраняем счет в базу данных
	query := `
		INSERT INTO accounts (id, user_id, number, balance, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = db.DB.Exec(query, account.ID, account.UserID, account.Number, account.Balance, account.CreatedAt)
	if err != nil {
		return fmt.Errorf("ошибка при создании счета: %w", err)
	}

	log.Printf("Счет %s создан для пользователя %s", account.ID, account.UserID)
	return nil
}

// GetAccount получает счет по его ID
// Возвращает счет и булево значение, указывающее, найден ли счет
func GetAccount(accountID string) (models.Account, bool) {
	var account models.Account
	query := `
		SELECT id, user_id, number, balance, created_at
		FROM accounts
		WHERE id = $1
	`
	err := db.DB.QueryRow(query, accountID).Scan(
		&account.ID,
		&account.UserID,
		&account.Number,
		&account.Balance,
		&account.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.Account{}, false
		}
		log.Printf("Ошибка при получении счета по ID: %v", err)
		return models.Account{}, false
	}

	return account, true
}

// GetUserAccounts получает все счета пользователя
// Возвращает срез счетов
func GetUserAccounts(userID string) []models.Account {
	query := `
		SELECT id, user_id, number, balance, created_at
		FROM accounts
		WHERE user_id = $1
		ORDER BY created_at
	`
	rows, err := db.DB.Query(query, userID)
	if err != nil {
		log.Printf("Ошибка при получении счетов пользователя: %v", err)
		return []models.Account{}
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		var account models.Account
		err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.Number,
			&account.Balance,
			&account.CreatedAt,
		)
		if err != nil {
			log.Printf("Ошибка при сканировании данных счета: %v", err)
			continue
		}
		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Ошибка при итерации по результатам запроса: %v", err)
	}

	return accounts
}

// UpdateAccountBalance обновляет баланс счета
// Возвращает ошибку, если счет не найден
func UpdateAccountBalance(accountID string, amount decimal.Decimal) error {
	// Проверяем, существует ли счет
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM accounts WHERE id = $1)", accountID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("ошибка при проверке существования счета: %w", err)
	}
	if !exists {
		return fmt.Errorf("account %s not found", accountID)
	}

	// Получаем текущий баланс
	var currentBalance decimal.Decimal
	err = db.DB.QueryRow("SELECT balance FROM accounts WHERE id = $1", accountID).Scan(&currentBalance)
	if err != nil {
		return fmt.Errorf("ошибка при получении текущего баланса: %w", err)
	}

	// Вычисляем новый баланс
	newBalance := currentBalance.Add(amount)

	// Обновляем баланс в базе данных
	_, err = db.DB.Exec("UPDATE accounts SET balance = $1 WHERE id = $2", newBalance, accountID)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении баланса: %w", err)
	}

	log.Printf("Баланс счета %s обновлен. Изменение: %s, новый баланс: %s", accountID, amount.String(), newBalance.String())
	return nil
}
