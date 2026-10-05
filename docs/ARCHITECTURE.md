# Architecture

## Request flow

```
client
  │
  ▼
RequestID ─► CacheControl ─► Metrics ─► TrafficLogger ─► RateLimiter ─► Recover ─► Gzip ─► Logger ─► CORS
  │
  ├─ /admin/*      ─► AdminAuth ─► admin handlers (CRUD, metrics, logs)
  ├─ /dashboard/*  ─► static React build
  └─ everything else ─► Registry.Handle (catch-all)
                          │  match method + path against the in-memory route table
                          ▼
                     route middleware chain (jwt / api-key / timeout / retry / circuit-breaker)
                          ▼
                     DynamicHandler
                          ├─ EndpointFilter is a built-in handler ─► that handler
                          └─ otherwise GenericProxyHandler
                                ├─ circuit breaker check (per service)
                                ├─ REST: reverse proxy to BaseURL
                                └─ gRPC: JSON ─► protobuf ─► Invoke ─► JSON
```

`RequestID` runs first so every later component (traffic log, trace events, gRPC metadata) sees the same ID.

## Components

**Route registry** (`route/registry.go`). Routes are loaded from the database (with their service and proto mapping) into an immutable table that is swapped atomically. The admin API calls `Reload()` after each service/route write, so changes apply without a restart. If a reload fails the previous table stays active. Matching is done by the registry rather than Echo's router, because Echo can't remove routes at runtime: patterns support `:param` and a trailing `*`; more fixed segments rank first. Each route's middleware chain is built once at reload time.

**Generic proxy** (`route/proxy.go`).
- *REST*: `httputil.ReverseProxy` over one shared transport (connection reuse, no environment proxy) whose dialer refuses blocked addresses.
- *gRPC*: one shared `ClientConn` per address (`route/grpc_pool.go`), method descriptors resolved via reflection and cached for 5 minutes. Request headers become gRPC metadata (hop-by-hop headers dropped), plus `x-request-id` and `x-forwarded-for`. Any gRPC failure is returned as `502`.

**Circuit breaker** (`util/health_registry.go`). One per service, always on. Closed → Open after 5 consecutive failures (REST 5xx / proxy errors / gRPC errors). After 30 s it becomes Half-open and admits a single probe; success closes it, failure reopens it. Open circuits answer `503` immediately.

**Route middleware** (`route/middleware/`). `jwt` and `api-key` are stdlib-only implementations. The JWT algorithm is chosen by which key is configured, never taken from the token header. `retry`/`timeout`/`circuit-breaker` are generic resilience wrappers.

**SSRF guard** (`util/netguard`). Validates hosts when services are saved, and again on every connection (on the resolved IP) for both REST and gRPC.

**Logging** (`database/log_writer.go`). Request logs and trace events go through bounded queues and are inserted in batches (100 rows or 1 s). A full queue drops entries instead of blocking requests. An hourly job hard-deletes rows older than `LOG_RETENTION_DAYS`.

**Health checker** (`cron/health.go`). Probes every service each 30 s and stores `Status`/`LastCheck`.

## Data model

| Table           | Purpose                                                                                |
| :-------------- | :------------------------------------------------------------------------------------- |
| `services`      | Upstream: `Name`, `Protocol`, `BaseURL`/`GRPCAddr`, health status                      |
| `routes`        | Method + `Path` → `ServiceID`, optional `ProtoMappingID`, `EndpointFilter`, `Tag`, `Middleware` |
| `proto_mappings`| gRPC package/service/method for a service                                              |
| `activity_logs` | Admin audit trail                                                                      |
| `request_logs`  | One row per gateway request                                                            |
| `trace_logs`    | Per-request timeline events                                                            |

All use `gorm.Model` (soft delete). Schema is migrated on start.

## Built-in handlers

`route/endpoint.go` registers hand-written handlers for the auth-service (`login`, `check-phone`, `refresh-token`, `logout`, `activation-*`, `otp-*`, `register-*`, `profile`, and `*-grpc` variants), which talk to `AUTH_SERVICE_BASE_URL` / `AUTH_SERVICE_GRPC_ADDR`. A route whose `EndpointFilter` equals one of these names runs that handler instead of the generic proxy. These predate the generic proxy; new integrations should use plain routes.

## Error responses

Errors are rendered by `util.CustomHTTPErrorHandler` with the real HTTP status:

| HTTP | `code` | Meaning                                      |
| :--- | :----- | :------------------------------------------- |
| 400  | `005`  | Bad request / validation failed              |
| 401  | `006`  | Unauthenticated                              |
| 403  | `007`  | Forbidden                                    |
| 404  | `008`  | No route / resource                          |
| 405  | `009`  | Path exists for other methods                |
| 408  | `011`  | Request timeout                              |
| 409  | `012`  | Conflict                                     |
| 413  | `013`  | Body too large                               |
| 414  | `014`  | URI too long                                 |
| 415  | `015`  | Unsupported media type                       |
| 429  | `016`  | Too many requests (rate limit)               |
| 431  | `017`  | Headers too large                            |
| 502/503/504 | `018` | Upstream failure, open circuit, timeout |
| 500  | `999`  | Anything else (details are not exposed)      |

Body shape: `{"status": false, "code": "...", "message": "...", "data": null}`. The rate limiter answers `429` with a plain-text body instead.

## Known limitations

- `Route.Path` and `Service.Name` are unique even among soft-deleted rows.
- Upstream proto changes can take up to 5 minutes to be seen (descriptor cache).
- gRPC upstreams are dialled without TLS.
- Rate-limit values (10 rps, burst 5) and the retry/timeout/breaker parameters are fixed in code, not per route.
