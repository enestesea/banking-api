package storage

import (
	"database/sql"
	"fmt"
	"log"

	"bankapp/internal/models"
)

// RegisterNewUser Добавляет нового пользователя в базу данных
// Проверяет уникальность имени пользователя и электронной почты
// Возвращает ошибку, если пользователь с таким же именем или электронной почтой уже существует
func RegisterNewUser(user models.User) error {
	// Проверяем, занято ли уже имя пользователя
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", user.Username).Scan(&exists)
	if err != nil {
		return fmt.Errorf("ошибка при проверке имени пользователя: %w", err)
	}
	if exists {
		return fmt.Errorf("username '%s' is already taken", user.Username)
	}

	// Проверяем, зарегистрирована ли уже электронная почта
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", user.Email).Scan(&exists)
	if err != nil {
		return fmt.Errorf("ошибка при проверке электронной почты: %w", err)
	}
	if exists {
		return fmt.Errorf("email '%s' is already registered", user.Email)
	}

	// Сохраняем пользователя в базу данных
	query := `
		INSERT INTO users (id, username, email, password_hash, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = db.DB.Exec(query, user.ID, user.Username, user.Email, user.PasswordHash, user.CreatedAt)
	if err != nil {
		return fmt.Errorf("ошибка при сохранении пользователя: %w", err)
	}

	log.Printf("Пользователь %s (ID: %s) успешно зарегистрирован в базе данных", user.Username, user.ID)
	return nil
}

// GetUserByUsername Получает пользователя по его имени пользователя
// Возвращает пользователя и булево значение, указывающее, найден ли пользователь
func GetUserByUsername(username string) (models.User, bool) {
	var user models.User
	query := `
		SELECT id, username, email, password_hash, created_at
		FROM users
		WHERE username = $1
	`
	err := db.DB.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, false
		}
		log.Printf("Ошибка при получении пользователя по имени: %v", err)
		return models.User{}, false
	}

	return user, true
}

// GetUserByID Получает пользователя по его ID
// Возвращает пользователя и булево значение, указывающее, найден ли пользователь
func GetUserByID(userID string) (models.User, bool) {
	var user models.User
	query := `
		SELECT id, username, email, password_hash, created_at
		FROM users
		WHERE id = $1
	`
	err := db.DB.QueryRow(query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, false
		}
		log.Printf("Ошибка при получении пользователя по ID: %v", err)
		return models.User{}, false
	}

	return user, true
}

// GetAllUsers Получает всех пользователей из базы данных
// Возвращает список пользователей
func GetAllUsers() ([]models.User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at
		FROM users
		ORDER BY created_at
	`
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении списка пользователей: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании данных пользователя: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по результатам запроса: %w", err)
	}

	return users, nil
}
