package services

import (
	"fmt"
	"log"
	"net/smtp"
)

// Конфигурация SMTP для отправки электронных писем
var smtpConfig = struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}{
	Host:     "smtp.example.com",
	Port:     587,
	Username: "your_email@example.com",
	Password: "your_password",
	From:     "bankapp@example.com",
}

// SendNotification отправляет электронное письмо пользователю
// Принимает адрес получателя, тему и текст сообщения
// Возвращает ошибку, если отправка не удалась
func SendNotification(to, subject, body string) error {
	// Проверяем, настроен ли SMTP-сервер
	if smtpConfig.Host == "smtp.example.com" {
		log.Printf("SMTP не настроен. Пропускаем отправку письма на %s: Тема: %s", to, subject)
		return nil
	}

	// Создаем аутентификацию для SMTP-сервера
	authData := smtp.PlainAuth("", smtpConfig.Username, smtpConfig.Password, smtpConfig.Host)

	// Формируем заголовки и тело сообщения
	emailContent := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s\r\n",
		smtpConfig.From, to, subject, body)

	// Формируем адрес SMTP-сервера с портом
	serverAddress := fmt.Sprintf("%s:%d", smtpConfig.Host, smtpConfig.Port)

	// Отправляем письмо
	err := smtp.SendMail(serverAddress, authData, smtpConfig.From, []string{to}, []byte(emailContent))
	if err != nil {
		log.Printf("Ошибка отправки письма на %s: %v", to, err)
		return fmt.Errorf("не удалось отправить письмо: %w", err)
	}

	log.Printf("Письмо успешно отправлено на %s", to)
	return nil
}
