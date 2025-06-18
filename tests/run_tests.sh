#!/bin/bash

# Test execution script for GeoSearch POC
# This script runs all tests in the project

set -e

echo "🧪 Running GeoSearch POC Tests"
echo "================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    print_error "go.mod not found. Please run this script from the project root."
    exit 1
fi

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker and try again."
    exit 1
fi

print_status "Stopping any existing test containers..."
docker compose -f tests/docker-compose.test.yml down -v

print_status "Starting test database..."
docker compose -f tests/docker-compose.test.yml up -d

# Wait for database to be ready
print_status "Waiting for database to be ready..."
sleep 10

# Run database migrations
print_status "Running database migrations..."
export DATABASE_URL="postgres://postgres:postgres@localhost:5433/geosearch_test?sslmode=disable"
go run cmd/migrate/main.go

# Run unit tests (these don't need database)
print_status "Running unit tests..."
go test -v ./tests/unit/...

# Run integration tests separately to avoid interference
print_status "Running integration tests..."
go test -v ./tests/integration/...

# Run tests with coverage separately for each package to avoid interference
print_status "Running tests with coverage..."

# Run handlers tests with coverage
print_status "Running handlers tests with coverage..."
go test -v -coverprofile=coverage_handlers.out ./tests/integration/handlers/...

# Run repository tests with coverage
print_status "Running repository tests with coverage..."
go test -v -coverprofile=coverage_repository.out ./tests/integration/repository/...

# Run services tests with coverage
print_status "Running services tests with coverage..."
go test -v -coverprofile=coverage_services.out ./tests/integration/services/...

# Run unit tests with coverage
print_status "Running unit tests with coverage..."
go test -v -coverprofile=coverage_unit.out ./tests/unit/...

# Merge coverage files
print_status "Merging coverage reports..."
echo "mode: set" > coverage.out
tail -n +2 coverage_handlers.out >> coverage.out
tail -n +2 coverage_repository.out >> coverage.out
tail -n +2 coverage_services.out >> coverage.out
tail -n +2 coverage_unit.out >> coverage.out

# Clean up individual coverage files
rm coverage_handlers.out coverage_repository.out coverage_services.out coverage_unit.out

# Generate coverage report
if command -v go tool cover > /dev/null 2>&1; then
    print_status "Generating coverage report..."
    go tool cover -html=coverage.out -o coverage.html
    print_status "Coverage report generated: coverage.html"
else
    print_warning "go tool cover not available, skipping HTML coverage report"
fi

# Show coverage summary
print_status "Coverage summary:"
go tool cover -func=coverage.out

# Cleanup
print_status "Cleaning up test database..."
docker compose -f tests/docker-compose.test.yml down -v

print_status "All tests completed successfully! 🎉" 