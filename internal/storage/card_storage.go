package storage

import (
	"database/sql"
	"fmt"
	"log"

	"bankapp/internal/models"
	"bankapp/pkg/utils"
)

// AddCard добавляет новую карту в базу данных
// Проверяет существование счета и добавляет карту
// Возвращает ошибку, если счет не найден
func AddCard(card models.Card) error {
	// Проверяем, существует ли счет
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM accounts WHERE id = $1)", card.AccountID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("ошибка при проверке существования счета: %w", err)
	}
	if !exists {
		return fmt.Errorf("account %s not found", card.AccountID)
	}

	// Сохраняем карту в базу данных
	query := `
		INSERT INTO cards (id, account_id, number, encrypted_number, number_hmac, 
						  expiry_month, expiry_year, cvv_hash, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err = db.DB.Exec(query, 
		card.ID, 
		card.AccountID, 
		card.Number, 
		card.EncryptedNumber, 
		card.NumberHMAC, 
		card.ExpiryMonth, 
		card.ExpiryYear, 
		card.CVVHash, 
		card.CreatedAt)

	if err != nil {
		return fmt.Errorf("ошибка при добавлении карты: %w", err)
	}

	log.Printf("Карта %s добавлена для счета %s", card.ID, card.AccountID)
	return nil
}

// GetAccountCards получает все карты для счета
// Возвращает срез карт
func GetAccountCards(accountID string) []models.Card {
	query := `
		SELECT id, account_id, number, encrypted_number, number_hmac, 
			   expiry_month, expiry_year, cvv_hash, created_at
		FROM cards
		WHERE account_id = $1
		ORDER BY created_at
	`
	rows, err := db.DB.Query(query, accountID)
	if err != nil {
		log.Printf("Ошибка при получении карт для счета: %v", err)
		return []models.Card{}
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var card models.Card
		err := rows.Scan(
			&card.ID,
			&card.AccountID,
			&card.Number,
			&card.EncryptedNumber,
			&card.NumberHMAC,
			&card.ExpiryMonth,
			&card.ExpiryYear,
			&card.CVVHash,
			&card.CreatedAt,
		)
		if err != nil {
			log.Printf("Ошибка при сканировании данных карты: %v", err)
			continue
		}
		cards = append(cards, card)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Ошибка при итерации по результатам запроса: %v", err)
	}

	return cards
}

// GetCardByNumber получает карту по ее номеру
// Возвращает карту и булево значение, указывающее, найдена ли карта
func GetCardByNumber(number string) (models.Card, bool) {
	// Генерируем HMAC для предоставленного номера для сравнения с сохраненными HMAC
	numberHMAC := utils.GenerateHMAC(number)

	// Сначала ищем карты с совпадающим HMAC
	query := `
		SELECT id, account_id, number, encrypted_number, number_hmac, 
			   expiry_month, expiry_year, cvv_hash, created_at
		FROM cards
		WHERE number_hmac = $1
	`
	rows, err := db.DB.Query(query, numberHMAC)
	if err != nil {
		log.Printf("Ошибка при поиске карты по HMAC: %v", err)
		return models.Card{}, false
	}
	defer rows.Close()

	// Проверяем каждую найденную карту
	for rows.Next() {
		var card models.Card
		err := rows.Scan(
			&card.ID,
			&card.AccountID,
			&card.Number,
			&card.EncryptedNumber,
			&card.NumberHMAC,
			&card.ExpiryMonth,
			&card.ExpiryYear,
			&card.CVVHash,
			&card.CreatedAt,
		)
		if err != nil {
			log.Printf("Ошибка при сканировании данных карты: %v", err)
			continue
		}

		// Расшифровываем номер карты, чтобы проверить точное совпадение
		decryptedNumber, err := utils.DecryptData(card.EncryptedNumber)
		if err == nil && decryptedNumber == number {
			// Обновляем поле с открытым номером карты для вызывающего кода
			card.Number = decryptedNumber
			return card, true
		}
	}

	if err := rows.Err(); err != nil {
		log.Printf("Ошибка при итерации по результатам запроса: %v", err)
	}

	return models.Card{}, false
}

// GetCard получает карту по ее ID
// Возвращает карту и булево значение, указывающее, найдена ли карта
func GetCard(cardID string) (models.Card, bool) {
	var card models.Card
	query := `
		SELECT id, account_id, number, encrypted_number, number_hmac, 
			   expiry_month, expiry_year, cvv_hash, created_at
		FROM cards
		WHERE id = $1
	`
	err := db.DB.QueryRow(query, cardID).Scan(
		&card.ID,
		&card.AccountID,
		&card.Number,
		&card.EncryptedNumber,
		&card.NumberHMAC,
		&card.ExpiryMonth,
		&card.ExpiryYear,
		&card.CVVHash,
		&card.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.Card{}, false
		}
		log.Printf("Ошибка при получении карты по ID: %v", err)
		return models.Card{}, false
	}

	return card, true
}
