# Swift-App (Rozwiązanie B)

Aplikacja do zarządzania kodami SWIFT – czysta architektura, MongoDB, Gin, testy jednostkowe i integracyjne.

## Wymagania

- Go 1.21+
- Docker & Docker Compose (opcjonalnie)

## Uruchomienie lokalne

1. Skopiuj `.env.example` do `.env` i uzupełnij zmienne.
2. Uruchom serwer Mongo:
   ```bash
   docker-compose up -d mongo