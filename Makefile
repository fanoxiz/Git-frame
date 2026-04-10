APP_NAME=gitframe
MAIN_PKG=./cmd/gitframe

.PHONY: build test clean run

build:
	@echo "=== Сборка $(APP_NAME)..."
	go build -o $(APP_NAME) $(MAIN_PKG)

run:
	@echo "=== Запуск $(APP_NAME) ==="
	go run $(MAIN_PKG) $(ARGS)

test:
	@echo "=== Запуск тестов ==="
	go test -v ./tests

clean:
	@echo "=== Очистка ==="
	rm -f $(APP_NAME)
