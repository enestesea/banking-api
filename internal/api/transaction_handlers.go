package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"bankapp/internal/models"
	"bankapp/internal/storage"
	"bankapp/pkg/utils"
)

// TransferHandler обрабатывает запросы на перевод денег между счетами
func TransferHandler(w http.ResponseWriter, r *http.Request) {
	var req models.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if req.FromAccountID == req.ToAccountID {
		respondError(w, http.StatusBadRequest, "Cannot transfer to the same account")
		return
	}
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		respondError(w, http.StatusBadRequest, "Transfer amount must be positive")
		return
	}

	fromAccount, okFrom := storage.GetAccount(req.FromAccountID)
	toAccount, okTo := storage.GetAccount(req.ToAccountID)

	if !okFrom {
		respondError(w, http.StatusNotFound, fmt.Sprintf("Source account %s not found", req.FromAccountID))
		return
	}
	if !okTo {
		respondError(w, http.StatusNotFound, fmt.Sprintf("Destination account %s not found", req.ToAccountID))
		return
	}

	if fromAccount.Balance.LessThan(req.Amount) {
		respondError(w, http.StatusPaymentRequired, "Insufficient funds in source account")
		return
	}

	err := storage.UpdateAccountBalance(req.FromAccountID, req.Amount.Neg())
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to debit source account: %v", err))
		return
	}

	err = storage.UpdateAccountBalance(req.ToAccountID, req.Amount)
	if err != nil {
		// Пытаемся откатить списание, если зачисление не удалось
		storage.UpdateAccountBalance(req.FromAccountID, req.Amount)
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to credit destination account: %v", err))
		return
	}

	tx := models.Transaction{
		ID:              utils.CreateUniqueIdentifier(),
		FromAccountID:   req.FromAccountID,
		ToAccountID:     req.ToAccountID,
		Amount:          req.Amount,
		Timestamp:       time.Now(),
		TransactionType: "transfer",
		Description:     fmt.Sprintf("Transfer from %s to %s", fromAccount.Number, toAccount.Number),
	}
	if err := storage.AddTransaction(tx); err != nil {
		log.Printf("Failed to record transaction: %v", err)
		// Продолжаем выполнение, так как деньги уже переведены
	}

	log.Printf("Transfer of %s from %s to %s successful", req.Amount.String(), req.FromAccountID, req.ToAccountID)
	respondJSON(w, http.StatusOK, map[string]string{"message": "Transfer successful"})
}

// DepositHandler handles requests to deposit money into an account
func DepositHandler(w http.ResponseWriter, r *http.Request) {
	var req models.DepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if req.Amount.LessThanOrEqual(decimal.Zero) {
		respondError(w, http.StatusBadRequest, "Deposit amount must be positive")
		return
	}

	err := storage.UpdateAccountBalance(req.ToAccountID, req.Amount)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, err.Error())
		} else {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to process deposit: %v", err))
		}
		return
	}

	account, _ := storage.GetAccount(req.ToAccountID)
	tx := models.Transaction{
		ID:              utils.CreateUniqueIdentifier(),
		FromAccountID:   "",
		ToAccountID:     req.ToAccountID,
		Amount:          req.Amount,
		Timestamp:       time.Now(),
		TransactionType: "deposit",
		Description:     fmt.Sprintf("Deposit to account %s", account.Number),
	}
	if err := storage.AddTransaction(tx); err != nil {
		log.Printf("Failed to record transaction: %v", err)
		// Продолжаем выполнение, так как деньги уже зачислены
	}

	log.Printf("Deposit of %s to account %s successful", req.Amount.String(), req.ToAccountID)
	respondJSON(w, http.StatusOK, map[string]string{"message": "Deposit successful"})
}
