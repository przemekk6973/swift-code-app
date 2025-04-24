# SWIFT Code App

A Go-based RESTful service for importing, storing and querying international bank SWIFT (BIC) codes. It provides endpoints for adding, retrieving and deleting SWIFT codes. The application can run both locally and in Docker, and uses MongoDB as its backing store.

---

## Table of Contents

- [Features](#features)
- [Tech Stack](#tech-stack)
   - [Go](#go)
   - [MongoDB](#mongodb)
   - [Gin](#gin)
   - [Testcontainers-Go](#testcontainers-go)
   - [Docker & Docker Compose](#docker--docker-compose)
   - [Swagger UI](#swagger-ui)
- [Prerequisites](#prerequisites)
   - [Local Run](#local-run)
   - [Docker Run](#docker-run)
- [Environment Variables](#environment-variables)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
   - [Running Locally](#running-locally)
   - [With Docker Compose](#with-docker-compose)
- [API Reference](#api-reference)
- [Testing](#testing)
   - [Unit Tests](#unit-tests)
   - [Integration Tests](#integration-tests)

---

## Features

**CSV Import**  
Quickly bulk-load SWIFT codes from a spreadsheet. Headquarters (XXX-suffix) and branch codes are split, each entry is validated and normalized (upper-cased), so you can import hundreds of records at once.

**MongoDB Storage**  
Store HQ and branch data together in a document store. Flexible schema and indexes on `swiftCode` and `countryISO2` give you fast lookups without complex joins.

**CRUD REST API**
- **GET** a single SWIFT code (head office + branches or branch only)
- **GET** all codes for a country
- **POST** a new head office or branch
- **DELETE** a head office (and its branches) or a single branch  
  Straightforward endpoints make integration easy.

**Health-check (`/healthz`)**  
Returns HTTP 200 when both the API and MongoDB are up. Ideal for liveness/readiness probes.

**Swagger UI**  
Interactive, always-up-to-date API docs at `/swagger/index.html`. Try requests right in your browser.

**Graceful Shutdown**  
Finishes ongoing requests and closes the MongoDB connection cleanly to avoid data loss on exit.

**Containerization**  
Docker images and a `docker-compose.yml` let you spin up the API and MongoDB with one command—no more “it works on my machine” issues.


---

## Tech Stack

### Go 1.21+
I wrote the API, models and business logic in Go. Its fast compilation and single-binary output let us build and deploy the service easily.

### MongoDB 6.0+
All SWIFT records (headquarters and branches) live in MongoDB. We embed branches in their headquarters documents and index by code and country for quick lookups. Its flexibility allowed each headquarter to include their branches directly, making queries more efficient and avoiding the complexity of joins. It also makes it easier to scale this solution.

### Gin
Gin defines our HTTP routes, parses JSON requests and handles errors. Every endpoint—from fetching a code to deleting a branch—uses Gin’s simple handlers and middleware.

### Testcontainers-Go
Integration tests spin up a real MongoDB in Docker, run imports and API calls, then tear it down. This checks database logic end-to-end without relying on a shared server.

### Docker & Docker Compose
We build a small Docker image for the API and use `docker-compose` to start both the service and MongoDB with a single command, ensuring identical environments everywhere.

### Swagger UI
We use Swaggo annotations in the code to generate an OpenAPI spec and serve it with Gin. Developers can point their browser at `/swagger/index.html` to see and try every endpoint without writing any docs by themselves.

---

## Prerequisites

### Local Run

- **Go 1.21+** installed on your PATH
- **MongoDB 6.0+** running locally (default `mongodb://localhost:27017`)

### Docker Run

- **Docker** (Engine) installed
- **Docker Compose** installed

---

## Environment Variables

Create a `.env` file in the project root:

```ini
MONGO_URI=mongodb://localhost:27017
MONGO_DB=swiftdb
MONGO_COLLECTION=swiftCodes
CSV_PATH=./pkg/data/Interns_2025_SWIFT_CODES.csv
COUNTRIES_CSV=./pkg/data/countries.csv
PORT=8080
```
## Project structure
```
swift-code-app/
│
├── app/                           # Executable entrypoint
│   └── cmd/
│       └── server/                # HTTP server setup
│           └── main.go            # Load config, import CSV, start Gin
│
├── docs/                          # Generated Swagger/OpenAPI specs
│   ├── docs.go                    # Embeds swagger.json/yaml into Go
│   ├── swagger.json               # OpenAPI spec
│   └── swagger.yaml               # OpenAPI spec (YAML)
│
├── integration/                   # Integration API tests (on Mongo running)
│   └── api_test.go                # Starts container, exercises all endpoints
│
├── internal/                      # Core application code
│   ├── adapter/
│   │   ├── api/
│   │   │   └── v1/                # Versioned HTTP handlers
│   │   │       ├── swift_handler.go
│   │   │       └── swift_handler_test.go
│   │   └── persistence/           # MongoDB repository implementation
│   │       ├── mongo_repo.go
│   │       └── mongo_repo_test.go
│   │
│   ├── domain/
│   │   ├── models/                # Data and response models
│   │   │   ├── swiftcode.go
│   │   │   ├── swiftbranch.go
│   │   │   ├── import_summary.go
│   │   │   └── country_swift_codes_response.go
│   │   └── usecases/              # Business logic / service layer
│   │       ├── swift_usecase.go
│   │       └── swift_usecase_test.go
│   │
│   ├── initializer/               # CSV import
│   │   ├── initializer.go
│   │   └── initializer_test.go
│   │
│   ├── port/                      # Interface definitions
│   │   └── repository.go
│   │
│   └── util/                      # Helpers & validation
│       ├── csv.go
│       ├── csv_test.go
│       ├── countries.go
│       ├── errors.go
│       ├── errors_test.go
│       ├── params.go
│       ├── validator.go
│       └── validator_test.go
│
├── pkg/
│   └── data/                      # Sample CSV files
│       ├── countries.csv
│       └── Interns_2025_SWIFT_CODES.csv
│
├── test/                          # Additional test fixtures/helpers
│
├── .env                           # Environment variables (MONGO_URI, CSV_PATH)
├── .gitignore
├── docker-compose.yml             # API + MongoDB development stack
├── Dockerfile                     # Multi-stage build for production image
├── go.mod                         # Go module declarations
├── go.sum                         # Dependency checksums
└── README.md                      # Project overview & instructions
```

## Getting Started

There are two ways to run the application: locally with Go and MongoDB, or using Docker Compose for a fully containerized setup.

---

### Running Locally

This method runs the Go application directly and connects to a locally running MongoDB instance.

#### 1. Install dependencies

```bash
go mod tidy
```
#### 2. Create a .env file
Create this file in the main (swift-code-app) folder. Configure it like this:

```ini
MONGO_URI=mongodb://localhost:27017
MONGO_DB=swiftdb
MONGO_COLLECTION=swiftCodes
CSV_PATH=./pkg/data/Interns_2025_SWIFT_CODES.csv
COUNTRIES_CSV=./pkg/data/countries.csv
PORT=8080
```

- `MONGO_URI`  
  Connection string for MongoDB

- `MONGO_DB`  
  Name of the MongoDB database to use

- `MONGO_COLLECTION`  
  Name of the collection where SWIFT codes are stored

- `CSV_PATH`  
  File path to the SWIFT codes CSV to import on startup

- `COUNTRIES_CSV`  
  File path to the countries lookup CSV (ISO2 : country name)

- `PORT`  
  TCP port where the HTTP server listens
- 
#### 3. Start MongoDB locally (if not already running)
#### 4. Run the app
```bash
go run app/cmd/server/main.go
```
The app will run at http://localhost:8080/ and Swagger UI will be available at: http://localhost:8080/swagger/index.html

## Running with Docker Compose (recommended)

This setup uses Docker to run both the API and MongoDB in containers.

### 1. Make sure Docker & Docker Compose are installed

### 2. Configure .env

Create this file in the main (swift-code-app) folder. It will be used by docker-compose.yml. Configure it like this:

```ini
MONGO_URI=mongodb://localhost:27017
MONGO_DB=swiftdb
MONGO_COLLECTION=swiftCodes
CSV_PATH=./pkg/data/Interns_2025_SWIFT_CODES.csv
COUNTRIES_CSV=./pkg/data/countries.csv
PORT=8080
```
### 3. Start everything

```bash
docker-compose up --build
```
**This will:**
- Build the Go API image
- Start MongoDB
- Import data from CSV (if CSV_PATH is set)

## API Reference

All endpoints are versioned under `/v1/swift-codes`.  
Requests and responses use JSON format.

---

### GET `/v1/swift-codes/{swiftCode}`

Returns full details for a SWIFT code.

- If it's a **headquarter code** (ends in `XXX`), returns its data plus all branches.
- If it's a **branch code**, returns only that branch.

#### Example (HQ):

```
{
  "address": "string",
  "bankName": "string",
  "countryISO2": "string",
  "countryName": "string",
  "isHeadquarter": true,
  "swiftCode": "string",
  "branches": [
    {
      "address": "string",
      "bankName": "string",
      "countryISO2": "string",
      "isHeadquarter": false,
      "swiftCode": "string"
    }
  ]
}
```


#### Example (Branch):
```
{
  "address": "string",
  "bankName": "string",
  "countryISO2": "string",
  "isHeadquarter": false,
  "swiftCode": "string"
}
```

#### Usage example (using curl)
```
curl -i http://localhost:8080/v1/swift-codes/*Swiftcode*
```

### GET `/v1/swift-codes/country/{countryISO2code}`

Returns all SWIFT codes for the given country (both HQ and branches).

```
{
  "countryISO2": "string",
  "countryName": "string",
  "swiftCodes": [
    {
      "address": "string",
      "bankName": "string",
      "countryISO2": "string",
      "isHeadquarter": false,
      "swiftCode": "string"
    }
  ]
}
```
#### Usage example (using curl)
```
curl -i http://localhost:8080/v1/swift-codes/country/*ISO2*
```



### POST `/v1/swift-codes`

Adds a new SWIFT code. Can be either a headquarter or branch.

```
{
  "address": "string",
  "bankName": "string",
  "countryISO2": "string",
  "countryName": "string",
  "isHeadquarter": boolean,
  "swiftCode": "string"
}
```
#### Response:

```
{
  "message": "swift code added"
}
```
Returns `409 Conflict` if the SWIFT code already exists.

#### Usage example (using curl)
```
curl -i -X POST http://localhost:8080/v1/swift-codes \
  -H "Content-Type: application/json" \
  -d '{
    "address":       "string",
    "bankName":      "string",
    "countryISO2":   "string",
    "countryName":   "string",
    "isHeadquarter": bool,
    "swiftCode":     "string"
  }' 
  ```

### DELETE `/v1/swift-codes/{swiftCode}`

Deletes a SWIFT code.
- If it's a headquarter, also deletes all its branches.
- If it's a branch, only that branch is removed.

#### Response:
```
{
  "message": "swift code deleted"
}
```

#### Response (200 OK):

```
{
  "status": "ok"
}
```
#### Usage example (using curl)
```
curl -i -X DELETE http://localhost:8080/v1/swift-codes/*Swiftcode*
```

### Swagger Documentation

Available at:
http://localhost:8080/swagger/index.html

Here you can:
- Try out all endpoints
- See request/response models
- Explore status codes and descriptions

## Testing

This project includes full test coverage:

- Unit tests for logic, validation, handlers
- Integration tests with MongoDB containers

---

### Run All Tests

#### Unit tests
You can run all unit tests in the project with a single command:

```bash
go test ./... -v
```

#### Integration tests

Before running integration tests, ensure you have a MongoDB instance available on localhost:27017. You can do this with Docker:

```bash
docker run -d --name mongo-test -p 27017:27017 mongo:6
```

If there is already container named `mongo` with required database, you can run this command:

```bash
docker start mongo
```

Then, from the project root, run:

```bash
go test ./app/integration -v
```
The test will:
- Connect to `mongodb://localhost:27017`
- Import a small in-memory CSV into a temporary database
- Exercise all CRUD endpoints via Gin’s router
- Report pass/fail results

## Test Coverage

The project includes both unit and integration tests to check core functionality and overall API behavior.

| Module / Location             | Description                                         |
|-------------------------------|-----------------------------------------------------|
| `pkg/csv`                     | Reading and parsing the CSV file                    |
| `internal/util`               | Validation helpers (country codes, SWIFT format)    |
| `internal/domain/usecases`    | Business logic (adding, fetching, deleting codes)   |
| `database/`                   | MongoDB setup and data operations                   |
| `cmd/router`                  | Route setup and HTTP request handling               |
| `app/internal/adapter/api/v1` | API handlers and JSON responses                     |
| `app/integration`             | End-to-end tests of the API with a live database    |


