package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
)

// Создает уникальный идентификатор для объектов в системе 20 чисел

func CreateUniqueIdentifier() string {
	// Генерируем новый UUID v4 и возвращаем его строковое представление
	return uuid.NewString()
}

// GenerateBankAccountNumber Создает номер банковского счета в формате 20-значным
func GenerateBankAccountNumber() string {
	// Генерируем случайное число для последних 10 цифр номера счета
	randomPart, _ := rand.Int(rand.Reader, big.NewInt(9000000000))

	// Формируем номер счета: 40817810 (фиксированный префикс) + 10 случайных цифр
	// 40817810 - код для рублевых счетов физических лиц
	return fmt.Sprintf("40817810%010d", randomPart.Int64()+1000000000)
}

// GenerateCardNumber Создает случайный номер карты в допустимом формате 16 чисел
func GenerateCardNumber() string {
	n1, _ := rand.Int(rand.Reader, big.NewInt(900))
	n2, _ := rand.Int(rand.Reader, big.NewInt(10000))
	n3, _ := rand.Int(rand.Reader, big.NewInt(10000))
	n4, _ := rand.Int(rand.Reader, big.NewInt(10000))
	return fmt.Sprintf("4%03d%04d%04d%04d", n1.Int64()+100, n2.Int64(), n3.Int64(), n4.Int64())
}

// GenerateCVV Создает случайный CVV-код для карты
func GenerateCVV() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(900))
	return fmt.Sprintf("%03d", n.Int64()+100)
}

// GenerateExpiryDate Создает дату истечения срока действия карты
func GenerateExpiryDate() (int, int) {
	now := time.Now()
	year := now.Year() + 4
	month := int(now.Month())
	return month, year
}
