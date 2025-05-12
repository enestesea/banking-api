package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
)

// Ключ шифрования для PGP
var encryptionKey = []byte(getEncryptionKey())

// Секрет HMAC для проверки целостности данных
var hmacSecret = []byte(getHMACSecret())

// getEncryptionKey возвращает ключ шифрования из переменной окружения или значение по умолчанию
func getEncryptionKey() string {
	key := os.Getenv("ENCRYPTION_KEY")
	if key == "" {
		// Ключ по умолчанию для разработки - в продакшене должен быть установлен через переменную окружения
		return "default-encryption-key-for-development-only"
	}
	return key
}

// getHMACSecret возвращает секрет HMAC из переменной окружения или значение по умолчанию
func getHMACSecret() string {
	secret := os.Getenv("HMAC_SECRET")
	if secret == "" {
		// Секрет по умолчанию для разработки - в продакшене должен быть установлен через переменную окружения
		return "default-hmac-secret-for-development-only"
	}
	return secret
}

// EncryptData шифрует данные с использованием простого XOR-шифра (имитация PGP )
func EncryptData(data string) (string, error) {
	if data == "" {
		return "", nil
	}

	dataBytes := []byte(data)

	result := make([]byte, len(dataBytes))

	for i := 0; i < len(dataBytes); i++ {
		result[i] = dataBytes[i] ^ encryptionKey[i%len(encryptionKey)]
	}

	// Кодируем результат в base64
	return base64.StdEncoding.EncodeToString(result), nil
}

// DecryptData расшифровывает данные, зашифрованные с помощью EncryptData
func DecryptData(encryptedData string) (string, error) {
	if encryptedData == "" {
		return "", nil
	}

	dataBytes, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", fmt.Errorf("не удалось декодировать base64: %w", err)
	}

	result := make([]byte, len(dataBytes))

	// XOR каждого байта с соответствующим байтом из ключа
	for i := 0; i < len(dataBytes); i++ {
		result[i] = dataBytes[i] ^ encryptionKey[i%len(encryptionKey)]
	}

	return string(result), nil
}

// GenerateHMAC Генерирует HMAC для проверки целостности данных
func GenerateHMAC(data string) string {
	h := hmac.New(sha256.New, hmacSecret)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// Проверяет, соответствуют ли данные предоставленному HMAC
func VerifyHMAC(data, providedHMAC string) bool {
	expectedHMAC := GenerateHMAC(data)
	return hmac.Equal([]byte(expectedHMAC), []byte(providedHMAC))
}

// HashCVV хеширует CVV с использованием SHA-256
func HashCVV(cvv string) (string, error) {
	if cvv == "" {
		return "", errors.New("CVV не может быть пустым")
	}
	//Модификатор входа хэш-функции
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", fmt.Errorf("не удалось сгенерировать соль: %w", err)
	}

	// Создаем хеш SHA-256
	h := sha256.New()
	h.Write(salt)
	h.Write([]byte(cvv))

	result := append(salt, h.Sum(nil)...)

	//Вв hex
	return hex.EncodeToString(result), nil
}

// VerifyCVV проверяет, соответствует ли CVV его хешу
func VerifyCVV(cvv, hash string) (bool, error) {
	if cvv == "" || hash == "" {
		return false, errors.New("CVV и хеш не могут быть пустыми")
	}

	hashBytes, err := hex.DecodeString(hash)
	if err != nil {
		return false, fmt.Errorf("не удалось декодировать хеш: %w", err)
	}

	// Извлекаем соль (первые 16 байт)
	if len(hashBytes) < 16 {
		return false, errors.New("неверный формат хеша")
	}
	salt := hashBytes[:16]
	storedHash := hashBytes[16:]

	// Создаем новый хеш
	h := sha256.New()
	h.Write(salt)
	h.Write([]byte(cvv))
	computedHash := h.Sum(nil)

	// Сравниваем хеши
	return hmac.Equal(storedHash, computedHash), nil
}
