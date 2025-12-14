# URL Shortener (Go + MariaDB + docker-compose)

## Endpoints

### Shorten
- `POST /api/v1/data/shorten`
- Body:
```json
{
  "originalUrl": "https://example.com/path",
  "expireAt": "2030-01-01T00:00:00Z"
}
```
`expireAt` is optional. If omitted or empty => permanent.

### Redirect (302 with Location)
- `GET /api/v1/shortUrl/{code}`
- `GET /s/{code}` (this is what `shortUrl` in the response uses)

### Soft delete
- `DELETE /api/v1/data/{code}`

### Not found
All "not found / expired / deleted" cases return `404` with JSON:
```json
{"error":"not_found","message":"not found"}
```

## Run
```bash
cp .env.example .env
docker compose up --build
```

## Test
```bash
curl -sS -X POST http://localhost:8080/api/v1/data/shorten \
  -H 'Content-Type: application/json' \
  -d '{"originalUrl":"https://example.com/hello","expireAt":"2030-01-01T00:00:00Z"}'

curl -v http://localhost:8080/s/<code>
curl -v http://localhost:8080/api/v1/shortUrl/<code>

curl -i -X DELETE http://localhost:8080/api/v1/data/<code>
```

## ID generator
`internal/idgen` is a placeholder. Replace it with your real 64-bit unique ID generator.
