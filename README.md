# Person Service

Сервис для обогащения данных о людях с использованием внешних API.

## Требования

- Go 1.21 или выше
- PostgreSQL
- Docker (опционально)

## Установка

1. Клонируйте репозиторий:
```bash
git clone https://github.com/yourusername/person-service.git
cd person-service
```

2. Установите зависимости:
```bash
go mod download
```

3. Обновите файл `.env` в корневой директории проекта:
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=person_service
```

## Запуск

1. Запустите сервис:
```bash
go run main.go
```

Сервис автоматически:
- Проверит наличие базы данных и создаст её при необходимости
- Подключится к базе данных
- Создаст необходимые таблицы
- Применит миграции
- Запустит HTTP сервер на порту 8080

## API Endpoints

### Создание записи о человеке
```http
POST /api/v1/person
Content-Type: application/json

{
    "name": "Иван",
    "surname": "Иванов",
    "patronymic": "Иванович"
}
```

### Получение списка людей
```http
GET /api/v1/person?page=1&limit=10
```

### Получение информации о человеке
```http
GET /api/v1/person/{id}
```

### Обновление информации о человеке
```http
PUT /api/v1/person/{id}
Content-Type: application/json

{
    "name": "Иван",
    "surname": "Иванов",
    "patronymic": "Иванович"
}
```

### Удаление информации о человеке
```http
DELETE /api/v1/person/{id}
```

## Swagger

Документация API доступна по адресу: http://localhost:8080/swagger/index.html