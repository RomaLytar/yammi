# Makefile для запуска integration тестов Board Service

.PHONY: test test-verbose test-single clean help

# Запуск всех integration тестов
test:
	@echo "Running integration tests..."
	@go test ./... -v -timeout 10m

# Запуск с подробным выводом
test-verbose:
	@echo "Running integration tests with verbose output..."
	@go test ./... -v -timeout 10m -count=1

# Запуск отдельного теста (использование: make test-single TEST=TestBoardRepository_Create)
test-single:
	@echo "Running single test: $(TEST)..."
	@go test ./... -run $(TEST) -v -timeout 5m

# Запуск с coverage
test-coverage:
	@echo "Running integration tests with coverage..."
	@go test ./... -v -timeout 10m -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report saved to coverage.html"

# Очистка
clean:
	@echo "Cleaning test artifacts..."
	@rm -f coverage.out coverage.html

# Справка
help:
	@echo "Available targets:"
	@echo "  test           - Run all integration tests"
	@echo "  test-verbose   - Run tests with verbose output"
	@echo "  test-single    - Run single test (usage: make test-single TEST=TestName)"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  clean          - Remove test artifacts"
	@echo "  help           - Show this help message"
