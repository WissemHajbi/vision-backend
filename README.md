# Vision Products API

A deliberately small Go API that teaches the usual backend layers while storing products in SQLite.

## API

### Find a product

```http
GET /api/v1/products?qrcode=6191234567890
```

Success (`200`):

```json
{"id":1,"qr_code":"6191234567890","name":"Coca Cola 1.5L","price":3.5}
```

An unknown code returns `404`.

### Create a product

```http
POST /api/v1/products
Content-Type: application/json

{"qr_code":"6191234567890","name":"Coca Cola 1.5L","price":3.5}
```

Success returns `201`. A duplicate code returns `409`. Prices accept at most two decimal places and are stored as integer cents.

## Structure

```text
cmd/api/main.go                  application entry point and dependency wiring
internal/config/                 environment configuration
internal/database/               SQLite setup and embedded SQL migrations
internal/product/                entity, business rules, repository
internal/httpapi/                HTTP routes and JSON translation
```

`internal` prevents other Go modules from importing implementation details. Dependencies point inward:

```text
HTTP handler -> product service -> repository interface <- SQLite repository
```

This is enough separation to test and learn from without introducing a framework.

## Run locally

Requires Go 1.26:

```bash
go run ./cmd/api
```

The default database is `./data/products.db` and the API listens on port `8080`.

Test it:

```bash
curl -i -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{"qr_code":"6191234567890","name":"Coca Cola 1.5L","price":3.5}'

curl -i "http://localhost:8080/api/v1/products?qrcode=6191234567890"
```

Run tests:

```bash
go test ./...
```

## Docker

```bash
docker compose up --build -d
docker compose logs -f
docker compose down
```

The named volume `product_data` preserves SQLite data when the container is replaced. Back it up from the VPS regularly.

For public deployment, place this API behind Caddy or Nginx with HTTPS. This learning version has no authentication, so exposing `POST` publicly lets anyone create products.

## Connect the Vision app

The neighboring `vision-app` reads its server address from Expo environment variables:

```bash
# vision-app/.env
EXPO_PUBLIC_API_URL=https://products.example.com
```

Restart Expo after changing it. The fallback `http://10.0.2.2:8080` reaches this API from an Android emulator; a physical phone needs your computer's LAN address or the HTTPS VPS address.

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `HTTP_ADDRESS` | `:8080` | Server listen address |
| `DATABASE_PATH` | `./data/products.db` | SQLite database file |
