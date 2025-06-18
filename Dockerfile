# Build stage
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Install git and build dependencies
RUN apk add --no-cache git build-base cmake

# Build H3 from source
RUN git clone https://github.com/uber/h3.git /tmp/h3 && \
    cd /tmp/h3 && \
    cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX=/usr . && \
    make && \
    make install

# Copy go mod files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .
COPY migrations ./migrations

# Build the application with CGO enabled
RUN go build -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates and build dependencies for H3
RUN apk --no-cache add ca-certificates build-base cmake git && \
    git clone https://github.com/uber/h3.git /tmp/h3 && \
    cd /tmp/h3 && \
    cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX=/usr . && \
    make && \
    make install && \
    rm -rf /tmp/h3

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["./main"] 