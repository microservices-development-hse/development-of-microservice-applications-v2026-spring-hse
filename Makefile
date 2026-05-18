# Запуск всего проекта с пересборкой
up:
	sudo docker-compose up -d --build

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

# Просмотр файловых логов
file-logs-backend:
	cat logs/backend/logs.log

file-logs-connector:
	cat logs/connector/logs.log
