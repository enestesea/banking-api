package api

import (
	"github.com/gorilla/mux"
)

// SetupRouter создает и настраивает маршрутизатор API
// Регистрирует все обработчики запросов
// Возвращает настроенный маршрутизатор
func SetupRouter() *mux.Router {
	r := mux.NewRouter()

	// Публичные маршруты (аутентификация не требуется)
	r.HandleFunc("/register", RegisterUserHandler).Methods("POST")
	r.HandleFunc("/login", LoginUserHandler).Methods("POST")

	// Защищенные маршруты (требуется аутентификация)
	// Создаем подмаршрутизатор для защищенных маршрутов
	protected := r.PathPrefix("").Subrouter()

	// Применяем middleware аутентификации ко всем защищенным маршрутам
	protected.Use(AuthMiddleware)

	// Маршруты управления счетами
	protected.HandleFunc("/accounts", CreateAccountHandler).Methods("POST")
	protected.HandleFunc("/users/{userId}/accounts", GetUserAccountsHandler).Methods("GET")

	// Маршруты управления картами
	protected.HandleFunc("/cards", GenerateCardHandler).Methods("POST")
	protected.HandleFunc("/accounts/{accountId}/cards", GetAccountCardsHandler).Methods("GET")
	protected.HandleFunc("/payments/card", PayWithCardHandler).Methods("POST")

	// Маршруты для переводов и пополнений
	protected.HandleFunc("/transfers", TransferHandler).Methods("POST")
	protected.HandleFunc("/deposits", DepositHandler).Methods("POST")

	// Маршруты для кредитов
	protected.HandleFunc("/loans", ApplyLoanHandler).Methods("POST")
	protected.HandleFunc("/loans/{loanId}/schedule", GetLoanScheduleHandler).Methods("GET")

	// Маршруты для аналитики
	protected.HandleFunc("/analytics/transactions/{accountId}", GetTransactionsHandler).Methods("GET")
	protected.HandleFunc("/analytics/summary/{userId}", GetFinancialSummaryHandler).Methods("GET")

	// Эндпоинт прогнозирования баланса
	protected.HandleFunc("/accounts/{accountId}/predict", BalancePredictionHandler).Methods("GET")

	return r
}
