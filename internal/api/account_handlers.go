package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"

	"bankapp/internal/models"
	"bankapp/internal/storage"
	"bankapp/pkg/utils"
)

// CreateAccountHandler обрабатывает реквесты к счету
func CreateAccountHandler(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if req.UserID == "" {
		respondError(w, http.StatusBadRequest, "UserID is required")
		return
	}

	account := models.Account{
		ID:        utils.CreateUniqueIdentifier(),
		UserID:    req.UserID,
		Number:    utils.GenerateBankAccountNumber(),
		Balance:   decimal.Zero,
		CreatedAt: time.Now(),
	}

	if err := storage.CreateBankAccount(account); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create account: %v", err))
		return
	}

	log.Printf("Account created: %s for user %s", account.Number, account.UserID)
	respondJSON(w, http.StatusCreated, account)
}

// GetUserAccountsHandler достает все счета пользователя
func GetUserAccountsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	accounts := storage.GetUserAccounts(userID)
	log.Printf("Fetched %d accounts for user %s", len(accounts), userID)
	respondJSON(w, http.StatusOK, accounts)
}
