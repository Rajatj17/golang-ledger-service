# Banking Ledger System

A Go-based banking ledger system that provides REST APIs for account and transaction management with PostgreSQL and MongoDB storage, RabbitMQ messaging, and a distributed worker architecture.

## Architecture

- **Server**: REST API server handling account and transaction requests
- **Worker Service**: Background service processing queued transactions
- **PostgreSQL**: Account data storage
- **MongoDB**: Transaction log storage
- **RabbitMQ**: Message queue for async transaction processing

## Prerequisites

- Go 1.24.2+
- Docker and Docker Compose
- PostgreSQL
- MongoDB
- RabbitMQ

## Quick Start

1. **Clone and setup**:
   ```bash
   git clone <repository-url>
   cd golang-ledger-service
   ```

2. **Start infrastructure with Docker**:
   ```bash
   docker-compose up -d
   ```

3. **Run the API server**:
   ```bash
   go run cmd/api/main.go
   ```

4. **Run the worker (in another terminal)**:
   ```bash
   go run cmd/worker/main.go
   ```

## Configuration

Configuration is managed via `config.yaml`:

```yaml
env: development
app:
  port: 8080
  name: ledger-management
db:
  postgres:
    host: localhost
    port: 5432
    username: postgres
    password: postgres
    name: ledger_service
  mongo:
    uri: mongodb://localhost:27017/transaction_logs
rabbitmq:
  host: localhost
  port: 5672
  user_name: guest
  password: guest
  queue: ledger_queue
```

## API Endpoints

### Health Check
- `GET /` - API health status

### Accounts
- `POST /api/v1/accounts` - Create account
- `GET /api/v1/accounts/:id` - Get account by ID
- `GET /api/v1/accounts/:id/balance` - Get account balance by ID
- `POST /api/v1/accounts/fund` - Deposit Or Withdraw


### Transactions
- `GET /api/v1/transactions/:id` - Get transaction by ID
- `GET /api/v1/transactions` - List transactions

## Testing

The project includes integration and end-to-end tests:

```bash
# Run all tests
go test ./...

# Run specific test suites
go test ./tests/integration/
go test ./tests/e2e/
```

### API Testing

A Postman collection is available for manual API testing. Import the collection file into Postman to test all available endpoints with pre-configured requests and examples.

## Development

### Project Structure
```
├── cmd/
│   ├── api/          # API server entry point
│   └── worker/       # Worker service entry point
├── config/           # Configuration management
├── internal/
│   ├── database/     # Database connections and models
│   ├── dto/          # Data transfer objects
│   ├── handler/      # HTTP handlers
│   ├── messaging/    # RabbitMQ messaging
│   ├── middleware/   # HTTP middleware
│   ├── repository/   # Data access layer
│   ├── router/       # Route definitions
│   └── service/      # Business logic
└── tests/           # Test suites
```

### Dependencies

Key dependencies include:
- **Gin**: HTTP web framework
- **GORM**: ORM for PostgreSQL
- **MongoDB Driver**: Native MongoDB driver
- **RabbitMQ (AMQP)**: Message queue client

## Docker Support

The project includes Docker configurations:

- `Dockerfile.api` - API server container
- `Dockerfile.worker` - Worker service container
- `docker-compose.yaml` - Complete stack orchestration

Build and run with Docker:
```bash
docker-compose up --build
```# golang-ledger-service
