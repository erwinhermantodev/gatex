# Dashboard Guide

The dashboard is a web UI for operating the gateway: configure services, routes and gRPC mappings, watch traffic, and debug individual requests. It talks to the [Admin API](ADMIN_API.md), so anything you can do here you can also script.

## Opening the dashboard

| Setup            | URL                                  | Notes                                                                                   |
| :--------------- | :----------------------------------- | :-------------------------------------------------------------------------------------- |
| Served by gateway | `http://<gateway-host>:8080/dashboard` | Needs a build first: `cd dashboard && npm install && npm run build`. Run the gateway from the repo root (it serves `dashboard/dist` by relative path). |
| Development       | `http://localhost:5173/dashboard`    | `npm run dev`; API calls are proxied to the gateway (see `dashboard/vite.config.ts`).   |

**Signing in.** The page itself is public, but all its data comes from the token-protected Admin API. The first time a request is rejected with `401`, the browser asks for the *Admin API token* (the server's `ADMIN_API_TOKEN`). It is stored in the browser's localStorage under `gateway_admin_token` and sent as a Bearer token on every call.

- Wrong token: you are prompted again on the next `401`.
- To sign out or change the token, clear that key (browser dev tools → Application → Local Storage).
- If the server has no `ADMIN_API_TOKEN` set, the Admin API answers `503` and the dashboard shows no data.

## Layout

A sidebar switches between sections; the main area shows the selected section. Lists are paged client-side, can be searched/filtered, and open **create/edit forms in a modal**. Deleting always asks for confirmation.

| Section          | Purpose                                                  |
| :--------------- | :------------------------------------------------------- |
| Dashboard        | Overview: health, volume, latency, recent admin activity |
| Services         | Upstream services (REST or gRPC)                         |
| Routes           | Gateway paths and where they go                          |
| Proto Mappings   | gRPC methods that routes can call                        |
| Traffic Monitor  | Recent requests, with a per-request trace                |
| System Logs      | Live gateway console output                              |

## Dashboard (overview)

- **Stat cards:** active services (online / total, from the 30-second health checks), total requests and average latency (from live metrics).
- **Routes table** with their upstream service and status.
- **Activity log:** the latest admin changes (who/what/when). Searchable, filterable by action, 5 per page.
- Metrics refresh automatically every 20 seconds; use the refresh button for an immediate update.

> The "Gateway Health 100%" card is a static label, not a computed value. Use the circuit-breaker badges and service status instead.

## Services

One card per upstream service (6 per page). Search by name and filter by protocol (REST / gRPC).

Each card shows:
- name, protocol and target (`BaseURL` or gRPC address),
- **CB badge**: the service's circuit-breaker state — `CLOSED` (healthy), `HALF-OPEN` (probing after an outage), `OPEN` (requests are rejected with 503 for 30 s),
- online/offline status from the health checker.

**Add / edit a service**

| Field         | Notes                                                                                           |
| :------------ | :---------------------------------------------------------------------------------------------- |
| Service Name  | Unique                                                                                          |
| Protocol      | REST or gRPC; the form then asks for **Base URL** (`https://users.internal:9000`) or **gRPC Address** (`orders.internal:50051`) |

Saving fails with a message in the browser console if the target is not allowed (cloud-metadata, link-local, or loopback addresses — see the [security notes](../README.md#security)). The form closes only on success.

## Routes

A sortable, filterable table (8 per page; filter by HTTP method, search by path; click a column header to sort).

**Add / edit a route**

| Field                    | Notes                                                                                                   |
| :----------------------- | :------------------------------------------------------------------------------------------------------ |
| Method                   | GET, POST, PUT, DELETE, PATCH                                                                           |
| Path                     | e.g. `/v1/auth/login`; supports `:id` parameters and a trailing `*`. `/admin` and `/dashboard` are reserved. |
| Upstream Service         | Pick one of your services                                                                                |
| gRPC Method Mapping      | **Shown only for gRPC services.** Choose the RPC this route calls. "Auto" uses the service's first mapping — fine for single-method services, ambiguous otherwise. Create mappings first in *Proto Mappings*. |
| Endpoint Filter / Handler| Identifier for the route. A built-in handler name (e.g. `login`) runs that handler; anything else uses the generic proxy. |

Changes take effect immediately — no gateway restart.

**What the form can't do (yet).** There is no field for route **middleware** (`jwt`, `api-key`, `timeout`, `retry`, `circuit-breaker`) and no `HEAD` method. Set those through the Admin API ([details](ADMIN_API.md#route-middleware)); the dashboard preserves a route's existing middleware when you edit it here.

## Proto Mappings

Sortable table (search by method, filter by proto package) with the *full signature* of each mapping (`package.Service/Method`).

| Field            | Example        | Notes                                                              |
| :--------------- | :------------- | :----------------------------------------------------------------- |
| Service          | `orders`       | Only gRPC services are listed                                      |
| RPC Method Name  | `GetOrder`     |                                                                    |
| Proto Package    | `orders.v1`    |                                                                    |
| gRPC Service Name| `OrderService` | Without the package                                                |
| Request / Response Type | `GetOrderRequest` / `GetOrderResponse` | Informational; the gateway resolves real types via server reflection |

The upstream must expose **gRPC server reflection**. After creating a mapping, point a route at it (Routes → *gRPC Method Mapping*).

**Typical gRPC setup:** add the gRPC service → add one mapping per RPC → add one route per RPC choosing its mapping → call the route with a JSON body.

## Traffic Monitor

The most recent 100 gateway requests (newest first; dashboard and admin traffic are not logged): time, method, path, status, latency, client IP and user agent. 10 per page.

- Filter by status class (2xx / 3xx / 4xx / 5xx) and method; search by path.
- **Click a row to open its trace timeline** — the events the gateway recorded for that request (interpreting, proxying/dialling, invoking, errors), each tagged INFO / WARN / ERROR with a time. Use it to see which service was hit and where it failed.
- The trace is looked up by the request's `X-Request-Id`. If a user reports a problem, ask for that header (see the [User Guide](../USER_GUIDE.md#troubleshooting)).
- Logs are kept for `LOG_RETENTION_DAYS` (default 7). Older requests show "No trace events recorded".

## System Logs

A live, auto-scrolling console of the gateway's own output (startup, route reloads, errors, access log lines). It refreshes every 5 seconds and holds only the most recent entries in memory — it is cleared when the gateway restarts.

## Known gaps

Parts of the UI are placeholders and do nothing yet:

- **Settings** appears in the sidebar but has no page.
- The header **search box**, the **notification bell** and the "Gateway Operational" badge are decorative; the sidebar brand and the "Admin User" footer are placeholder text.
- Lists load once per visit and on the refresh button; only metrics (20 s) and System Logs (5 s) poll automatically.
- Save and delete failures are logged to the browser console rather than shown on screen — if a form doesn't close, open dev tools to read the error.

## Troubleshooting

| Symptom                                   | Likely cause / fix                                                                              |
| :---------------------------------------- | :---------------------------------------------------------------------------------------------- |
| Blank page at `/dashboard`                | `dashboard/dist` missing — run `npm run build`; start the gateway from the repo root            |
| Token prompt keeps coming back            | Wrong token, or it differs from the server's `ADMIN_API_TOKEN`                                  |
| Every list is empty, no prompt            | `ADMIN_API_TOKEN` not set on the server (503) — set it and restart                              |
| Service/route form won't save             | Validation rejected it (reserved path, blocked address, unknown service) — check the console     |
| New route returns 404                     | Check method and path match exactly; a path that exists for another method returns 405           |
| No trace events for a request             | Older than the retention window, or the request was served before logging caught up (batched, ~1 s) |
| Service shows `OPEN` circuit              | The upstream failed 5 times in a row; it retries automatically after 30 s                        |
