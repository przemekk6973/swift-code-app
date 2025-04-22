# Stage 1: build
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o swift-api app/cmd/server/main.go

# Stage 2: runtime
FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /root/
COPY --from=builder /app/swift-api .
# Jeśli masz pliki CSV + countries.csv w pkg/data, kopiuj je:
COPY --from=builder /app/pkg/data/countries.csv        ./data/
COPY --from=builder /app/pkg/data/Interns_2025_SWIFT_CODES.csv ./data/

# Ustaw domyślne ENV
ENV MONGO_URI="mongodb://mongo:27017" \
    MONGO_DB="swiftdb" \
    MONGO_COLLECTION="swiftCodes" \
    CSV_PATH="/root/data/Interns_2025_SWIFT_CODES.csv" \
    COUNTRIES_CSV="/root/data/countries.csv" \
    PORT="8080"
EXPOSE 8080
ENTRYPOINT ["./swift-api"]
