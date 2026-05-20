@echo off
chcp 65001 > nul
echo === Запуск дистрибутива Jira Analytics System ===

if not exist .env (
    echo [!] Файл .env не найден. Создаю .env на основе .env.example...
    copy .env.example .env
)

echo [*] Сборка и запуск микросервисов в Docker...
docker-compose up --build -d

if %errorlevel% equ 0 (
    echo ==================================================================
    echo [✔] УСПЕХ: Все сервисы и инфраструктура успешно запущены!
    echo ==================================================================
    echo   БИЗНЕС-МИКРОСЕРВИСЫ:
    echo   [*] Gateway Backend : http://localhost:8000
    echo   [*] Auth Service    : http://localhost:8083
    echo   [*] Jira Connector  : http://localhost:8081 (gRPC: 50051)
    echo   [*] Kafka Service   : http://localhost:8082
    echo   [*] UI (Frontend)   : http://localhost:3000
    echo ==================================================================
) else (
    echo [✘] ОШИБКА: Не удалось поднять контейнеры.
)
pause