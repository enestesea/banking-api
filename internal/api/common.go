package api

import (
	"encoding/json"
	"log"
	"net/http"
)

// respondJSON преобразует данные в JSON и отправляет их в качестве ответа
// с указанным HTTP-кодом статуса
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Ошибка маршалинга JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal server error"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

// respondError отправляет ответ с ошибкой с указанным HTTP-кодом статуса и сообщением
func respondError(w http.ResponseWriter, code int, message string) {
	log.Printf("HTTP-ошибка %d: %s", code, message)
	respondJSON(w, code, map[string]string{"error": message})
}
