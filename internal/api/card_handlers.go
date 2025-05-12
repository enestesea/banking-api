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

// GenerateCardHandler обрабатывает запросы на генерацию новой карты для счета
func GenerateCardHandler(w http.ResponseWriter, r *http.Request) {
	var req models.GenerateCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if _, ok := storage.GetAccount(req.AccountID); !ok {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Account %s not found", req.AccountID))
		return
	}

	// Генерируем данные карты
	month, year := utils.GenerateExpiryDate()
	cardNumber := utils.GenerateCardNumber()
	cvv := utils.GenerateCVV()

	// Encrypt and hash sensitive data
	encryptedNumber, err := utils.EncryptData(cardNumber)
	if err != nil {
		log.Printf("Error encrypting card number: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to secure card data")
		return
	}

	numberHMAC := utils.GenerateHMAC(cardNumber)

	cvvHash, err := utils.HashCVV(cvv)
	if err != nil {
		log.Printf("Error hashing CVV: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to secure card data")
		return
	}

	// Create the card with encrypted data
	card := models.Card{
		ID:              utils.CreateUniqueIdentifier(),
		AccountID:       req.AccountID,
		Number:          cardNumber,
		EncryptedNumber: encryptedNumber,
		NumberHMAC:      numberHMAC,
		ExpiryMonth:     month,
		ExpiryYear:      year,
		CVV:             cvv,
		CVVHash:         cvvHash,
		CreatedAt:       time.Now(),
	}

	if err := storage.AddCard(card); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate card: %v", err))
		return
	}

	log.Printf("Card generated for account %s", card.AccountID)

	// Возвращаем безопасную версию карты
	respondJSON(w, http.StatusCreated, card.SecureCard())
}

// GetAccountCardsHandler обрабатывает запросы на получение всех карт для счета
func GetAccountCardsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID := vars["accountId"]

	if _, ok := storage.GetAccount(accountID); !ok {
		respondError(w, http.StatusNotFound, fmt.Sprintf("Account %s not found", accountID))
		return
	}

	cards := storage.GetAccountCards(accountID)

	// Создаем безопасные версии карт для ответа
	secureCards := make([]models.Card, len(cards))
	for i, card := range cards {
		secureCards[i] = card.SecureCard()
	}

	log.Printf("Fetched %d cards for account %s", len(cards), accountID)
	respondJSON(w, http.StatusOK, secureCards)
}

// PayWithCardHandler обрабатывает запросы на совершение платежа с использованием карты
func PayWithCardHandler(w http.ResponseWriter, r *http.Request) {
	var req models.PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if req.Amount.LessThanOrEqual(decimal.Zero) {
		respondError(w, http.StatusBadRequest, "Payment amount must be positive")
		return
	}

	card, ok := storage.GetCardByNumber(req.CardNumber)
	if !ok {
		respondError(w, http.StatusNotFound, "Card not found")
		return
	}

	now := time.Now()
	expiry := time.Date(card.ExpiryYear, time.Month(card.ExpiryMonth)+1, 0, 23, 59, 59, 0, time.UTC) // Last day of the month
	if now.After(expiry) {
		respondError(w, http.StatusBadRequest, "Card expired")
		return
	}

	account, ok := storage.GetAccount(card.AccountID)
	if !ok {
		respondError(w, http.StatusInternalServerError, "Associated account not found")
		return
	}

	if account.Balance.LessThan(req.Amount) {
		respondError(w, http.StatusPaymentRequired, "Insufficient funds")
		return
	}

	err := storage.UpdateAccountBalance(account.ID, req.Amount.Neg())
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to process payment: %v", err))
		return
	}

	tx := models.Transaction{
		ID:              utils.CreateUniqueIdentifier(),
		FromAccountID:   account.ID,
		ToAccountID:     "",
		Amount:          req.Amount,
		Timestamp:       time.Now(),
		TransactionType: "payment",
		Description:     fmt.Sprintf("Payment to %s", req.Merchant),
	}
	if err := storage.AddTransaction(tx); err != nil {
		log.Printf("Failed to record transaction: %v", err)
		// Продолжаем выполнение, так как деньги уже списаны
	}

	log.Printf("Payment of %s processed from account %s (card %s) to %s", req.Amount.String(), account.ID, card.Number[:4]+"...", req.Merchant)
	respondJSON(w, http.StatusOK, map[string]string{"message": "Payment successful"})
}
