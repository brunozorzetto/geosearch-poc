# GeoSearch POC API - Collection do Postman

Esta collection do Postman contém todos os endpoints da API GeoSearch POC para facilitar os testes e desenvolvimento.

## 📋 Índice

- [Instalação](#instalação)
- [Configuração](#configuração)
- [Estrutura da Collection](#estrutura-da-collection)
- [Como Usar](#como-usar)
- [Endpoints Disponíveis](#endpoints-disponíveis)
- [Exemplos de Uso](#exemplos-de-uso)
- [Variáveis de Ambiente](#variáveis-de-ambiente)
- [Testes Automatizados](#testes-automatizados)

## 🚀 Instalação

1. **Baixe o arquivo da collection:**
   - `GeoSearch_POC_API.postman_collection.json`

2. **Importe no Postman:**
   - Abra o Postman
   - Clique em "Import" (canto superior esquerdo)
   - Arraste o arquivo ou clique em "Upload Files"
   - Selecione o arquivo `GeoSearch_POC_API.postman_collection.json`

## ⚙️ Configuração

### Pré-requisitos

1. **Servidor rodando:**
   ```bash
   # Certifique-se de que o servidor está rodando na porta 8080
   go run main.go
   ```

2. **Banco de dados configurado:**
   ```bash
   # Execute as migrações
   make migrate-up
   ```

3. **Variáveis de ambiente:**
   - Crie um arquivo `.env` com as configurações necessárias
   - Verifique se o `DATABASE_URL` está correto

### Configuração das Variáveis

A collection usa as seguintes variáveis:

| Variável | Valor Padrão | Descrição |
|----------|---------------|-----------|
| `base_url` | `http://localhost:8080/api/v1` | URL base da API |
| `store_id` | (vazio) | ID da store criada (preenchido automaticamente) |
| `product_id` | (vazio) | ID do produto criado (preenchido automaticamente) |
| `category_id` | `550e8400-e29b-41d4-a716-446655440000` | ID de categoria de exemplo |

## 📁 Estrutura da Collection

A collection está organizada em 4 pastas principais:

### 1. **Stores** 📍
Endpoints para gerenciamento de stores (lojas):
- `POST /stores` - Criar store
- `GET /stores/:id` - Buscar store por ID
- `PUT /stores/:id` - Atualizar store
- `DELETE /stores/:id` - Deletar store

### 2. **Products** 📦
Endpoints para gerenciamento de produtos:
- `POST /products` - Criar produto
- `GET /products/:id` - Buscar produto por ID
- `PUT /products/:id` - Atualizar produto
- `DELETE /products/:id` - Deletar produto

### 3. **Search** 🔍
Endpoints para busca geográfica e de produtos:
- `GET /search` - Buscar stores por localização
- `GET /products/search` - Buscar produtos com Vertex AI

### 4. **Error Examples** ❌
Exemplos de requisições que geram erros para testar validações.

## 🎯 Como Usar

### Fluxo Básico de Teste

1. **Criar uma Store:**
   - Execute "Create Store" ou "Create Store with Category"
   - Copie o `id` da resposta
   - Atualize a variável `store_id` com este valor

2. **Criar um Produto:**
   - Execute "Create Product" ou "Create Product (Minimal)"
   - Use o `store_id` configurado anteriormente
   - Copie o `id` da resposta
   - Atualize a variável `product_id` com este valor

3. **Testar Busca:**
   - Execute "Search Stores by Location"
   - Execute "Search Products"

4. **Testar Atualizações:**
   - Execute "Update Store" ou "Update Product"
   - Verifique se os dados foram atualizados

5. **Testar Deleção:**
   - Execute "Delete Product" e "Delete Store"
   - Verifique se os recursos foram removidos

### Atualizando Variáveis Automaticamente

Para facilitar os testes, você pode configurar scripts para atualizar automaticamente as variáveis:

1. **Para stores criadas:**
   ```javascript
   // Adicione este script no evento "test" do endpoint "Create Store"
   if (pm.response.code === 201) {
       const response = pm.response.json();
       pm.collectionVariables.set("store_id", response.id);
   }
   ```

2. **Para produtos criados:**
   ```javascript
   // Adicione este script no evento "test" do endpoint "Create Product"
   if (pm.response.code === 201) {
       const response = pm.response.json();
       pm.collectionVariables.set("product_id", response.id);
   }
   ```

## 📡 Endpoints Disponíveis

### Stores

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| POST | `/stores` | Criar nova store |
| GET | `/stores/:id` | Buscar store por ID |
| PUT | `/stores/:id` | Atualizar store |
| DELETE | `/stores/:id` | Deletar store |

### Products

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| POST | `/products` | Criar novo produto |
| GET | `/products/:id` | Buscar produto por ID |
| PUT | `/products/:id` | Atualizar produto |
| DELETE | `/products/:id` | Deletar produto |

### Search

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| GET | `/search` | Buscar stores por localização |
| GET | `/products/search` | Buscar produtos com Vertex AI |

## 💡 Exemplos de Uso

### Criando uma Store

```json
POST /api/v1/stores
{
  "name": "Loja Exemplo",
  "latitude": -23.550520,
  "longitude": -46.633308,
  "address": "Av. Paulista, 1000",
  "delivery_radius": 5.0
}
```

### Criando um Produto

```json
POST /api/v1/products
{
  "store_id": "{{store_id}}",
  "name": "Produto Exemplo",
  "description": "Descrição do produto exemplo",
  "price": 29.99,
  "category": "Electronics",
  "brand": "Marca Exemplo",
  "sku": "PROD-001",
  "stock": 100,
  "images": [
    "https://example.com/image1.jpg"
  ]
}
```

### Buscando Stores por Localização

```
GET /api/v1/search?lat=-23.550520&lng=-46.633308&radius=5.0
```

### Buscando Produtos

```
GET /api/v1/products/search?q=smartphone&lat=-23.550520&lng=-46.633308&radius=5.0&category=Electronics&min_price=100&max_price=1000
```

## 🔧 Variáveis de Ambiente

### Configuração Local

Para desenvolvimento local, use estas configurações:

```json
{
  "base_url": "http://localhost:8080/api/v1",
  "store_id": "",
  "product_id": "",
  "category_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Configuração de Produção

Para ambiente de produção, ajuste a `base_url`:

```json
{
  "base_url": "https://api.geosearch.com/api/v1",
  "store_id": "",
  "product_id": "",
  "category_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

## 🧪 Testes Automatizados

A collection inclui testes automatizados que verificam:

- ✅ Códigos de status válidos
- ✅ Tempo de resposta < 5 segundos
- ✅ Resposta JSON válida
- ✅ Ausência de erros em respostas de sucesso
- ✅ Presença de mensagem de erro em respostas de erro

### Executando Todos os Testes

1. **No Postman:**
   - Clique com botão direito na collection
   - Selecione "Run collection"
   - Configure as variáveis necessárias
   - Clique em "Run"

2. **Via Newman (CLI):**
   ```bash
   # Instalar Newman
   npm install -g newman
   
   # Executar collection
   newman run GeoSearch_POC_API.postman_collection.json
   
   # Com relatório HTML
   newman run GeoSearch_POC_API.postman_collection.json -r html
   ```

## 🐛 Troubleshooting

### Problemas Comuns

1. **Erro de conexão:**
   - Verifique se o servidor está rodando na porta 8080
   - Confirme se a `base_url` está correta

2. **Erro de foreign key constraint:**
   - Certifique-se de criar a store antes do produto
   - Verifique se o `store_id` está correto

3. **Erro de UUID inválido:**
   - Use UUIDs válidos no formato: `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`
   - Verifique se as variáveis estão sendo atualizadas corretamente

4. **Erro de parâmetros obrigatórios:**
   - Para stores: `name`, `latitude`, `longitude` são obrigatórios
   - Para produtos: `store_id`, `name`, `price`, `sku`, `stock` são obrigatórios

### Logs e Debug

Para habilitar logs detalhados:

1. **No Postman:**
   - Abra o Console (View > Show Postman Console)
   - Execute as requisições
   - Verifique os logs de request/response

2. **No servidor:**
   ```bash
   # Execute com logs detalhados
   GIN_MODE=debug go run main.go
   ```

## 📚 Recursos Adicionais

- [Documentação da API](README.md)
- [Testes de Integração](tests/integration/)
- [Testes Unitários](tests/unit/)
- [Script de Testes](tests/run_tests.sh)

## 🤝 Contribuição

Para contribuir com melhorias na collection:

1. Teste todos os endpoints
2. Adicione novos casos de teste
3. Atualize a documentação
4. Submeta um pull request

---

**Nota:** Esta collection é específica para a API GeoSearch POC. Certifique-se de que o servidor está configurado corretamente antes de usar. 