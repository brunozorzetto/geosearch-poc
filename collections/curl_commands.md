# GeoSearch POC API - cURL Commands

Base URL: `http://localhost:8080/api/v1`

## Stores Endpoints

### 1. Create Store
```bash
curl -X POST "http://localhost:8080/api/v1/stores" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Loja Exemplo",
    "latitude": -23.550520,
    "longitude": -46.633308,
    "address": "Av. Paulista, 1000",
    "delivery_radius": 5.0
  }'
```

### 2. Create Store with Category
```bash
curl -X POST "http://localhost:8080/api/v1/stores" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Loja com Categoria",
    "category_id": "550e8400-e29b-41d4-a716-446655440000",
    "latitude": -23.551520,
    "longitude": -46.634308,
    "address": "Rua Augusta, 500",
    "delivery_radius": 3.0
  }'
```

### 3. Get Store by ID
```bash
curl -X GET "http://localhost:8080/api/v1/stores/{store_id}"
```

### 4. Update Store
```bash
curl -X PUT "http://localhost:8080/api/v1/stores/{store_id}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Loja Atualizada",
    "latitude": -23.560520,
    "longitude": -46.643308,
    "address": "Rua Atualizada, 456",
    "delivery_radius": 7.0
  }'
```

### 5. Delete Store
```bash
curl -X DELETE "http://localhost:8080/api/v1/stores/{store_id}"
```

## Products Endpoints

### 6. Create Product
```bash
curl -X POST "http://localhost:8080/api/v1/products" \
  -H "Content-Type: application/json" \
  -d '{
    "store_id": "{store_id}",
    "name": "Produto Exemplo",
    "description": "Descrição do produto exemplo",
    "price": 29.99,
    "category": "Electronics",
    "brand": "Marca Exemplo",
    "sku": "PROD-001",
    "stock": 100,
    "h3_index": "8928308280fffff",
    "images": [
      "https://example.com/image1.jpg",
      "https://example.com/image2.jpg"
    ]
  }'
```

### 7. Create Product (Minimal)
```bash
curl -X POST "http://localhost:8080/api/v1/products" \
  -H "Content-Type: application/json" \
  -d '{
    "store_id": "{store_id}",
    "name": "Produto Mínimo",
    "price": 19.99,
    "sku": "PROD-MIN-001",
    "stock": 50
  }'
```

### 8. Get Product by ID
```bash
curl -X GET "http://localhost:8080/api/v1/products/{product_id}"
```

### 9. Update Product
```bash
curl -X PUT "http://localhost:8080/api/v1/products/{product_id}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Produto Atualizado",
    "description": "Nova descrição do produto",
    "price": 39.99,
    "category": "Updated Electronics",
    "brand": "Updated Brand",
    "sku": "UPDATED-SKU-001",
    "stock": 150,
    "h3_index": "8928308280fffff",
    "images": [
      "https://example.com/updated-image.jpg"
    ]
  }'
```

### 10. Update Product (Partial)
```bash
curl -X PUT "http://localhost:8080/api/v1/products/{product_id}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Produto Parcialmente Atualizado",
    "price": 49.99,
    "stock": 200
  }'
```

### 11. Delete Product
```bash
curl -X DELETE "http://localhost:8080/api/v1/products/{product_id}"
```

## Search Endpoints

### 12. Search Stores by Location
```bash
curl -X GET "http://localhost:8080/api/v1/search?lat=-23.550520&lng=-46.633308&radius=5.0"
```

### 13. Search Stores (1km radius)
```bash
curl -X GET "http://localhost:8080/api/v1/search?lat=-23.550520&lng=-46.633308&radius=1.0"
```

### 14. Search Products
```bash
curl -X GET "http://localhost:8080/api/v1/products/search?q=smartphone&lat=-23.550520&lng=-46.633308&radius=5.0"
```

### 15. Search Products with Filters
```bash
curl -X GET "http://localhost:8080/api/v1/products/search?q=electronics&lat=-23.550520&lng=-46.633308&radius=5.0&category=Electronics&brand=Apple&min_price=100&max_price=1000&page_size=20"
```

## Error Examples

### 16. Create Store - Missing Required Fields
```bash
curl -X POST "http://localhost:8080/api/v1/stores" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Loja sem campos obrigatórios"
  }'
```

### 17. Create Store - Invalid Category ID
```bash
curl -X POST "http://localhost:8080/api/v1/stores" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Loja com categoria inválida",
    "category_id": "invalid-uuid",
    "latitude": -23.550520,
    "longitude": -46.633308,
    "address": "Rua Teste, 123",
    "delivery_radius": 5.0
  }'
```

### 18. Create Product - Missing Required Fields
```bash
curl -X POST "http://localhost:8080/api/v1/products" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Produto sem campos obrigatórios",
    "price": 29.99
  }'
```

### 19. Create Product - Invalid Store ID
```bash
curl -X POST "http://localhost:8080/api/v1/products" \
  -H "Content-Type: application/json" \
  -d '{
    "store_id": "invalid-uuid",
    "name": "Produto com store inválida",
    "price": 29.99,
    "sku": "PROD-ERROR-001",
    "stock": 100
  }'
```

### 20. Get Store - Invalid UUID
```bash
curl -X GET "http://localhost:8080/api/v1/stores/invalid-uuid"
```

### 21. Get Store - Non-existent
```bash
curl -X GET "http://localhost:8080/api/v1/stores/00000000-0000-0000-0000-000000000000"
```

### 22. Search Stores - Missing Parameters
```bash
curl -X GET "http://localhost:8080/api/v1/search?lat=-23.550520"
```

### 23. Search Products - Missing Query
```bash
curl -X GET "http://localhost:8080/api/v1/products/search?lat=-23.550520&lng=-46.633308&radius=5.0"
```

## Usage Notes

- Replace `{store_id}` with an actual store UUID when testing
- Replace `{product_id}` with an actual product UUID when testing
- The base URL assumes the API is running on `localhost:8080`
- All requests use JSON content type for POST/PUT operations
- Error examples are designed to test validation and error handling

## Import to Postman

To import these cURL commands into Postman:

1. Copy any cURL command above
2. In Postman, click "Import" 
3. Select "Raw text" tab
4. Paste the cURL command
5. Click "Continue" and then "Import"
6. The request will be created with all headers and body data
7. Update the `{store_id}` and `{product_id}` placeholders with actual values 