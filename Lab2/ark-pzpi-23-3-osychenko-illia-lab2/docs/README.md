# BusOptima API Documentation

## Swagger UI

Після запуску сервера, Swagger UI доступний за адресою:
```
http://localhost:8080/swagger/index.html
```

## Автентифікація

Більшість ендпоінтів потребують автентифікації. Використовуйте JWT токен в заголовку:
```
Authorization: Bearer <your-jwt-token>
```

### Отримання токена

1. Використовуйте ендпоінт `/auth/login` для отримання токена
2. Або `/auth/device` для автентифікації IoT-пристрою

## Основні групи ендпоінтів

### Authentication
- `POST /auth/login` - Автентифікація користувача
- `POST /auth/device` - Автентифікація IoT-пристрою  
- `POST /auth/refresh` - Оновлення токена

### Routes (Маршрути)
- `GET /routes` - Список маршрутів
- `GET /routes/{id}` - Маршрут за ID
- `POST /routes` - Створити маршрут
- `PUT /routes/{id}` - Оновити маршрут
- `DELETE /routes/{id}` - Видалити маршрут

### Buses (Автобуси)
- `GET /buses` - Список автобусів
- `GET /buses/{id}` - Автобус за ID
- `POST /buses` - Створити автобус
- `PUT /buses/{id}` - Оновити автобус
- `DELETE /buses/{id}` - Видалити автобус

### Trips (Рейси)
- `GET /trips` - Список рейсів
- `GET /trips/{id}` - Рейс за ID
- `POST /trips` - Створити рейс
- `PUT /trips/{id}` - Оновити рейс
- `GET /trips/{id}/events` - Події пасажирів рейсу
- `GET /trips/{id}/analytics` - Аналітика рейсу

### IoT
- `POST /iot/events` - Синхронізація подій пасажирів
- `POST /iot/price` - Рекомендація ціни
- `GET /iot/config/{tripId}` - Конфігурація рейсу

### Analytics (Аналітика)
- `GET /analytics/dashboard` - Дашборд
- `GET /analytics/forecast` - Прогноз попиту
- `GET /analytics/profitability` - Аналіз прибутковості

### Admin (Адміністрування)
- `GET /admin/users` - Список користувачів
- `POST /admin/users` - Створити користувача
- `PUT /admin/users/{id}` - Оновити користувача
- `PUT /admin/users/{id}/role` - Оновити роль користувача
- `GET /admin/audit-logs` - Журнал аудиту
- `POST /admin/backup` - Створити резервну копію

## Генерація документації

Для оновлення Swagger документації виконайте:
```bash
swag init -g cmd/api/main.go -o docs
```

## Приклади використання

### Автентифікація
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password"}'
```

### Отримання маршрутів
```bash
curl -X GET http://localhost:8080/api/routes \
  -H "Authorization: Bearer <your-token>"
```

### Створення маршруту
```bash
curl -X POST http://localhost:8080/api/routes \
  -H "Authorization: Bearer <your-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "origin_city": "Харків",
    "destination_city": "Київ",
    "distance_km": 480.5,
    "base_price": 250.00,
    "fuel_cost_per_km": 2.50,
    "driver_cost_per_trip": 800.00,
    "estimated_duration_minutes": 360
  }'
```