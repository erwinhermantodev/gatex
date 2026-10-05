# Gateway Service

An API gateway written in Go (Echo, GORM/Postgres) with a React admin dashboard. Routes, upstream services and gRPC mappings live in the database and are managed at runtime — no redeploy or restart to add, change or remove a route.

- **REST proxy** to any HTTP upstream.
- **REST-to-gRPC transcoding** using gRPC server reflection (JSON in, JSON out).
- **Per-route middleware**: JWT / API-key auth, timeout, retry, circuit breaker.
- **Observability**: request log, per-request trace timeline, live server console, service health and circuit state.

## Documentation

| Doc                                      | For                                                      |
| :--------------------------------------- | :------------------------------------------------------- |
| [USER_GUIDE.md](USER_GUIDE.md)           | Developers calling APIs through the gateway              |
| [docs/DASHBOARD.md](docs/DASHBOARD.md)   | Operators using the web UI                               |
| [docs/ADMIN_API.md](docs/ADMIN_API.md)   | Scripting services, routes and mappings                  |
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | Contributors: request flow, components, design choices |
| This README                              | Setup, configuration, development                        |

## Quick start

Requirements: Go 1.24+, PostgreSQL, Node.js 18+ (only to build the dashboard).

```bash
cp .env.example .env        # then fill in DB_* and set ADMIN_API_TOKEN
go run main.go              # listens on :8080 (APP_PORT)
```

Tables are created automatically on start (GORM AutoMigrate).

**Dashboard** (served by the gateway at `/dashboard` from `dashboard/dist`):

```bash
cd dashboard
npm install
npm run build               # gateway serves dashboard/dist; run the gateway from the repo root
# or, for development with hot reload:
npm run dev                 # http://localhost:5173/dashboard
```

The dashboard asks for the admin token the first time the Admin API answers 401, and keeps it in the browser's localStorage. See the [Dashboard Guide](docs/DASHBOARD.md).

**Your first route** (replace the token and upstream):

```bash
TOKEN=change-me
curl -X POST localhost:8080/admin/services -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"Name":"users","Protocol":"rest","BaseURL":"http://users.internal:9000"}'

curl -X POST localhost:8080/admin/routes -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"Path":"/api/users/:id","Method":"GET","ServiceID":1,"EndpointFilter":"user-get","Tag":"users"}'

curl localhost:8080/api/users/42      # proxied to http://users.internal:9000/api/users/42
```

## Configuration

Set as environment variables or in `.env` (loaded at start). See [.env.example](.env.example).

| Variable                   | Default | Description                                                                                                  |
| :------------------------- | :------ | :----------------------------------------------------------------------------------------------------------- |
| `APP_PORT`                 | `8080`  | HTTP listen port                                                                                             |
| `DB_HOST` `DB_PORT` `DB_USER` `DB_PASSWORD` `DB_NAME` | `localhost` / `5432` / – | PostgreSQL connection                                                          |
| `ADMIN_API_TOKEN`          | –       | **Required.** Bearer token for `/admin/*`. If unset, the Admin API returns 503 (fails closed).               |
| `CORS_ALLOW_ORIGINS`       | `*`     | Comma-separated allowed origins                                                                              |
| `LOG_RETENTION_DAYS`       | `7`     | Request/trace logs older than this are deleted hourly. `0` keeps everything.                                 |
| `TRUST_PROXY_HEADERS`      | `true`  | Use `X-Forwarded-For` for the client IP (rate limiting, logs). Set `false` if clients connect directly.      |
| `ALLOW_LOOPBACK_UPSTREAMS` | `false` | Allow upstreams on `localhost`/`127.0.0.1`. Development only — see [Security](#security).                    |
| `JWT_SECRET`               | –       | Enables HS256 for the `jwt` route middleware                                                                 |
| `JWT_PUBLIC_KEY_PEM`       | –       | Enables RS256 for the `jwt` route middleware (PEM public key or certificate; `\n` escapes are accepted)      |
| `JWT_ISSUER` `JWT_AUDIENCE`| –       | If set, tokens must match `iss` / `aud`                                                                      |
| `API_KEYS`                 | –       | Comma-separated keys accepted by the `api-key` route middleware (`x-api-key` header)                         |
| `AUTH_SERVICE_BASE_URL` `AUTH_SERVICE_GRPC_ADDR` | – | Upstream for the built-in auth handlers (see [Architecture](docs/ARCHITECTURE.md#built-in-handlers)) |
| `DEFAULT_LANG`             | `id`    | Default language for the built-in auth handlers                                                              |

## Security

- **Admin API** requires `Authorization: Bearer <ADMIN_API_TOKEN>` (or `X-Admin-Token`). Treat the token like a root credential: anyone holding it can point the gateway at any allowed upstream.
- **Upstream restrictions (SSRF).** Services may not target cloud-metadata hosts, link-local, unspecified or multicast addresses, or (by default) loopback. Private ranges such as `10.0.0.0/8` are allowed, since that is where internal services live. The check runs when a service is saved *and* again at connect time, so DNS rebinding can't bypass it.
- **Route auth fails closed.** A route using `jwt` or `api-key` returns 503 if that mechanism isn't configured, and a route listing an unknown middleware name is disabled (500) rather than served unprotected.
- **Secrets stay out of git.** `.env`, `*.db` and the built binary are git-ignored. Never commit real credentials; if credentials were ever committed, rotate them.
- The dashboard page (`/dashboard`) is public; its data comes from the token-protected Admin API.

## Operations notes

- **Rate limit:** 10 requests/second per client IP, burst 5, on gateway routes (not `/admin` or `/dashboard`). Idle limiters are evicted after 10 minutes.
- **Health checks:** every 30 s each service is probed (REST: `GET <BaseURL>/health`, falling back to `GET <BaseURL>`; gRPC: TCP connect). Results show in the dashboard.
- **Circuit breaker:** 5 consecutive failures open a service's circuit for 30 s (503 to clients). Then one probe request is let through: success closes the circuit, failure reopens it.
- **Logs:** request and trace logs are written asynchronously in batches. If the database can't keep up, entries are dropped rather than slowing requests down.
- **Upgrading:** new columns (such as `routes.proto_mapping_id`) are added automatically on start.

## Development

```bash
go build ./...
go vet ./...
go test ./...            # unit + integration tests (in-memory SQLite, no Postgres needed)
go test -race ./route/
```

The integration tests (`route/integration_test.go`) run the whole gateway in-process against `httptest` servers and a real gRPC server with reflection.

`cmd/migrate` seeds the database from `route/gate/auth.json` (legacy bootstrap); routes are normally managed through the Admin API or dashboard.

## Project layout

```
main.go                  entry point
config/                  environment configuration
database/                GORM models, connection, async log writer + retention
route/                   router setup, route registry, proxy, gRPC pool, middleware wiring
route/middleware/        admin auth, route auth (jwt/api-key), resilience, metrics, traffic log
domain/admin/            Admin API handlers and input validation
domain/auth/             built-in auth-service handlers (REST + gRPC)
cron/                    service health checker
util/                    helpers, circuit breaker state, metrics, tracing, SSRF guard (netguard)
dashboard/               React + Vite admin UI
```
