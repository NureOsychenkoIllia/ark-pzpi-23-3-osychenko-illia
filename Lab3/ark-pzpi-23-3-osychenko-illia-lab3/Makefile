# Makefile для BusOptima API

.PHONY: help build run test clean docker-up docker-down migrate

# Змінні
APP_NAME=busoptima
DOCKER_COMPOSE=docker-compose
GO_FILES=$(shell find . -name "*.go" -type f)

# Допомога
help: ## Показати цю допомогу
	@echo "Доступні команди:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Збірка
build: ## Зібрати додаток
	@echo "Збірка $(APP_NAME)..."
	@go build -o bin/$(APP_NAME) ./cmd/api

# Запуск
run: ## Запустити додаток локально
	@echo "Запуск $(APP_NAME)..."
	@go run ./cmd/api/main.go

# Тестування
test: ## Запустити тести
	@echo "Запуск тестів..."
	@go test -v ./...

# Очищення
clean: ## Очистити артефакти збірки
	@echo "Очищення..."
	@rm -rf bin/
	@go clean

# Docker команди
docker-up: ## Запустити всі сервіси через Docker Compose
	@echo "Запуск Docker сервісів..."
	@$(DOCKER_COMPOSE) up -d

docker-down: ## Зупинити всі Docker сервіси
	@echo "Зупинка Docker сервісів..."
	@$(DOCKER_COMPOSE) down

docker-logs: ## Показати логи Docker сервісів
	@$(DOCKER_COMPOSE) logs -f

docker-rebuild: ## Перезібрати та перезапустити Docker сервіси
	@echo "Перезбірка Docker сервісів..."
	@$(DOCKER_COMPOSE) down
	@$(DOCKER_COMPOSE) build --no-cache
	@$(DOCKER_COMPOSE) up -d

# База даних
migrate-up: ## Застосувати міграції БД
	@echo "Застосування міграцій..."
	@docker exec -i busoptima_db psql -U busoptima_user -d busoptima < migrations/001_initial_schema.sql
	@docker exec -i busoptima_db psql -U busoptima_user -d busoptima < migrations/002_initial_data.sql

migrate-down: ## Відкатити міграції БД
	@echo "Відкат міграцій..."
	@docker exec -i busoptima_db psql -U busoptima_user -d busoptima -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

regenerate-hashes: ## Регенерувати хеші паролів та токенів в БД
	@echo "Регенерація хешів..."
	@go run scripts/regenerate_hashes.go

regenerate-hashes-sql: ## Регенерувати хеші через SQL скрипт
	@echo "Регенерація хешів через SQL..."
	@docker exec -i busoptima_db psql -U busoptima_user -d busoptima < scripts/regenerate_hashes.sql

db-shell: ## Підключитися до БД через psql
	@docker exec -it busoptima_db psql -U busoptima_user -d busoptima

# Розробка
dev: docker-up ## Запустити середовище розробки
	@echo "Середовище розробки готове!"
	@echo "API: http://localhost:8080"
	@echo "Adminer: http://localhost:8081"
	@echo "PostgreSQL: localhost:5432"

# Форматування коду
fmt: ## Форматувати код
	@echo "Форматування коду..."
	@go fmt ./...

# Лінтинг
lint: ## Запустити лінтер
	@echo "Лінтинг коду..."
	@golangci-lint run

# Встановлення залежностей
deps: ## Встановити залежності
	@echo "Встановлення залежностей..."
	@go mod download
	@go mod tidy

# Генерація документації
docs: ## Генерувати документацію API
	@echo "Генерація документації..."
	@swag init -g cmd/api/main.go -o docs

docs-serve: ## Запустити сервер з документацією
	@echo "Запуск сервера з документацією..."
	@echo "Swagger UI буде доступний за адресою: http://localhost:8080/swagger/index.html"
	@$(MAKE) run

docs-validate: ## Перевірити валідність Swagger документації
	@echo "Перевірка Swagger документації..."
	@swagger validate docs/swagger.json || echo "Swagger CLI не встановлено. Встановіть: npm install -g swagger"

# Повна перевірка
check: fmt lint test ## Повна перевірка коду (форматування, лінтинг, тести)

# Швидкий старт
quickstart: deps docker-up migrate-up ## Швидкий старт проекту
	@echo "Проект готовий до використання!"
	@echo "Тестові користувачі:"
	@echo "  admin@busoptima.ua / password123"
	@echo "  dispatcher@busoptima.ua / password123"