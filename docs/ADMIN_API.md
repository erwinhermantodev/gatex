# Admin API

Manage services, routes and gRPC mappings at runtime. The [dashboard](DASHBOARD.md) is a UI over this API (it doesn't yet expose route middleware — use this API for that). Changes to services and routes take effect immediately — the gateway reloads its route table after every successful write.

## Authentication

Every `/admin/*` request needs the token from `ADMIN_API_TOKEN`:

```
Authorization: Bearer <token>        (or)        X-Admin-Token: <token>
```

| Response | Meaning                                  |
| :------- | :--------------------------------------- |
| `401`    | Missing or wrong token                   |
| `503`    | `ADMIN_API_TOKEN` is not set on the server |

Admin endpoints are not rate limited and are excluded from the request log.

## Conventions

- JSON field names are **PascalCase** (`BaseURL`, `ServiceID`) because the models have no JSON tags. Records also carry `ID`, `CreatedAt`, `UpdatedAt`, `DeletedAt`.
- Updates (`PUT`) replace the editable fields; `ID` and `CreatedAt` can't be changed.
- Deletes are soft deletes (the row is hidden, not removed). Note that a deleted service or route still holds its unique `Name`/`Path`, so you can't re-create one with the same value.
- Validation failures return `400` with a message. Unknown IDs on update return `404`.

## Services

An upstream the gateway forwards to.

| Method & path               | Description        |
| :-------------------------- | :----------------- |
| `GET /admin/services`       | List               |
| `POST /admin/services`      | Create (`201`)     |
| `PUT /admin/services/:id`   | Update             |
| `DELETE /admin/services/:id`| Delete (`204`)     |

| Field      | Notes                                                                                                   |
| :--------- | :------------------------------------------------------------------------------------------------------ |
| `Name`     | Required, unique                                                                                        |
| `Protocol` | `rest` (default) or `grpc`                                                                              |
| `BaseURL`  | `rest`: absolute `http(s)` URL, no credentials. Trailing `/` is trimmed.                                |
| `GRPCAddr` | `grpc`: `host:port`                                                                                     |
| `Status`, `LastCheck` | Read-only in practice — maintained by the health checker                                     |

Hosts are rejected if they are cloud-metadata names/IPs (`169.254.169.254`, `metadata.google.internal`), link-local, unspecified, multicast, or loopback (unless `ALLOW_LOOPBACK_UPSTREAMS=true`). Hostnames are resolved and every address is checked.

```bash
curl -X POST localhost:8080/admin/services -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"Name":"orders","Protocol":"grpc","GRPCAddr":"orders.internal:50051"}'
```

## Routes

Maps a method + path on the gateway to a service.

| Method & path             | Description    |
| :------------------------ | :------------- |
| `GET /admin/routes`       | List (includes `Service` and `ProtoMapping`) |
| `POST /admin/routes`      | Create (`201`) |
| `PUT /admin/routes/:id`   | Update         |
| `DELETE /admin/routes/:id`| Delete (`204`) |

| Field            | Notes                                                                                                         |
| :--------------- | :------------------------------------------------------------------------------------------------------------ |
| `Method`         | `GET` `POST` `PUT` `PATCH` `DELETE` `HEAD`                                                                    |
| `Path`           | Starts with `/`; unique. Supports `:name` parameters and a trailing `*` (e.g. `/files/*`). `/admin` and `/dashboard` are reserved. |
| `ServiceID`      | Must exist                                                                                                    |
| `ProtoMappingID` | Optional, gRPC services only: which RPC this route calls. Must belong to the same service. If omitted, the service's first mapping is used. |
| `EndpointFilter` | Handler identifier. A name matching a built-in handler (see [Architecture](ARCHITECTURE.md#built-in-handlers)) runs that handler; anything else uses the generic proxy. |
| `Tag`            | Label shown in metrics                                                                                        |
| `Middleware`     | **A JSON array encoded as a string**, e.g. `"[\"jwt\",\"timeout\"]"`. Applied in the listed order, first = outermost. |

Matching: more specific routes win (more fixed segments first). A path that exists for another method returns `405`; no match returns `404`.

### Route middleware

| Name              | Effect                                                                                                      |
| :---------------- | :---------------------------------------------------------------------------------------------------------- |
| `jwt`             | Requires `Authorization: Bearer <JWT>`. HS256 (`JWT_SECRET`) and/or RS256 (`JWT_PUBLIC_KEY_PEM`). `exp` is required; `nbf`, `iss`, `aud` checked when configured. Claims are available to later handlers in the request context. `401` if invalid, `503` if not configured. |
| `api-key`         | Requires `x-api-key` matching one of `API_KEYS`. `401` if invalid, `503` if not configured.                 |
| `timeout`         | 10 s overall deadline, then `504`.                                                                          |
| `retry`           | Retries up to 3 times on 5xx. Only use on idempotent routes.                                                |
| `circuit-breaker` | Opens after 5 consecutive 5xx and answers `503` for 30 s. (Separate from the always-on per-service breaker.) |

Put authentication first so unauthenticated requests never reach `retry`/upstream.

```bash
curl -X POST localhost:8080/admin/routes -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"Path":"/api/orders/:id","Method":"GET","ServiceID":2,"EndpointFilter":"order-get","Tag":"orders","Middleware":"[\"jwt\",\"timeout\"]"}'
```

## Proto mappings

Describe one gRPC method of a `grpc` service. The gateway discovers message types through server reflection, so no `.proto` files are needed — the upstream must have reflection enabled.

| Method & path                     | Description    |
| :-------------------------------- | :------------- |
| `GET /admin/proto-mappings`       | List           |
| `POST /admin/proto-mappings`      | Create (`201`) |
| `PUT /admin/proto-mappings/:id`   | Update         |
| `DELETE /admin/proto-mappings/:id`| Delete (`204`) |

| Field          | Example        | Notes                          |
| :------------- | :------------- | :----------------------------- |
| `ServiceID`    | `2`            | Must be a `grpc` service       |
| `ProtoPackage` | `orders.v1`    | Required                       |
| `ServiceName`  | `OrderService` | Required (without the package) |
| `RPCMethod`    | `GetOrder`     | Required                       |

The full method called is `/<ProtoPackage>.<ServiceName>/<RPCMethod>`. Create one mapping per RPC, then point each route at its mapping with `ProtoMappingID`. Method descriptors are cached for 5 minutes, so a changed upstream proto can take that long to be picked up.

## Monitoring endpoints

| Endpoint                   | Description                                                            |
| :------------------------- | :--------------------------------------------------------------------- |
| `GET /admin/metrics`       | Per-service traffic, health score and circuit state (`CLOSED`/`OPEN`/`HALF-OPEN`) |
| `GET /admin/request-logs`  | Latest 100 requests (dashboard/admin traffic excluded)                 |
| `GET /admin/traces/:id`    | Trace events for one `X-Request-Id`                                    |
| `GET /admin/logs`          | Latest 50 admin actions (audit log)                                    |
| `GET /admin/server-logs`   | Recent gateway console output                                          |

Request and trace logs are kept for `LOG_RETENTION_DAYS` (default 7).
