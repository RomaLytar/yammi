#!/bin/bash

# Integration тесты для Board Service с testcontainers
# Убедитесь, что Docker запущен перед выполнением

set -e

echo "🧪 Running Board Service Integration Tests"
echo "=========================================="

# Проверка, что Docker доступен
if ! docker ps > /dev/null 2>&1; then
    echo "❌ Error: Docker is not running"
    echo "Please start Docker and try again"
    exit 1
fi

# Переход в директорию сервиса
cd "$(dirname "$0")/.." || exit 1

echo "📦 Installing dependencies..."
go mod download

echo ""
echo "🏃 Running integration tests..."
echo ""

# Запуск тестов с verbose output и timeout 10 минут
go test ./tests/integration/... \
    -v \
    -timeout 10m \
    -count=1 \
    "$@"

echo ""
echo "✅ All integration tests passed!"
