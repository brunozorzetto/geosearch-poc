#!/bin/bash

# Remove any existing test containers to avoid name conflicts
docker rm -f geosearch-poc-test geosearch-postgres-test 2>/dev/null || true

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Stop and remove existing containers and volumes
docker-compose -f "$SCRIPT_DIR/docker-compose.test.yml" down -v --remove-orphans

# Function to cleanup containers
cleanup() {
    echo "Cleaning up containers..."
    docker-compose -f "$SCRIPT_DIR/docker-compose.test.yml" down --volumes --remove-orphans
}

# Set up trap to ensure cleanup happens on script exit
trap cleanup EXIT

# Build and run tests
docker-compose -f "$SCRIPT_DIR/docker-compose.test.yml" up --build --exit-code-from test

# The cleanup will be automatically called by the trap when the script exits 