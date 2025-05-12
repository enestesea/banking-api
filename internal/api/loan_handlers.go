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
	"bankapp/internal/services"
	"bankapp/internal/storage"
	"bankapp/pkg/utils"
)

// ApplyLoanHandler обрабатывает запросы на получение кредита
func ApplyLoanHandler(w http.ResponseWriter, r *http.Request) {
	var req models.ApplyLoanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if req.Amount.LessThanOrEqual(decimal.Zero) || req.TermMonths <= 0 {
		respondError(w, http.StatusBadRequest, "Loan amount and term must be positive")
		return
	}

	_, userExists := storage.GetUserByID(req.UserID)
	_, accountExists := storage.GetAccount(req.AccountID)

	if !userExists {
		respondError(w, http.StatusNotFound, fmt.Sprintf("User %s not found", req.UserID))
		return
	}
	if !accountExists {
		respondError(w, http.StatusNotFound, fmt.Sprintf("Account %s not found", req.AccountID))
		return
	}

	baseRate, err := services.FetchCentralBankRate()
	if err != nil {
		log.Printf("Warning: Failed to get key rate, using default 10%%: %v", err)
		baseRate = decimal.NewFromInt(10)
	}

	interestRate := baseRate.Add(decimal.NewFromInt(5))

	monthlyPayment := utils.CalculateMonthlyPayment(req.Amount, interestRate, req.TermMonths)
	startDate := time.Now()
	schedule := utils.GeneratePaymentSchedule(req.Amount, interestRate, req.TermMonths, startDate, monthlyPayment)

	loan := models.Loan{
		ID:              utils.CreateUniqueIdentifier(),
		UserID:          req.UserID,
		AccountID:       req.AccountID,
		Amount:          req.Amount,
		InterestRate:    interestRate,
		TermMonths:      req.TermMonths,
		StartDate:       startDate,
		PaymentSchedule: schedule,
		RemainingAmount: req.Amount,
	}

	if err := storage.AddLoan(loan); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to save loan: %v", err))
		return
	}

	err = storage.UpdateAccountBalance(req.AccountID, req.Amount)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to disburse loan funds: %v", err))
		return
	}

	tx := models.Transaction{
		ID:              utils.CreateUniqueIdentifier(),
		FromAccountID:   "",
		ToAccountID:     req.AccountID,
		Amount:          req.Amount,
		Timestamp:       time.Now(),
		TransactionType: "loan_disbursement",
		Description:     fmt.Sprintf("Loan disbursement (ID: %s)", loan.ID),
	}
	if err := storage.AddTransaction(tx); err != nil {
		log.Printf("Failed to record loan disbursement transaction: %v", err)
		// Продолжаем выполнение, так как кредит уже выдан и деньги зачислены
	}

	log.Printf("Loan %s approved for user %s, amount %s, rate %s%%, term %d months. Funds disbursed to account %s.",
		loan.ID, req.UserID, req.Amount.String(), interestRate.String(), req.TermMonths, req.AccountID)

	respondJSON(w, http.StatusCreated, loan)
}

// GetLoanScheduleHandler обрабатывает запросы на получение графика платежей по кредиту
func GetLoanScheduleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	loanID := vars["loanId"]

	loan, ok := storage.GetLoan(loanID)
	if !ok {
		respondError(w, http.StatusNotFound, fmt.Sprintf("Loan %s not found", loanID))
		return
	}

	log.Printf("Fetched payment schedule for loan %s", loanID)
	respondJSON(w, http.StatusOK, loan.PaymentSchedule)
}
