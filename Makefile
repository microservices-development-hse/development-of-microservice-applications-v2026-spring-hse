# Запуск всего проекта с пересборкой
up:
	sudo docker-compose up -d --build

# Просто запуск без пересборки
start:
	sudo docker-compose up -d

# Остановка проекта
stop:
	sudo docker-compose stop

# Полная очистка системы (удаление контейнеров и данных БД)
clean:
	sudo docker-compose down -v

# Просмотр логов бэкенда
logs-backend:
	docker logs -f dev_backend

# Просмотр логов коннектора
logs-connector:
	docker logs -f dev_connector