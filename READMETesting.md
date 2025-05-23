# Тестирование Banking API

В данном документе описаны процедуры тестирования Banking API, включая примеры запросов, ответов и рекомендации по проверке функциональности.


## Подготовка к тестированию

Перед началом тестирования необходимо убедиться, что:
- API-сервер запущен и доступен
- База данных PostgreSQL настроена и работает
- Установлен инструмент для тестирования API (например, Postman)



## Регистрация пользователя

Для регистрации нового пользователя отправьте POST-запрос на эндпоинт `/register` с данными пользователя.

**Пример запроса:**
```json
{
  "username": "testuser",
  "email": "test@example.com",
  "password": "securepassword"
}
```

**Ожидаемый ответ:** Статус 201 Created и данные созданного пользователя.
![Снимок экрана 2025-05-12 в 17.35.30.png](screenshots/%D0%A1%D0%BD%D0%B8%D0%BC%D0%BE%D0%BA%20%D1%8D%D0%BA%D1%80%D0%B0%D0%BD%D0%B0%202025-05-12%20%D0%B2%2017.35.30.png)


## Вход в систему

После регистрации выполните вход в систему, отправив POST-запрос на эндпоинт `/login`.

**Пример запроса:**
```json
{
  "username": "testuser",
  "password": "securepassword"
}
```

**Ожидаемый ответ:** Статус 200 OK и JWT-токен для аутентификации.
![Снимок экрана 2025-05-12 в 17.35.56.png](screenshots/%D0%A1%D0%BD%D0%B8%D0%BC%D0%BE%D0%BA%20%D1%8D%D0%BA%D1%80%D0%B0%D0%BD%D0%B0%202025-05-12%20%D0%B2%2017.35.56.png)

Сохраните полученный JWT-токен для использования в последующих запросах.

## Создание счета

Для создания нового счета отправьте POST-запрос на эндпоинт `/accounts` с указанием ID пользователя.

**Пример запроса:**
```json
{
  "user_id": "полученный_id_пользователя"
}
```

**Ожидаемый ответ:** Статус 201 Created и данные созданного счета.
![Снимок экрана 2025-05-12 в 17.40.33.png](screenshots/%D0%A1%D0%BD%D0%B8%D0%BC%D0%BE%D0%BA%20%D1%8D%D0%BA%D1%80%D0%B0%D0%BD%D0%B0%202025-05-12%20%D0%B2%2017.40.33.png)


После создания счета можно проверить его наличие, отправив GET-запрос на эндпоинт `/users/{userId}/accounts`.


## Выпуск карты

Для выпуска новой карты отправьте POST-запрос на эндпоинт `/cards` с указанием ID счета.

**Пример запроса:**
```json
{
  "account_id": "id_счета"
}
```

**Ожидаемый ответ:** Статус 201 Created и данные выпущенной карты.
![Снимок экрана 2025-05-12 в 17.46.26.png](screenshots/%D0%A1%D0%BD%D0%B8%D0%BC%D0%BE%D0%BA%20%D1%8D%D0%BA%D1%80%D0%B0%D0%BD%D0%B0%202025-05-12%20%D0%B2%2017.46.26.png)


## Оплата картой

Для совершения платежа с использованием карты отправьте POST-запрос на эндпоинт `/payments/card`.

**Пример запроса:**
```json
{
  "card_number": "номер_карты",
  "amount": 150.75,
  "merchant": "Интернет-магазин"
}
```

**Ожидаемый ответ:** Статус 200 OK и сообщение об успешном платеже.
![Снимок экрана 2025-05-12 в 18.40.38.png](screenshots/%D0%A1%D0%BD%D0%B8%D0%BC%D0%BE%D0%BA%20%D1%8D%D0%BA%D1%80%D0%B0%D0%BD%D0%B0%202025-05-12%20%D0%B2%2018.40.38.png)


## Перевод между счетами

Для тестирования перевода между счетами отправьте POST-запрос на эндпоинт `/transfers`.

**Пример запроса:**
```json
{
  "from_account_id": "id_счета_отправителя",
  "to_account_id": "id_счета_получателя",
  "amount": 200.00
}
```

**Ожидаемый ответ:** Статус 200 OK и сообщение об успешном переводе.
![Снимок экрана 2025-05-12 в 17.54.36.png](screenshots/%D0%A1%D0%BD%D0%B8%D0%BC%D0%BE%D0%BA%20%D1%8D%D0%BA%D1%80%D0%B0%D0%BD%D0%B0%202025-05-12%20%D0%B2%2017.54.36.png)

## Оформление кредита

Для тестирования оформления кредита отправьте POST-запрос на эндпоинт `/loans`.

**Пример запроса:**
```json
{
  "user_id": "id_пользователя",
  "account_id": "id_счета",
  "amount": 10000.00,
  "term_months": 12
}
```

**Ожидаемый ответ:** Статус 201 Created и данные оформленного кредита.
![Снимок экрана 2025-05-12 в 18.06.02.png](screenshots/%D0%A1%D0%BD%D0%B8%D0%BC%D0%BE%D0%BA%20%D1%8D%D0%BA%D1%80%D0%B0%D0%BD%D0%B0%202025-05-12%20%D0%B2%2018.06.02.png)

## График платежей по кредиту

Для получения графика платежей по кредиту отправьте GET-запрос на эндпоинт `/loans/{loanId}/schedule`.

**Ожидаемый ответ:** Статус 200 OK и массив платежей с датами и суммами.
![Снимок экрана 2025-05-12 в 18.06.50.png](screenshots/%D0%A1%D0%BD%D0%B8%D0%BC%D0%BE%D0%BA%20%D1%8D%D0%BA%D1%80%D0%B0%D0%BD%D0%B0%202025-05-12%20%D0%B2%2018.06.50.png)
## История транзакций

Для получения истории транзакций по счету отправьте GET-запрос на эндпоинт `/analytics/transactions/{accountId}`.

**Ожидаемый ответ:** Статус 200 OK и массив транзакций с деталями.
![Снимок экрана 2025-05-12 в 18.07.39.png](screenshots/%D0%A1%D0%BD%D0%B8%D0%BC%D0%BE%D0%BA%20%D1%8D%D0%BA%D1%80%D0%B0%D0%BD%D0%B0%202025-05-12%20%D0%B2%2018.07.39.png)


## Прогнозирование и финансовая сводка

### Прогнозирование баланса

Для получения прогноза баланса отправьте GET-запрос на эндпоинт `/accounts/{accountId}/predict`.

**Ожидаемый ответ:** Статус 200 OK и данные прогноза.
![Снимок экрана 2025-05-12 в 18.09.18.png](screenshots/%D0%A1%D0%BD%D0%B8%D0%BC%D0%BE%D0%BA%20%D1%8D%D0%BA%D1%80%D0%B0%D0%BD%D0%B0%202025-05-12%20%D0%B2%2018.09.18.png)
### Финансовая сводка

Для получения финансовой сводки пользователя отправьте GET-запрос на эндпоинт `/analytics/summary/{userId}`.

**Ожидаемый ответ:** Статус 200 OK и данные финансовой сводки.
![Снимок экрана 2025-05-12 в 18.29.23.png](screenshots/%D0%A1%D0%BD%D0%B8%D0%BC%D0%BE%D0%BA%20%D1%8D%D0%BA%D1%80%D0%B0%D0%BD%D0%B0%202025-05-12%20%D0%B2%2018.29.23.png)
