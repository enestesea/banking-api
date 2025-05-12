# Banking API

## Описание проекта
Banking API - это RESTful API для банковского приложения, разработанное на языке Go. Система предоставляет полный набор функций для управления банковскими счетами, картами, переводами, кредитами и аналитикой.

## Тестирование
READMETesting
## Функциональные возможности
- Регистрация и аутентификация пользователей с использованием JWT
- Создание и управление банковскими счетами
- Выпуск и использование банковских карт
- Переводы между счетами и пополнение счетов
- Оформление и обслуживание кредитов
- Аналитика и история транзакций
- Прогнозирование баланса

## Системные требования
- Go 1.24 или выше
- PostgreSQL 12 или выше
- Доступ к SMTP-серверу для отправки уведомлений (опционально)

## Установка и запуск

### Клонирование репозитория
```bash
git clone https://github.com/yourusername/banking-api.git
cd banking-api
```

### Установка зависимостей
```bash
go mod download
```

### Настройка базы данных
1. Создайте базу данных PostgreSQL:
```sql
CREATE DATABASE banking;
```

2. Настройте переменные окружения для подключения к базе данных:
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD='your-password'
export DB_NAME=banking
export DB_SSLMODE=disable
```

Или используйте значения по умолчанию, указанные в файле `internal/config/db_config.go`.

### Запуск приложения
```bash
go run cmd/bankapp/main.go
```

По умолчанию сервер запускается на порту 8080.

## Структура проекта
```
banking-api/
├── cmd/
│   └── bankapp/           # Точка входа в приложение
├── internal/
│   ├── api/               # Обработчики HTTP-запросов
│   ├── auth/              # Аутентификация и авторизация
│   ├── config/            # Конфигурация приложения
│   ├── models/            # Модели данных
│   ├── services/          # Бизнес-логика
│   └── storage/           # Работа с базой данных
├── pkg/
│   └── utils/             # Вспомогательные функции
└── go.mod, go.sum         # Зависимости проекта
```

## API Endpoints

### Аутентификация
- **POST /register** - Регистрация нового пользователя
- **POST /login** - Вход в систему и получение JWT-токена

### Управление счетами
- **POST /accounts** - Создание нового счета
- **GET /users/{userId}/accounts** - Получение всех счетов пользователя

### Управление картами
- **POST /cards** - Выпуск новой карты
- **GET /accounts/{accountId}/cards** - Получение всех карт счета
- **POST /payments/card** - Оплата картой

### Переводы и пополнения
- **POST /transfers** - Перевод между счетами
- **POST /deposits** - Пополнение счета

### Кредиты
- **POST /loans** - Оформление кредита
- **GET /loans/{loanId}/schedule** - Получение графика платежей по кредиту

### Аналитика
- **GET /analytics/transactions/{accountId}** - История транзакций по счету
- **GET /analytics/summary/{userId}** - Финансовая сводка пользователя
- **GET /accounts/{accountId}/predict** - Прогнозирование баланса

## Аутентификация
Все защищенные эндпоинты требуют JWT-токен в заголовке Authorization:
```
Authorization: Bearer <ваш_токен>
```

Токен можно получить, выполнив запрос на эндпоинт `/login`.

## Примеры использования API

### Регистрация пользователя
```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "securepassword"
  }'
```

### Вход в систему
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "securepassword"
  }'
```

### Создание счета
```bash
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <ваш_токен>" \
  -d '{
    "user_id": "<id_пользователя>"
  }'
```

### Выпуск карты
```bash
curl -X POST http://localhost:8080/cards \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <ваш_токен>" \
  -d '{
    "account_id": "<id_счета>"
  }'
```

### Перевод между счетами
```bash
curl -X POST http://localhost:8080/transfers \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <ваш_токен>" \
  -d '{
    "from_account_id": "<id_счета_отправителя>",
    "to_account_id": "<id_счета_получателя>",
    "amount": 100.50
  }'
```

## Планировщик платежей
Приложение включает планировщик, который автоматически обрабатывает платежи по кредитам каждые 12 часов. Если платеж просрочен и на счете достаточно средств, платеж будет выполнен автоматически. В случае недостаточности средств будет начислен штраф в размере 10% от суммы платежа.

## Безопасность
- Пароли пользователей хранятся в виде хешей с использованием bcrypt
- Данные карт (номер, CVV) хранятся в зашифрованном виде
- Все API-запросы к защищенным эндпоинтам требуют JWT-аутентификации
- Используется HTTPS для защиты данных при передаче (требуется настройка в production)

## Логирование
Приложение ведет подробное логирование всех операций, включая:
- Регистрацию и вход пользователей
- Создание счетов и карт
- Финансовые операции
- Обработку кредитов
- Ошибки и предупреждения

## Уведомления
Система поддерживает отправку email-уведомлений пользователям о важных событиях. Для включения этой функции необходимо настроить параметры SMTP-сервера в файле `internal/services/notification_service.go`.

