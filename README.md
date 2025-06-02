# Hiper GeoSearch

Hiper GeoSearch is a Go-based HTTP service for geolocation-based store and product search, using the Gin framework and Uber's H3 for spatial indexing. The project follows Clean Architecture, SOLID principles, modular design, and comprehensive testing (unit/integration). Security best practices (OWASP) are encouraged throughout the codebase.

## Project Status
- **Handlers** for health check and search are centralized in the `handlers/` folder.
- **Domain models** for stores, products, and categories are in `domain/`.
- **Repositories** for PostgreSQL are in `repository/postgres/`.
- **Service** layer for search logic is in `service/`.
- **H3 utilities** are in `pkg/h3/`.
- **Distance calculations** are in `pkg/distance/`.
- **Integration tests** are in `tests/integration/`.
- **Database migrations** are in `migrations/`.
- **Docker** is used for local development and database setup.

All integration tests for the product repository are passing and cover creation, retrieval, update, deletion, and pagination, including foreign key constraints.

## Requirements
- Go 1.23+
- Docker & Docker Compose
- PostgreSQL with PostGIS extension

## Project Structure
```
.
├── cmd/                    # Application entry points
├── config/                 # Configuration management
├── domain/                 # Domain models and interfaces
├── handlers/              # HTTP handlers (controllers)
├── migrations/            # Database migration files
├── pkg/                   # Shared packages
│   ├── distance/         # Distance calculation utilities
│   └── h3/              # H3 geospatial utilities
├── repository/           # Repository interfaces and implementations
│   └── postgres/        # PostgreSQL implementations
├── service/             # Business logic layer
├── tests/               # Test files
│   └── integration/    # Integration tests
│       ├── config/     # Test configurations
│       └── repository/ # Repository tests
├── .env                 # Environment variables (development)
├── .env.example        # Example environment variables
├── .env.testing        # Environment variables for tests
├── docker-compose.yml  # Docker services configuration
├── Dockerfile         # Application container definition
├── go.mod            # Go module definition
└── README.md         # Project documentation
```

## Setup

### 1. Clone the repository
```bash
git clone <repo-url>
cd geosearch-poc
```

### 2. Configure environment variables
- Copy `.env.example` to `.env` for development:
```bash
cp .env.example .env
```

- Copy `.env.example` to `.env.testing` for tests:
```bash
cp .env.example .env.testing
```

Edit both files to match your environment settings.

### 3. Start the database
```bash
docker-compose up -d
```
This will start a PostgreSQL instance with PostGIS.

### 4. Apply database migrations
You can use a migration tool (e.g., [golang-migrate](https://github.com/golang-migrate/migrate)) or run the SQL scripts manually:

```bash
# For development database
psql -h localhost -U postgres -d geosearch < migrations/001_initial_schema.sql

# For test database
psql -h localhost -U postgres -d geosearch_test < migrations/001_initial_schema.sql
```

### 5. Run the application
```bash
go run main.go
```
The server will be available at `http://localhost:8080` (or the port set in your `.env`).

## Testing

### Running Tests

1. **Ensure test environment is ready:**
   - PostgreSQL is running (`docker-compose up -d`)
   - `.env.testing` is configured correctly
   - Test database migrations are applied

2. **Run all tests:**
```bash
go test ./... -v
```

3. **Run only integration tests:**
```bash
go test ./tests/integration/... -v
```

4. **Run specific test package:**
```bash
# Run repository tests
go test ./tests/integration/repository/... -v

# Run specific repository tests
go test ./tests/integration/repository/postgres/... -v
```

### Test Coverage

To generate test coverage report:
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Development

### Adding New Tests

1. **Integration Tests:**
   - Place new integration tests in `tests/integration/`
   - Follow the existing pattern of test organization
   - Ensure proper setup and teardown of test data

2. **Unit Tests:**
   - Place unit tests alongside the code they test
   - Use the `_test.go` suffix
   - Follow Go's standard testing patterns

### Best Practices

- Write tests before implementing features (TDD)
- Keep tests focused and isolated
- Use meaningful test names
- Clean up test data after each test
- Use test fixtures when appropriate
- Mock external dependencies

## License

[Add your license here] 