.PHONY: build run test lint docker compose-up

# Buduje binarkę w katalogu bin/
build:
	go build -o bin/swift-app ./cmd

# Uruchamia aplikację po kompilacji
run: build
	./bin/swift-app

# Testy jednostkowe
test:
	go test ./...

# Linter
tlint:
	golangci-lint run

# Buduje obraz Dockera
docker:
	docker build -t yourdocker/swift-app:latest .

# Uruchamia Docker Compose (API + Mongo)
compose-up:
	docker-compose up --build