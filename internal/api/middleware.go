package api

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"bankapp/internal/auth"
)

// Тип ключа для значений контекста
type contextKey string

// UserIDKey - ключ для ID пользователя в контексте запроса
const UserIDKey contextKey = "userID"

// LoggingMiddleware создает обертку для HTTP-обработчиков,
// которая логирует информацию о входящих запросах и времени их выполнения
// Возвращает HTTP-обработчик с добавленным функционалом логирования
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Запоминаем время начала обработки запроса
		timeStart := time.Now()

		// Логируем информацию о входящем запросе
		log.Printf("➡️ Входящий запрос: %s %s %s", r.Method, r.RequestURI, r.Proto)

		// Передаем управление следующему обработчику в цепочке
		next.ServeHTTP(w, r)

		// Логируем информацию о завершении обработки запроса и затраченном времени
		requestDuration := time.Since(timeStart)
		log.Printf("⬅️ Запрос обработан: %s %s (время: %v)", r.Method, r.RequestURI, requestDuration)
	})
}

// AuthMiddleware создает обертку для HTTP-обработчиков,
// которая проверяет JWT-токены и добавляет ID пользователя в контекст запроса
// Возвращает HTTP-обработчик с добавленным функционалом аутентификации
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем заголовок Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondError(w, http.StatusUnauthorized, "Authorization header is required")
			return
		}

		// Проверяем, имеет ли заголовок Authorization правильный формат
		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			respondError(w, http.StatusUnauthorized, "Invalid authorization format. Format is: Bearer {token}")
			return
		}

		// Извлекаем токен
		tokenString := headerParts[1]

		// Проверяем токен
		claims, err := auth.ValidateJWT(tokenString)
		if err != nil {
			log.Printf("Недействительный токен: %v", err)
			respondError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		// Добавляем ID пользователя в контекст запроса
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)

		// Создаем новый запрос с обновленным контекстом
		r = r.WithContext(ctx)

		// Передаем управление следующему обработчику в цепочке
		next.ServeHTTP(w, r)
	})
}

// GetUserIDFromContext извлекает ID пользователя из контекста запроса
// Возвращает ID пользователя и булево значение, указывающее, найден ли ID
func GetUserIDFromContext(r *http.Request) (string, bool) {
	userID, ok := r.Context().Value(UserIDKey).(string)
	return userID, ok
}
