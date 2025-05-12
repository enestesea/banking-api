package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// User представляет информацию о зарегистрированном пользователе в системе
type User struct {
	ID           string    `json:"id"`         // Уникальный идентификатор пользователя
	Username     string    `json:"username"`   // Имя пользователя для входа в систему
	Email        string    `json:"email"`      // Электронная почта пользователя
	PasswordHash string    `json:"-"`          // Хеш пароля (не отправляется в JSON)
	CreatedAt    time.Time `json:"created_at"` // Дата и время регистрации
}

// Account представляет банковский счет пользователя
type Account struct {
	ID        string          `json:"id"`         // Уникальный идентификатор счета
	UserID    string          `json:"user_id"`    // Идентификатор владельца счета
	Number    string          `json:"number"`     // Номер счета в банковском формате
	Balance   decimal.Decimal `json:"balance"`    // Текущий баланс счета
	CreatedAt time.Time       `json:"created_at"` // Дата и время создания счета
}

// Card представляет платежную карту, привязанную к счету
type Card struct {
	ID              string    `json:"id"`         // Уникальный идентификатор карты
	AccountID       string    `json:"account_id"` // Счет, к которому привязана карта
	Number          string    `json:"number"`     // Номер карты (маскируется в JSON)
	EncryptedNumber string    `json:"-"`          // Зашифрованный номер карты (хранится в БД)
	NumberHMAC      string    `json:"-"`          // HMAC для проверки целостности номера карты
	ExpiryMonth     int       `json:"expiry_month"`
	ExpiryYear      int       `json:"expiry_year"`
	CVV             string    `json:"-"` // Код безопасности (не отправляется в JSON)
	CVVHash         string    `json:"-"` // Хешированный CVV (хранится в БД)
	CreatedAt       time.Time `json:"created_at"`
}

// SecureCard создает безопасную версию карты с маскированным номером для ответов API
func (c Card) SecureCard() Card {
	// Создаем копию карты
	secureCard := c

	// Маскируем номер карты (показываем только последние 4 цифры)
	if len(c.Number) > 4 {
		secureCard.Number = "****-****-****-" + c.Number[len(c.Number)-4:]
	}

	// Очищаем конфиденциальные данные
	secureCard.EncryptedNumber = ""
	secureCard.NumberHMAC = ""
	secureCard.CVV = ""
	secureCard.CVVHash = ""

	return secureCard
}

// Transaction представляет финансовую операцию в системе
type Transaction struct {
	ID              string          `json:"id"`
	FromAccountID   string          `json:"from_account_id,omitempty"`
	ToAccountID     string          `json:"to_account_id,omitempty"`
	Amount          decimal.Decimal `json:"amount"`
	Timestamp       time.Time       `json:"timestamp"`
	TransactionType string          `json:"transaction_type"` //Тип транзакции например платеж
	Description     string          `json:"description,omitempty"`
}

// Loan представляет информацию о выданном кредите
type Loan struct {
	ID              string          `json:"id"`
	UserID          string          `json:"user_id"`
	AccountID       string          `json:"account_id"`
	Amount          decimal.Decimal `json:"amount"`
	InterestRate    decimal.Decimal `json:"interest_rate"`    // Процентная ставка
	TermMonths      int             `json:"term_months"`      // Срок кредита в месяцах
	StartDate       time.Time       `json:"start_date"`       // Дата выдачи кредита
	PaymentSchedule []Payment       `json:"payment_schedule"` // График платежей
	RemainingAmount decimal.Decimal `json:"remaining_amount"` // Оставшаяся сумма долга
}

// Payment представляет информацию о платеже по кредиту
type Payment struct {
	DueDate       time.Time       `json:"due_date"`
	Amount        decimal.Decimal `json:"amount"`
	PrincipalPart decimal.Decimal `json:"principal_part"` // Часть платежа, идущая на погашение основного долга
	InterestPart  decimal.Decimal `json:"interest_part"`  // Часть платежа, идущая на погашение процентов
	Paid          bool            `json:"paid"`           // Флаг, указывающий, был ли платеж совершен
}
