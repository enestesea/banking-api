package storage

import (
	"database/sql"
	"fmt"
	"log"

	"bankapp/internal/models"
)

// AddLoan adds a new loan to the database
// Checks for user and account existence and adds the loan to the database
// Returns an error if the user or account is not found
func AddLoan(loan models.Loan) error {
	// Проверяем, существует ли пользователь
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", loan.UserID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("ошибка при проверке существования пользователя: %w", err)
	}
	if !exists {
		return fmt.Errorf("user %s not found", loan.UserID)
	}

	// Проверяем, существует ли счет
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM accounts WHERE id = $1)", loan.AccountID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("ошибка при проверке существования счета: %w", err)
	}
	if !exists {
		return fmt.Errorf("account %s not found", loan.AccountID)
	}

	// Начинаем транзакцию для атомарной вставки кредита и графика платежей
	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("ошибка при начале транзакции: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Сохраняем кредит в базу данных
	query := `
		INSERT INTO credits (id, user_id, account_id, amount, interest_rate, term_months, 
							start_date, remaining_amount, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err = tx.Exec(query,
		loan.ID,
		loan.UserID,
		loan.AccountID,
		loan.Amount,
		loan.InterestRate,
		loan.TermMonths,
		loan.StartDate,
		loan.RemainingAmount,
		loan.StartDate) // используем StartDate как created_at

	if err != nil {
		return fmt.Errorf("ошибка при сохранении кредита: %w", err)
	}

	// Сохраняем график платежей
	for _, payment := range loan.PaymentSchedule {
		query := `
			INSERT INTO payment_schedules (credit_id, due_date, amount, principal_part, interest_part, paid)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err = tx.Exec(query,
			loan.ID,
			payment.DueDate,
			payment.Amount,
			payment.PrincipalPart,
			payment.InterestPart,
			payment.Paid)

		if err != nil {
			return fmt.Errorf("ошибка при сохранении графика платежей: %w", err)
		}
	}

	// Фиксируем транзакцию
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("ошибка при фиксации транзакции: %w", err)
	}

	log.Printf("Кредит %s добавлен для пользователя %s на сумму %s", loan.ID, loan.UserID, loan.Amount.String())
	return nil
}

// GetUserLoans retrieves all loans for a user
// Returns a slice of loans
func GetUserLoans(userID string) []models.Loan {
	loans, err := getLoansWithFilter("user_id = $1", userID)
	if err != nil {
		log.Printf("Ошибка при получении кредитов пользователя: %v", err)
		return []models.Loan{}
	}
	return loans
}

// GetLoan retrieves a loan by its ID
// Returns the loan and a boolean indicating if the loan was found
func GetLoan(loanID string) (models.Loan, bool) {
	loans, err := getLoansWithFilter("id = $1", loanID)
	if err != nil {
		log.Printf("Ошибка при получении кредита: %v", err)
		return models.Loan{}, false
	}
	if len(loans) == 0 {
		return models.Loan{}, false
	}
	return loans[0], true
}

// GetAccountLoans retrieves all loans associated with an account
// Returns a slice of loans
func GetAccountLoans(accountID string) []models.Loan {
	loans, err := getLoansWithFilter("account_id = $1", accountID)
	if err != nil {
		log.Printf("Ошибка при получении кредитов счета: %v", err)
		return []models.Loan{}
	}
	return loans
}

// GetAllLoans retrieves all loans from the database
// Returns a slice of all loans
func GetAllLoans() []models.Loan {
	loans, err := getLoansWithFilter("1=1", nil)
	if err != nil {
		log.Printf("Ошибка при получении всех кредитов: %v", err)
		return []models.Loan{}
	}
	return loans
}

// UpdateLoan updates a loan in the database
// Returns an error if the loan is not found
func UpdateLoan(loan models.Loan) error {
	// Проверяем, существует ли кредит
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM credits WHERE id = $1)", loan.ID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("ошибка при проверке существования кредита: %w", err)
	}
	if !exists {
		return fmt.Errorf("loan %s not found", loan.ID)
	}

	// Начинаем транзакцию для атомарного обновления кредита и графика платежей
	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("ошибка при начале транзакции: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Обновляем данные кредита
	query := `
		UPDATE credits
		SET user_id = $2, account_id = $3, amount = $4, interest_rate = $5, 
			term_months = $6, start_date = $7, remaining_amount = $8
		WHERE id = $1
	`
	_, err = tx.Exec(query,
		loan.ID,
		loan.UserID,
		loan.AccountID,
		loan.Amount,
		loan.InterestRate,
		loan.TermMonths,
		loan.StartDate,
		loan.RemainingAmount)

	if err != nil {
		return fmt.Errorf("ошибка при обновлении кредита: %w", err)
	}

	// Удаляем существующий график платежей
	_, err = tx.Exec("DELETE FROM payment_schedules WHERE credit_id = $1", loan.ID)
	if err != nil {
		return fmt.Errorf("ошибка при удалении графика платежей: %w", err)
	}

	// Сохраняем обновленный график платежей
	for _, payment := range loan.PaymentSchedule {
		query := `
			INSERT INTO payment_schedules (credit_id, due_date, amount, principal_part, interest_part, paid)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err = tx.Exec(query,
			loan.ID,
			payment.DueDate,
			payment.Amount,
			payment.PrincipalPart,
			payment.InterestPart,
			payment.Paid)

		if err != nil {
			return fmt.Errorf("ошибка при сохранении обновленного графика платежей: %w", err)
		}
	}

	// Фиксируем транзакцию
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("ошибка при фиксации транзакции: %w", err)
	}

	log.Printf("Кредит %s обновлен. Оставшаяся сумма: %s", loan.ID, loan.RemainingAmount.String())
	return nil
}

// getLoansWithFilter is a helper function to get loans with a specific filter
// Returns a slice of loans and an error
func getLoansWithFilter(filter string, param interface{}) ([]models.Loan, error) {
	var rows *sql.Rows
	var err error

	// Формируем запрос с фильтром
	query := fmt.Sprintf(`
		SELECT id, user_id, account_id, amount, interest_rate, term_months, 
			   start_date, remaining_amount, created_at
		FROM credits
		WHERE %s
		ORDER BY created_at DESC
	`, filter)

	// Выполняем запрос с параметром или без
	if param != nil {
		rows, err = db.DB.Query(query, param)
	} else {
		rows, err = db.DB.Query(query)
	}

	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса: %w", err)
	}
	defer rows.Close()

	var loans []models.Loan
	for rows.Next() {
		var loan models.Loan
		err := rows.Scan(
			&loan.ID,
			&loan.UserID,
			&loan.AccountID,
			&loan.Amount,
			&loan.InterestRate,
			&loan.TermMonths,
			&loan.StartDate,
			&loan.RemainingAmount,
			&loan.StartDate, // используем StartDate для created_at
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании данных кредита: %w", err)
		}

		// Получаем график платежей для этого кредита
		paymentQuery := `
			SELECT due_date, amount, principal_part, interest_part, paid
			FROM payment_schedules
			WHERE credit_id = $1
			ORDER BY due_date
		`
		paymentRows, err := db.DB.Query(paymentQuery, loan.ID)
		if err != nil {
			return nil, fmt.Errorf("ошибка при получении графика платежей: %w", err)
		}
		defer paymentRows.Close()

		var payments []models.Payment
		for paymentRows.Next() {
			var payment models.Payment
			err := paymentRows.Scan(
				&payment.DueDate,
				&payment.Amount,
				&payment.PrincipalPart,
				&payment.InterestPart,
				&payment.Paid,
			)
			if err != nil {
				return nil, fmt.Errorf("ошибка при сканировании данных платежа: %w", err)
			}
			payments = append(payments, payment)
		}

		if err := paymentRows.Err(); err != nil {
			return nil, fmt.Errorf("ошибка при итерации по результатам запроса платежей: %w", err)
		}

		loan.PaymentSchedule = payments
		loans = append(loans, loan)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по результатам запроса кредитов: %w", err)
	}

	return loans, nil
}
