#!/bin/bash

# Базовый URL API
BASE_URL="http://localhost:8080"
API_KEY="your-secret-key"

# Функция для отправки запросов с авторизацией
api_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    
    curl -s -X $method \
        -H "Content-Type: application/json" \
        -H "Key: $API_KEY" \
        -d "$data" \
        "$BASE_URL$endpoint"
}

# Тест 1: Создание пользователя
echo "=== Тест 1: Создание пользователя ==="
USER_DATA='{"name": "John Doe", "email": "john@example.com"}'
response=$(api_request POST "/users" "$USER_DATA")
echo "Response: $response"

# Извлекаем ID пользователя
USER_ID=$(echo $response | grep -o '"id":[0-9]*' | cut -d: -f2)
echo "Created user ID: $USER_ID"

# Тест 2: Создание объявления (нужен реальный файл)
echo -e "\n=== Тест 2: Создание объявления ==="
echo "Note: This requires a real image file. Use curl with -F flag in practice."

# Тест 3: Получение всех объявлений
echo -e "\n=== Тест 3: Получение всех объявлений ==="
response=$(api_request GET "/ads" "")
echo "Response: $response"

# Тест 4: Удаление пользователя
echo -e "\n=== Тест 4: Удаление пользователя ==="
response=$(api_request DELETE "/users/$USER_ID" "")
echo "Response: $response"

echo -e "\n=== Тестирование завершено ==="