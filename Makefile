# ============================================
# ПЕРЕМЕННЫЕ
# ============================================
PROJECT_NAME = weather-parser
DOCKER_COMPOSE = docker-compose

# ============================================
# ЦЕЛИ
# ============================================

# Сборка и запуск
.PHONY: up
up:
	$(DOCKER_COMPOSE) up -d
	$(DOCKER_COMPOSE) logs -f

# Сборка без запуска
.PHONY: build
build:
	$(DOCKER_COMPOSE) build

# Остановка
.PHONY: down
down:
	$(DOCKER_COMPOSE) down

# Остановка с удалением данных
.PHONY: clean
clean:
	$(DOCKER_COMPOSE) down -v
	docker system prune -f

# Перезапуск
.PHONY: restart
restart: down up

# Просмотр логов
.PHONY: logs
logs:
	$(DOCKER_COMPOSE) logs -f

# Просмотр логов приложения
.PHONY: logs-app
logs-app:
	$(DOCKER_COMPOSE) logs -f weather-parser

# Просмотр логов БД
.PHONY: logs-db
logs-db:
	$(DOCKER_COMPOSE) logs -f postgres

# Зайти в контейнер приложения
.PHONY: shell
shell:
	docker exec -it weather-parser sh

# Зайти в контейнер БД
.PHONY: shell-db
shell-db:
	docker exec -it weather-postgres psql -U postgres -d weather

# Статус контейнеров
.PHONY: status
status:
	$(DOCKER_COMPOSE) ps

# Локальный запуск (без Docker)
.PHONY: local
local:
	go run . -city=Volgograd

# Тесты
.PHONY: test
test:
	go test ./...

# Сборка бинарника для Linux
.PHONY: build-linux
build-linux:
	CGO_ENABLED=0 GOOS=linux go build -o $(PROJECT_NAME) .

# ============================================
# HELP
# ============================================
.PHONY: help
help:
	@echo "📋 Доступные команды:"
	@echo "  make up          - Запустить все контейнеры"
	@echo "  make build       - Собрать образы"
	@echo "  make down        - Остановить контейнеры"
	@echo "  make clean       - Остановить и удалить все данные"
	@echo "  make restart     - Перезапустить"
	@echo "  make logs        - Показать логи всех контейнеров"
	@echo "  make logs-app    - Показать логи приложения"
	@echo "  make logs-db     - Показать логи БД"
	@echo "  make shell       - Зайти в контейнер приложения"
	@echo "  make shell-db    - Зайти в БД (psql)"
	@echo "  make status      - Статус контейнеров"
	@echo "  make local       - Запустить локально (без Docker)"
	@echo "  make test        - Запустить тесты"
	@echo "  make help        - Показать эту справку"