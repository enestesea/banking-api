package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Секретный ключ JWT
var jwtSecret = []byte(getJWTSecret())

// Время истечения токена по умолчанию - 24 часа
const tokenExpirationHours = 24

// Claims Структура пользовательских JWT
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// getJWTSecret возвращает JWT из переменной окружения или значение по умолчанию
func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// Секрет по умолчанию для разработки - в продакшене должен быть установлен через переменную окружения
		return "your-256-bit-secret-key-for-jwt-token-generation"
	}
	return secret
}

// EncryptUserPassword шифрует пароль пользователя с использованием bcrypt
// Возвращает зашифрованную строку и ошибку, если процесс шифрования не удался
func EncryptUserPassword(password string) (string, error) {
	// Преобразуем пароль в байты и применяем алгоритм bcrypt со стандартной сложностью
	encryptedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(encryptedBytes), err
}

// VerifyPasswordMatch проверяет соответствие пароля его хешу
// Возвращает true, если пароль соответствует хешу, и false в противном случае
func VerifyPasswordMatch(password, hash string) bool {
	// Сравниваем хеш и пароль с помощью bcrypt
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	// Если ошибки нет, значит пароль верный
	return err == nil
}

// GenerateJWT генерирует JWT-токен для пользователя
// Принимает ID пользователя и возвращает строку токена и ошибку, если генерация не удалась
func GenerateJWT(userID string) (string, error) {
	// Устанавливаем время истечения - 24 часа от текущего момента
	expirationTime := time.Now().Add(tokenExpirationHours * time.Hour)

	// Создаем JWT-утверждения, которые включают ID пользователя и время истечения
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			// Устанавливаем стандартные утверждения
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "bankapp",
			Subject:   userID,
		},
	}

	// Создаем токен с утверждениями и подписываем его секретным ключом
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)

	return tokenString, err
}

// ValidateJWT проверяет JWT-токен
// Принимает строку токена и возвращает утверждения и ошибку, если проверка не удалась
func ValidateJWT(tokenString string) (*Claims, error) {
	// Разбираем токен
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
