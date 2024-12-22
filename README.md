# Веб-сервис для вычисления арифметических выражений

## Описание
Этот проект реализует веб-сервис, который вычисляет арифметические выражения, переданные пользователем через HTTP-запрос.
Сервис поддерживает:

Вычисление арифметических выражений с операторами: `+`, `-`, `*`, `/`, `(`, `)`

Обработку ошибок, если выражение некорректно или произошла внутренняя ошибка сервиса.
## Эндпоинт
### URL:
```
/api/v1/calculate
```
### Тип запроса:
`POST`

### Формат запроса:
```json
{"expression": "арифметическое выражение"}
```

### Ответы сервиса:
1. Успешное вычисление выражения:
   - HTTP код: `200`
   - Тело ответа:
     ```json
     {"result": "результат выражения"}
     ```

2. Некорректное выражение:
   - HTTP код: `422`
   - Тело ответа:
     ```json
     {"error": "Expression is not valid"}
     ```

3. Внутренняя ошибка сервера:
   - HTTP код: `500`
   - Тело ответа:
     ```json
     {"error": "Internal server error"}
     ```

## Примеры использования

### Успешный запрос:
```bash
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{"expression": "2+2*2"}'
```
Ответ:
```json
{"result": "6"}
```

### Некорректное выражение:
```bash
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{"expression": "2+2*"}'
```
Ответ:
```json
{"error": "Expression is not valid"}
```

## Установка и запуск

1. Склонируйте репозиторий:
```bash
git clone https://github.com/jaam8/web_calculator.git
cd web_calculator
```

2. Запустите проект с помощью команды:
```bash
go run ./cmd/main.go
```

3. Сервис будет доступен по адресу: [http://localhost:8080/api/v1/calculate](http://localhost:8080/api/v1/calculate)

## Тестирование

Для запуска тестов выполните:
```bash
go test ./internal/...
```

## Структура проекта

```
web_calculator
├── cmd
│   └── main.go          # Точка входа в приложение
├── go.mod               # Зависимости проекта
├── internal             # Внутренняя логика
│   ├── api              # HTTP API обработка
│   │   ├── api.go
│   │   └── api_test.go
│   └── calculator       # Логика вычислений
│       ├── calculator.go
│       ├── calculator_test.go
│       └── errors.go    # Кастомные ошибки
└── README.md            # Документация
```