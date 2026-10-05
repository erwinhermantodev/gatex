# Gateway User Guide

How to call APIs through the Gateway Service. For operating the gateway, see [docs/ADMIN_API.md](docs/ADMIN_API.md).

## Base URL

`http://<gateway-host>:8080`

Ask your gateway administrator which paths are available; they are configured per environment.

## Request headers

| Header          | Description                                                                                                           | Required                  |
| :-------------- | :-------------------------------------------------------------------------------------------------------------------- | :------------------------ |
| `Content-Type`  | `application/json` for requests with a body.                                                                          | When sending a body       |
| `X-Request-Id`  | Your own correlation ID. If you don't send one, the gateway generates one. It is returned in the response headers.    | Optional                  |
| `Authorization` | `Bearer <token>` for routes protected by JWT auth. Also forwarded to the upstream service.                            | Depends on route          |
| `x-api-key`     | API key for routes protected by API-key auth.                                                                         | Depends on route          |

Other request headers are forwarded to the upstream service as well (for gRPC upstreams they become gRPC metadata).

## Calling REST endpoints

Call the gateway path exactly as configured; method, path and body are proxied to the upstream, and the upstream's response is returned unchanged.

```bash
curl http://localhost:8080/api/users/42 \
     -H "Authorization: Bearer $TOKEN" \
     -H "X-Request-Id: client-req-123"
```

Paths may contain parameters (`/api/users/:id` matches `/api/users/42`). A trailing slash is ignored.

## Calling gRPC services (JSON transcoding)

gRPC methods are exposed as ordinary JSON endpoints. Send the request message as JSON; field names follow the protobuf message (e.g. `phone_number`).

```bash
curl -X POST http://localhost:8080/auth/check-phone \
     -H "Content-Type: application/json" \
     -d '{"phone_number": "0812345678"}'
```

- The JSON body becomes the protobuf request; the protobuf response is returned as JSON with HTTP `200`.
- Authentication headers are passed to the gRPC service as metadata.
- If the gRPC call fails, you get `502` with the reason in `message`.

## Authentication

Some routes require authentication at the gateway itself, before the request reaches the service:

| Route protection | What to send                                                                  | Failure              |
| :--------------- | :---------------------------------------------------------------------------- | :------------------- |
| JWT              | `Authorization: Bearer <JWT>` (the token must not be expired)                 | `401` invalid/missing |
| API key          | `x-api-key: <key>`                                                            | `401` invalid/missing |

`503` on such a route means authentication isn't configured on the gateway — contact the administrator.

## Rate limiting

Each client IP may send **10 requests per second** (short bursts of 5). Beyond that the gateway answers `429 Too Many Requests` (plain-text body). Back off and retry with a short delay.

## Errors

Errors use the real HTTP status and a JSON body:

```json
{
  "status": false,
  "code": "008",
  "message": "Not Found",
  "data": null
}
```

| HTTP | `code` | Typical cause                                                         |
| :--- | :----- | :-------------------------------------------------------------------- |
| 400  | `005`  | Invalid request body or parameters                                    |
| 401  | `006`  | Missing/invalid token or API key                                      |
| 403  | `007`  | Not allowed                                                           |
| 404  | `008`  | No route for this path                                                |
| 405  | `009`  | The path exists, but not for this HTTP method                         |
| 429  | `016`  | Rate limit exceeded                                                   |
| 502  | `018`  | The upstream service failed or returned an invalid response           |
| 503  | `018`  | The upstream is temporarily disabled by the circuit breaker, or auth isn't configured |
| 504  | `018`  | The upstream took too long                                            |
| 500  | `999`  | Unexpected gateway error                                              |

Errors returned by the upstream service itself (REST) are passed through as the upstream sent them, so their format may differ.

**Circuit breaker:** after repeated upstream failures the gateway answers `503` immediately for about 30 seconds, then lets a single request through to check whether the service has recovered. Retrying after a short pause is safe.

## Troubleshooting

Every response carries an `X-Request-Id` header. When reporting a problem, include that ID: administrators can open the full timeline of the request — when it arrived, which service handled it, how long the upstream took, and any errors — in the dashboard.
