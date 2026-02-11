# Gateway Service

A high-performance, feature-rich API Gateway designed for the ecosystem. Built with Go and React, it provides seamless service orchestration, gRPC transcoding, and deep observability through a premium monitoring dashboard.

## 📖 Documentation

- **[User Guide (Customer API Docs)](file:///Users/Project/agen-pos/gateway-service/USER_GUIDE.md)**: How to consume APIs through the gateway.
- **Admin API**: See the [API Documentation](#-api-documentation) section below for managing the gateway.

## 🚀 Key Features

### 🔄 Dynamic gRPC Transcoding

- **REST-to-gRPC Bridge**: Automatically transcode JSON/REST requests into gRPC calls based on configurable proto mappings.
- **Reflection Support**: Uses gRPC reflection to dynamically resolve service methods and message types.
- **Service Discovery**: Seamlessly routes traffic to upstream gRPC services with built-in connection management.

### 🕵️ Distributed Tracing

- **Request Timelines**: Granular tracking of request lifecycles (Interpreting -> Proxying -> Invocation -> Response).
- **RequestID Propagation**: End-to-end tracing linked by unique request identifiers.
- **Visual Trace View**: Interactive event timeline in the dashboard for deep-dive debugging.

### 📊 Premium Monitoring Dashboard

- **Traffic Monitor**: Real-time request logging with advanced filtering (Path, Method, Status, RequestID).
- **System Logs**: Centralized "System Console" for real-time server output monitoring.
- **Live Metrics**: At-a-glance view of system health, CPU/Memory usage, and traffic throughput.
- **Configuration Management**: Full CRUD interface for managing Services, Routes, and Proto Mappings.

### 🛡️ Robust Security & Traffic Control

- **Rate Limiting**: Intelligent IP-based throttling to protect upstream services.
- **Administrative Exemptions**: Enhanced logic to ensure management dashboard remains responsive under load.
- **CORS & Middleware**: Pre-configured security headers and a flexible middleware chain.

## 🛠️ Tech Stack

- **Backend**: [Go](https://go.dev/), [Echo Framework](https://echo.labstack.com/), [GORM](https://gorm.io/) (SQLite/PostgreSQL)
- **Frontend**: [React](https://reactjs.org/), [Vite](https://vitejs.dev/), [Tailwind CSS](https://tailwindcss.com/), [Framer Motion](https://www.framer.com/motion/)
- **Protocol**: gRPC, Protobuf, gRPC-Reflection

## 📦 Getting Started

### Prerequisites

- Go 1.20+
- Node.js 18+ & npm

### Backend Setup

1. Install dependencies:
   ```bash
   go mod download
   ```
2. Run the server:
   ```bash
   go run main.go
   ```
   _The gateway will start on port `8080` by default._

### Dashboard Setup

1. Navigate to the dashboard directory:
   ```bash
   cd dashboard
   ```
2. Install dependencies:
   ```bash
   npm install
   ```
3. Run in development mode:
   ```bash
   npm run dev
   ```
   _The dashboard will be available at `http://localhost:5173/dashboard`._

## 📖 API Documentation

The gateway exposes an administrative API for configuration and monitoring:

| Endpoint                | Method   | Description                             |
| :---------------------- | :------- | :-------------------------------------- |
| `/admin/services`       | GET/POST | Manage upstream services                |
| `/admin/routes`         | GET/POST | Manage routing rules                    |
| `/admin/proto-mappings` | GET/POST | Manage REST-to-gRPC mappings            |
| `/admin/metrics`        | GET      | System health and traffic stats         |
| `/admin/request-logs`   | GET      | Traffic history                         |
| `/admin/traces/:id`     | GET      | Detailed trace for a specific RequestID |
| `/admin/server-logs`    | GET      | Real-time server console output         |

---

Developed with ❤️ for the Agen-POS Team.
