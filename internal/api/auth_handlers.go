package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"bankapp/internal/auth"
	"bankapp/internal/models"
	"bankapp/internal/services"
	"bankapp/internal/storage"
	"bankapp/pkg/utils"
)

// RegisterUserHandler обрабатывает запросы на регистрацию пользователей
func RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if req.Username == "" || req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Username, email, and password are required")
		return
	}

	hashedPassword, err := auth.EncryptUserPassword(req.Password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	user := models.User{
		ID:           utils.CreateUniqueIdentifier(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
	}

	if err := storage.RegisterNewUser(user); err != nil {
		respondError(w, http.StatusConflict, err.Error())
		return
	}

	go func() {
		subject := "Welcome to Simple Bank!"
		body := fmt.Sprintf("Hello %s,\n\nThank you for registering at Simple Bank.", user.Username)
		err := services.SendNotification(user.Email, subject, body)
		if err != nil {
			log.Printf("Failed to send registration email to %s: %v", user.Email, err)
		}
	}()

	log.Printf("User registered: %s (ID: %s)", user.Username, user.ID)
	user.PasswordHash = ""
	respondJSON(w, http.StatusCreated, user)
}

// LoginUserHandler обрабатывает запросы на вход пользователей
func LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	user, ok := storage.GetUserByUsername(req.Username)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	if !auth.VerifyPasswordMatch(req.Password, user.PasswordHash) {
		respondError(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	// Генерируем JWT токен
	token, err := auth.GenerateJWT(user.ID)
	if err != nil {
		log.Printf("Failed to generate JWT token: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to generate authentication token")
		return
	}

	log.Printf("User logged in: %s", user.Username)
	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Login successful",
		"user_id": user.ID,
		"token":   token,
	})
}
