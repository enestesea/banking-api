package services

import (
	"log"
	"sync"
	"time"

	"github.com/shopspring/decimal"

	"bankapp/internal/models"
	"bankapp/internal/storage"
)

var (
	schedulerRunning bool
	schedulerMutex   sync.Mutex
)

// StartPaymentScheduler запускает планировщик для обработки платежей по кредитам
// Планировщик работает каждые 12 часов, как указано в требованиях
func StartPaymentScheduler() {
	schedulerMutex.Lock()
	if schedulerRunning {
		schedulerMutex.Unlock()
		return
	}
	schedulerRunning = true
	schedulerMutex.Unlock()

	log.Println("Запуск планировщика платежей")

	// Запускаем сразу при старте
	processOverduePayments()

	// Затем запускаем каждые 12 часов
	ticker := time.NewTicker(12 * time.Hour)
	go func() {
		for range ticker.C {
			processOverduePayments()
		}
	}()
}

// processOverduePayments обрабатывает все просроченные платежи по кредитам
func processOverduePayments() {
	log.Println("Обработка просроченных платежей")

	// Получаем все кредиты
	loans := storage.GetAllLoans()
	now := time.Now()

	for _, loan := range loans {
		// Пропускаем полностью погашенные кредиты
		if loan.RemainingAmount.IsZero() {
			continue
		}

		// Проверяем каждый платеж в графике
		modified := false
		for i, payment := range loan.PaymentSchedule {
			// Пропускаем уже оплаченные платежи
			if payment.Paid {
				continue
			}

			// Проверяем, просрочен ли платеж
			if payment.DueDate.Before(now) {
				account, ok := storage.GetAccount(loan.AccountID)
				if !ok {
					log.Printf("Счет %s не найден для кредита %s", loan.AccountID, loan.ID)
					continue
				}

				// Проверяем, достаточно ли средств на счете
				if account.Balance.GreaterThanOrEqual(payment.Amount) {
					// Обрабатываем платеж
					err := storage.UpdateAccountBalance(loan.AccountID, payment.Amount.Neg())
					if err != nil {
						log.Printf("Не удалось списать средства со счета %s для платежа по кредиту: %v", loan.AccountID, err)
						continue
					}

					// Обновляем статус платежа
					loan.PaymentSchedule[i].Paid = true

					// Обновляем оставшуюся сумму
					loan.RemainingAmount = loan.RemainingAmount.Sub(payment.PrincipalPart)

					// Записываем транзакцию
					tx := models.Transaction{
						ID:              storage.GenerateTransactionID(),
						FromAccountID:   loan.AccountID,
						Amount:          payment.Amount,
						Timestamp:       time.Now(),
						TransactionType: "loan_payment",
						Description:     "Автоматический платеж по кредиту",
					}
					storage.AddTransaction(tx)

					log.Printf("Обработан платеж %s для кредита %s", payment.Amount.String(), loan.ID)
					modified = true
				} else {
					// Применяем штраф (10% от суммы платежа) согласно требованиям
					penaltyAmount := payment.Amount.Mul(decimal.NewFromFloat(0.1))

					// Создаем новый платеж со штрафом
					penaltyPayment := models.Payment{
						DueDate:       time.Now().AddDate(0, 0, 7), // Срок через 7 дней
						Amount:        penaltyAmount,
						PrincipalPart: decimal.Zero,                // Весь штраф идет на проценты
						InterestPart:  penaltyAmount,
						Paid:          false,
					}

					// Добавляем штрафной платеж в график
					loan.PaymentSchedule = append(loan.PaymentSchedule, penaltyPayment)

					log.Printf("Применен штраф %s за просроченный платеж по кредиту %s", penaltyAmount.String(), loan.ID)
					modified = true
				}
			}
		}

		// Обновляем кредит, если были изменения
		if modified {
			err := storage.UpdateLoan(loan)
			if err != nil {
				log.Printf("Не удалось обновить кредит %s: %v", loan.ID, err)
			}
		}
	}

	log.Println("Завершена обработка просроченных платежей")
}
