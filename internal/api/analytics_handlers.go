package api

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"

	"bankapp/internal/models"
	"bankapp/internal/storage"
)

// GetTransactionsHandler обрабатывает запросы на получение транзакций для счета
func GetTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID := vars["accountId"]

	if _, ok := storage.GetAccount(accountID); !ok {
		respondError(w, http.StatusNotFound, fmt.Sprintf("Account %s not found", accountID))
		return
	}

	transactions := storage.GetAccountTransactions(accountID)

	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].Timestamp.After(transactions[j].Timestamp)
	})

	log.Printf("Fetched %d transactions for account %s", len(transactions), accountID)
	respondJSON(w, http.StatusOK, transactions)
}

// GetFinancialSummaryHandler обрабатывает запросы на получение финансовой сводки для пользователя
func GetFinancialSummaryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	accounts := storage.GetUserAccounts(userID)
	loans := storage.GetUserLoans(userID)

	totalBalance := decimal.Zero
	for _, acc := range accounts {
		totalBalance = totalBalance.Add(acc.Balance)
	}

	totalLoanDebt := decimal.Zero
	activeLoans := 0
	for _, loan := range loans {
		totalLoanDebt = totalLoanDebt.Add(loan.RemainingAmount)
		if loan.RemainingAmount.GreaterThan(decimal.Zero) {
			activeLoans++
		}
	}

	summary := map[string]interface{}{
		"user_id":               userID,
		"total_account_balance": totalBalance,
		"number_of_accounts":    len(accounts),
		"total_loan_debt":       totalLoanDebt,
		"active_loans":          activeLoans,
	}

	log.Printf("Generated financial summary for user %s", userID)
	respondJSON(w, http.StatusOK, summary)
}

// BalancePredictionHandler обрабатывает запросы на прогнозирование баланса счета на будущие дни
func BalancePredictionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID := vars["accountId"]

	// Получаем параметр дней (по умолчанию 30, если не указан)
	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		var err error
		days, err = strconv.Atoi(daysStr)
		if err != nil || days <= 0 {
			respondError(w, http.StatusBadRequest, "Invalid days parameter. Must be a positive integer.")
			return
		}

		// Устанавливаем максимальный период прогнозирования
		if days > 365 {
			respondError(w, http.StatusBadRequest, "Maximum prediction period is 365 days")
			return
		}
	}

	// Получаем счет
	account, ok := storage.GetAccount(accountID)
	if !ok {
		respondError(w, http.StatusNotFound, fmt.Sprintf("Account %s not found", accountID))
		return
	}

	// Получаем все транзакции по счету для анализа шаблонов
	transactions := storage.GetAccountTransactions(accountID)

	// Получаем все кредиты, связанные с этим счетом
	loans := storage.GetAccountLoans(accountID)

	// Текущий баланс является отправной точкой
	currentBalance := account.Balance

	// Рассчитываем среднедневные расходы на основе последних 30 дней
	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	var outgoingAmount decimal.Decimal
	var incomingAmount decimal.Decimal
	outgoingCount := 0
	incomingCount := 0

	for _, tx := range transactions {
		if tx.Timestamp.Before(thirtyDaysAgo) {
			continue
		}

		if tx.FromAccountID == accountID {
			outgoingAmount = outgoingAmount.Add(tx.Amount)
			outgoingCount++
		}

		if tx.ToAccountID == accountID {
			incomingAmount = incomingAmount.Add(tx.Amount)
			incomingCount++
		}
	}

	// Рассчитываем среднедневные показатели
	avgDailyOutgoing := decimal.Zero
	if outgoingCount > 0 {
		avgDailyOutgoing = outgoingAmount.Div(decimal.NewFromInt(30))
	}

	avgDailyIncoming := decimal.Zero
	if incomingCount > 0 {
		avgDailyIncoming = incomingAmount.Div(decimal.NewFromInt(30))
	}

	// Прогнозируем будущие балансы
	dailyPredictions := make([]models.BalancePrediction, days)
	predictedBalance := currentBalance

	for i := 0; i < days; i++ {
		predictionDate := now.AddDate(0, 0, i+1)

		// Применяем среднедневные изменения
		predictedBalance = predictedBalance.Add(avgDailyIncoming).Sub(avgDailyOutgoing)

		// Проверяем запланированные платежи по кредитам
		for _, loan := range loans {
			for _, payment := range loan.PaymentSchedule {
				// Если платеж должен быть выполнен в этот день и еще не оплачен
				if payment.DueDate.Year() == predictionDate.Year() && 
				   payment.DueDate.Month() == predictionDate.Month() && 
				   payment.DueDate.Day() == predictionDate.Day() && 
				   !payment.Paid {
					// Вычитаем сумму платежа из прогнозируемого баланса
					predictedBalance = predictedBalance.Sub(payment.Amount)
					break
				}
			}
		}

		// Создаем запись прогноза
		dailyPredictions[i] = models.BalancePrediction{
			Date:    predictionDate,
			Balance: predictedBalance,
		}
	}

	// Создаем ответ
	response := models.BalancePredictionResponse{
		AccountID:           accountID,
		CurrentBalance:      currentBalance,
		PredictionDays:      days,
		AverageDailyOutflow: avgDailyOutgoing,
		AverageDailyInflow:  avgDailyIncoming,
		Predictions:         dailyPredictions,
	}

	log.Printf("Generated %d-day balance prediction for account %s", days, accountID)
	respondJSON(w, http.StatusOK, response)
}
