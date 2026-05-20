.PHONY: up start stop clean logs-backend logs-connector logs-kafka file-logs-backend file-logs-connector file-logs-kafka

# Запуск всего проекта с пересборкой
up:
	@if [ ! -f .env ]; then \
		echo "[!] Файл .env не найден. Создаю .env на основе .env.example..."; \
		cp .env.example .env; \
	fi
	sudo docker-compose up -d --build
	@echo "=================================================================="
	@echo "[✔] УСПЕХ: Все сервисы и инфраструктура успешно запущены!"
	@echo "=================================================================="
	@echo "  БИЗНЕС-МИКРОСЕРВИСЫ:"
	@echo "  [*] Gateway Backend : http://localhost:8000"
	@echo "  [*] Auth Service    : http://localhost:8083"
	@echo "  [*] Jira Connector  : http://localhost:8081 (gRPC: 50051)"
	@echo "  [*] Kafka Service   : http://localhost:8082"
	@echo "  [*] UI (Frontend)   : http://localhost:3000"
	@echo "------------------------------------------------------------------"
	@echo "  ИНФРАСТРУКТУРА:"
	@echo "  [*] Внешние порты базы данных и pgAdmin берутся из .env"
	@echo "=================================================================="

# Просто запуск без пересборки
start:
	sudo docker-compose up -d

# Остановка проекта
stop:
	sudo docker-compose stop

# Полная очистка системы (удаление контейнеров, данных БД и файловых логов)
clean:
	sudo docker-compose down -v
	sudo rm -rf logs/*/

# Просмотр консольных логов
logs-backend:
	docker logs -f dev_backend

logs-connector:
	docker logs -f dev_connector

logs-kafka:
	docker logs -f dev_kafka_service

# Просмотр файловых логов
file-logs-backend:
	cat logs/backend/logs.log

file-logs-connector:
	cat logs/connector/logs.log

file-logs-kafka:
	cat logs/kafka/logs.log

dist: clean
	tar --exclude='logs' --exclude='.git' --exclude='pgdata' -czf ../jira_analytics_v1.0.tar.gz .
	@echo "=================================================================="
	@echo "[✔] Дистрибутив успешно собран в файл: jira_analytics_v1.0.tar.gz"
	@echo "=================================================================="
