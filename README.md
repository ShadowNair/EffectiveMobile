## README.md

````md
# Subscriptions Service (Effective Mobile test task)

REST-сервис для агрегации данных о подписках пользователей.

## Возможности
- CRUDL операции над подписками
- Подсчет суммарной стоимости подписок за период с фильтрацией по:
  - user_id (UUID)
  - service_name (string)
- PostgreSQL + миграции
- Swagger UI
- Логи (slog + middleware)

## Формат дат
MM-YYYY (например 07-2025).  
В базе даты хранятся как первое число месяца (day=1), чтобы было проще считать месяцы.

## Запуск через Docker Compose
docker compose up --build
````

Сервис:

* API: `http://localhost:8080`
* Swagger UI: `http://localhost:8080/swagger/`
* Healthcheck: `http://localhost:8080/healthz`

## Переменные окружения (.env)

Пример (минимум):

```env
APP_HOST=0.0.0.0
APP_PORT=8080

POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=subscriptions

LOG_LEVEL=info
```

## API

### Create
```
POST /api/v1/subscriptions/
```
```
{
  "service_name": "Yandex Plus",
  "price": 400,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "07-2025"
}
```

### Get
```
GET /api/v1/subscriptions/{id}/
```
### Update
```
PUT /api/v1/subscriptions/{id}/
```
### Delete
```
DELETE /api/v1/subscriptions/{id}/
```
### List
```
GET /api/v1/subscriptions/?user_id=...&service_name=...&limit=50&offset=0
```
### Summary
```
GET /api/v1/subscriptions/summary?from=07-2025&to=12-2025&user_id=...&service_name=...
```
Ответ:
```
{
  "total": 2400,
  "currency": "RUB",
  "from": "07-2025",
  "to": "12-2025",
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "service_name": "Yandex Plus"
}
```
## Dev команды

### Тесты
```
make test
```
### Линтер
```
make lint
```
### Форматирование
```
make fmt
```