# Hiper GeoSearch

A simple HTTP service built with Go and Gin.

## Project Structure

```
.
├── handlers/         # HTTP handlers (controllers)
│   └── health.go
├── main.go           # Application entry point and routes
├── Dockerfile        # Docker build file
├── go.mod            # Go module definition
└── README.md         # Project documentation
```

## Prerequisites

- Go 1.22+
- Docker
- (Optional) Docker Compose

## Running Locally

1. Instale as dependências:

```bash
go mod tidy
```

2. Execute o servidor:

```bash
go run main.go
```

3. Teste o endpoint de health:

```bash
curl http://localhost:8080/health
```

## Rodando com Docker

1. Construa a imagem:

```bash
docker build -t geosearch-poc .
```

2. Suba o container:

```bash
docker run -p 8080:8080 geosearch-poc
```

## Rodando com Docker Compose

1. Crie um arquivo `docker-compose.yml` com o conteúdo:

```yaml
version: '3.8'
services:
  app:
    build: .
    container_name: geosearch-poc
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
```

2. Suba o serviço:

```bash
docker compose up --force-recreate
```

## Endpoints

- `GET /health` — Health check endpoint

## License

[Add your license here] 