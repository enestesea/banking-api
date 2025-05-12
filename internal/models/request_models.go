package models

import (
	"github.com/shopspring/decimal"
)

// RegisterRequest содержит данные для регистрации нового пользователя
type RegisterRequest struct {
	Username string `json:"username"` // Имя пользователя
	Email    string `json:"email"`    // Адрес электронной почты
	Password string `json:"password"` // Пароль (в открытом виде, будет хешироваться)
}

// LoginRequest содержит данные для входа пользователя
type LoginRequest struct {
	Username string `json:"username"` // Имя пользователя
	Password string `json:"password"` // Пароль
}

// CreateAccountRequest содержит данные для создания нового банковского счета
type CreateAccountRequest struct {
	UserID string `json:"user_id"`    // ID пользователя, для которого создается счет
}

// GenerateCardRequest содержит данные для выпуска новой банковской карты
type GenerateCardRequest struct {
	AccountID string `json:"account_id"` // ID счета, к которому будет привязана карта
}

// PaymentRequest содержит данные для совершения платежа по карте
type PaymentRequest struct {
	CardNumber string          `json:"card_number"` // Номер карты
	Amount     decimal.Decimal `json:"amount"`      // Сумма платежа
	Merchant   string          `json:"merchant"`    // Получатель платежа (магазин, сервис)
}

// TransferRequest содержит данные для перевода средств между счетами
type TransferRequest struct {
	FromAccountID string          `json:"from_account_id"` // Счет отправителя
	ToAccountID   string          `json:"to_account_id"`   // Счет получателя
	Amount        decimal.Decimal `json:"amount"`          // Сумма перевода
}

// DepositRequest содержит данные для пополнения счета
type DepositRequest struct {
	ToAccountID string          `json:"to_account_id"` // Счет для пополнения
	Amount      decimal.Decimal `json:"amount"`        // Сумма пополнения
}

// ApplyLoanRequest содержит данные для оформления кредита
type ApplyLoanRequest struct {
	UserID     string          `json:"user_id"`      // ID заемщика
	AccountID  string          `json:"account_id"`   // Счет для выдачи кредита
	Amount     decimal.Decimal `json:"amount"`       // Сумма кредита
	TermMonths int             `json:"term_months"`  // Срок кредита в месяцах
}
