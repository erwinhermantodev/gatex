# Gateway User Guide

This guide provides information on how to consume APIs through the Gateway Service.

## 🛰️ Base URL

All API requests should be sent to:
`http://<gateway-host>:8080`

## 📑 Request Headers

The gateway supports and propagates several important headers:

| Header          | Description                                                                                                          | Required           |
| :-------------- | :------------------------------------------------------------------------------------------------------------------- | :----------------- |
| `Content-Type`  | Set to `application/json` for most requests.                                                                         | Yes                |
| `X-Request-Id`  | A unique identifier for the request. If not provided, the gateway generates one. Use this for tracing and debugging. | Optional           |
| `Authorization` | Bearer token or API key as required by the upstream service.                                                         | Depends on service |

## 🛠️ Consuming APIs

The gateway routes requests based on the **Path** and **HTTP Method**.

### REST Endpoints

Simply call the path defined in the Gateway configuration. The gateway will proxy the request to the configured upstream service.

**Example:**

```bash
curl -X GET http://localhost:8080/v1/users/profile \
     -H "X-Request-Id: client-req-123" \
     -H "Content-Type: application/json"
```

### gRPC Endpoints (via Transcoding)

The gateway allows you to call gRPC services using standard REST/JSON semantics.

- Send a `POST` or `GET` request to the mapped path.
- The gateway converts your JSON body into a Protobuf message and invokes the gRPC service.
- The response is converted back to JSON.

**Example:**

```bash
curl -X POST http://localhost:8080/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username": "admin", "password": "password"}'
```

## 🔍 Observability & Support

If you encounter an issue with a request, please provide the `X-Request-Id` returned in the response headers.

### Why the `X-Request-Id` matters:

Every request is tracked using **Distributed Tracing**. Providing this ID allows the Gateway administrators to view a complete timeline of your request, including:

- Exactly when the gateway received the request.
- Which upstream service was targeted.
- The latency of the upstream call.
- Any internal errors or warnings that occurred during proxying.

## ❌ Error Handling

The gateway returns standard HTTP status codes:

- `2xx`: Success
- `4xx`: Client Error (e.g., `404 Not Found`, `429 Too Many Requests`)
- `5xx`: Server Error (e.g., upstream service is down)

Errors are returned in a consistent JSON format:

```json
{
  "status": false,
  "code": "ERROR_CODE",
  "message": "A human-readable error message",
  "data": null
}
```
