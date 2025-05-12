package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"bankapp/internal/api"
	"bankapp/internal/services"
	"bankapp/internal/storage"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Println("Запуск Simple Bank API...")

	// Инициализируем соединение с базой данных
	if err := storage.InitStorage(); err != nil {
		log.Fatalf("Не удалось инициализировать хранилище: %v", err)
	}
	log.Println("Соединение с базой данных PostgreSQL установлено.")

	// Настраиваем корректное закрытие соединения с базой данных при завершении работы
	setupGracefulShutdown()

	// Запускаем планировщик платежей
	services.StartPaymentScheduler()
	log.Println("Планировщик платежей запущен.")

	r := api.SetupRouter()

	port := "8080"
	log.Printf("Сервер запускается на порту %s", port)

	loggedRouter := api.LoggingMiddleware(r)

	err := http.ListenAndServe(":"+port, loggedRouter)
	if err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}
}

// setupGracefulShutdown настраивает корректное завершение работы приложения
// при получении сигналов SIGINT или SIGTERM
func setupGracefulShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Получен сигнал завершения, закрываем соединения...")

		if err := storage.CloseStorage(); err != nil {
			log.Printf("Ошибка при закрытии соединения с базой данных: %v", err)
		} else {
			log.Println("Соединение с базой данных закрыто успешно")
		}

		log.Println("Завершение работы приложения")
		os.Exit(0)
	}()
}
