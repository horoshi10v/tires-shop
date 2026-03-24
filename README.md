# Tires Shop Backend

A production-style Go backend for a Telegram-first tire and wheel commerce platform.

This service powers three operational surfaces:
- a buyer-facing Telegram Mini App for catalog browsing and checkout,
- a staff-facing warehouse and order management panel,
- an admin layer for reporting, exports, user management, and operational auditability.

The codebase is structured around clear application boundaries: HTTP transport, services, repositories, infrastructure integrations, and domain models. It is intentionally practical rather than over-engineered, with a focus on shipping warehouse workflows, Telegram automation, and e-commerce operations end to end.

## Highlights

- Telegram authentication for Mini Apps
- Public catalog API for client storefronts
- Order creation and order history for buyers
- Staff inventory management with lots, photos, QR codes, and printable price tags
- Warehouse CRUD and inter-warehouse transfers
- Order status workflow and buyer messaging via Telegram bot
- Admin-only reports, exports, notifications, audit logs, and user management
- MinIO-backed media storage
- Google Sheets export support
- PostgreSQL persistence with automatic schema migration on startup

## Product Scope

The backend is designed for a tire retail operation, but the data model is broader than tires only.

Supported product groups include:
- Tires
- Rims
- Accessories

Accessory support includes categories such as:
- Fasteners
- Hub rings
- Spacers
- Tire bags

This makes the service suitable for a compact but realistic automotive commerce and warehouse management project.

## Architecture

The project follows a layered structure:

- `internal/domain` contains business entities, DTOs, filters, and repository/service contracts.
- `internal/service` contains business workflows and orchestration.
- `internal/repository/pg` contains PostgreSQL persistence logic.
- `internal/repository/models` contains database models used by GORM.
- `internal/transport/http/v1` contains HTTP handlers and request parsing.
- `internal/infrastructure` contains external integrations such as Telegram, MinIO, QR generation, and Google Sheets export.
- `pkg` contains lower-level technical packages such as database and JWT setup.
- `cmd/api` contains the application entrypoint and dependency wiring.

This is a pragmatic clean-architecture variant: domain and use-case logic stay separated from transport and infrastructure, but the project remains compact enough for a commercial pet project.

## Core Capabilities

### Catalog and Checkout
- Browse lots through a public API
- Create buyer orders
- Retrieve buyer order history
- Preserve order item snapshots for reliable post-purchase order details

### Inventory Operations
- Manage lots with prices, stock, status, photos, and warehouse assignment
- Generate QR codes for lots
- Upload and remove lot photos
- Filter inventory across tire, rim, and accessory-specific attributes

### Warehousing and Transfers
- Manage warehouses
- Create transfers between warehouses
- Accept or cancel transfers
- Track transfer items and stock flow

### Order Operations
- List staff orders
- Update order status
- Send messages to buyers through the client Telegram bot
- Persist order message threads
- Receive buyer replies through a webhook

### Admin Operations
- Profit and loss reports
- Inventory and P&L export to Google Sheets
- User management and role changes
- Audit log browsing with filters
- Internal notification center for admins

## Telegram Bot Topology

The backend supports **two Telegram bots** with distinct responsibilities.

### 1. Staff / Internal Bot
Configured through:
- `TELEGRAM_BOT_TOKEN`

Used for:
- internal admin notifications,
- staff-facing communication flows,
- operational alerts.

### 2. Client Bot
Configured through:
- `CLIENT_TELEGRAM_BOT_TOKEN`

Used for:
- buyer-facing Telegram Mini App authentication,
- buyer message delivery related to orders,
- receiving buyer replies through a webhook.

This separation matters. Buyer communication should go through the bot the buyer already interacted with, while internal alerts should stay inside the staff/admin bot channel.

## API Surface

High-level route groups:

### Public
- `POST /api/v1/auth/telegram`
- `GET /api/v1/lots`
- `POST /api/v1/telegram/client/webhook`

### Buyer
- `POST /api/v1/orders`
- `GET /api/v1/orders`

### Staff
- `GET /api/v1/staff/lots`
- `POST /api/v1/staff/lots`
- `PUT /api/v1/staff/lots/:id`
- `DELETE /api/v1/staff/lots/:id`
- `GET /api/v1/staff/lots/:id/qr`
- `POST /api/v1/staff/lots/upload`
- `DELETE /api/v1/staff/lots/:id/photos`
- `GET /api/v1/staff/orders`
- `PATCH /api/v1/staff/orders/:id/status`
- `POST /api/v1/staff/orders/:id/message`
- `GET /api/v1/staff/orders/:id/messages`
- `GET /api/v1/staff/transfers`
- `GET /api/v1/staff/transfers/:id`
- `POST /api/v1/staff/transfers`
- `POST /api/v1/staff/transfers/:id/accept`
- `POST /api/v1/staff/transfers/:id/cancel`
- `GET /api/v1/staff/warehouses`

### Admin
- `GET /api/v1/admin/reports/pnl`
- `GET /api/v1/admin/exports/inventory`
- `GET /api/v1/admin/exports/pnl`
- `GET /api/v1/admin/users`
- `POST /api/v1/admin/users`
- `PUT /api/v1/admin/users/:id/role`
- `DELETE /api/v1/admin/users/:id`
- `POST /api/v1/admin/warehouses`
- `PUT /api/v1/admin/warehouses/:id`
- `DELETE /api/v1/admin/warehouses/:id`
- `GET /api/v1/admin/audit-logs`
- `GET /api/v1/admin/notifications`
- `POST /api/v1/admin/notifications/:id/read`

For the exact request and response contracts, use the generated Swagger docs.

## Tech Stack

- Go
- Gin
- GORM
- PostgreSQL
- JWT
- Telegram Bot API
- MinIO (S3-compatible object storage)
- Google Sheets API
- Swaggo / Swagger
- Docker / Docker Compose

## Integrations

### PostgreSQL
Primary transactional database for users, warehouses, lots, orders, transfers, audit logs, and notifications.

### MinIO
Used for storing lot photos and serving them through public URLs.

### Telegram
Used for:
- Mini App authentication,
- buyer order messaging,
- internal admin notifications.

### Google Sheets
Used for admin export flows, including inventory and P&L exports.

### QR Code Generation
Used for warehouse operations and printable lot labels.

## Environment Variables

Create a `.env` file in the repository root. A working example is provided in `.env.example`.

| Variable | Required | Description |
| --- | --- | --- |
| `HTTP_PORT` | No | API port. Default: `8083` |
| `ENV` | No | Runtime environment name, e.g. `local` |
| `POSTGRES_HOST` | Yes | PostgreSQL host |
| `POSTGRES_PORT` | Yes | PostgreSQL port |
| `POSTGRES_USER` | Yes | PostgreSQL username |
| `POSTGRES_PASSWORD` | Yes | PostgreSQL password |
| `POSTGRES_DB` | Yes | PostgreSQL database name |
| `POSTGRES_SSLMODE` | No | PostgreSQL SSL mode |
| `JWT_SECRET` | Yes | JWT signing secret |
| `JWT_TTL` | No | JWT lifetime. Default: `72h` |
| `TELEGRAM_BOT_TOKEN` | Yes | Staff/internal Telegram bot token |
| `CLIENT_TELEGRAM_BOT_TOKEN` | Yes | Buyer/client Telegram bot token |
| `CLIENT_BOT_WEBHOOK_URL` | Recommended | Public webhook URL for buyer replies |
| `MINIO_ENDPOINT` | Yes | MinIO endpoint |
| `MINIO_ACCESS_KEY` | Yes | MinIO access key |
| `MINIO_SECRET_KEY` | Yes | MinIO secret key |
| `MINIO_BUCKET_NAME` | Yes | MinIO bucket name |
| `MINIO_PUBLIC_URL` | Yes | Public base URL for stored files |
| `MINIO_USE_SSL` | No | Whether MinIO uses SSL |
| `GOOGLE_SPREADSHEET_ID` | Optional | Spreadsheet used for export workflows |

## Local Development

### Prerequisites
- Go toolchain compatible with the project
- Docker and Docker Compose, or local PostgreSQL + MinIO
- Telegram bot tokens for both staff and client flows
- Optional: Google service credentials if you plan to use Sheets export

### 1. Prepare Environment
```bash
cp .env.example .env
```

Then update secrets and environment-specific values.

### 2. Run Infrastructure with Docker Compose
```bash
docker compose up -d db minio pgadmin
```

Available local services from `docker-compose.yml`:
- API: `http://localhost:8083`
- MinIO API: `http://localhost:9000`
- MinIO Console: `http://localhost:9001`
- PgAdmin: `http://localhost:5050`

### 3. Run the API
```bash
go run ./cmd/api/main.go
```

On startup the application:
- loads configuration,
- connects to PostgreSQL,
- auto-migrates database tables,
- ensures the MinIO bucket exists,
- wires Telegram senders and webhook support,
- seeds a default warehouse if none exists.

### 4. Run Everything with Docker Compose
If you want to run the API in Docker as well:
```bash
docker compose up --build
```

## Swagger

The API exposes Swagger UI at:
- `http://localhost:8083/swagger/index.html`

If you update handler annotations and want to regenerate Swagger docs:
```bash
swag init -g cmd/api/main.go
```

## Health Check

A simple health route is available at:
- `GET /health`

## Telegram Webhook Setup

Buyer replies to order messages are received through the client bot webhook:
- `POST /api/v1/telegram/client/webhook`

For this to work reliably, `CLIENT_BOT_WEBHOOK_URL` must point to a **public HTTPS endpoint** that Telegram can reach.

Example:
```env
CLIENT_BOT_WEBHOOK_URL=https://api.example.com/api/v1/telegram/client/webhook
```

The application can ensure the webhook automatically on startup when this value is configured.

Important constraints:
- Telegram cannot deliver webhooks to plain `localhost`.
- For local development, use a tunnel or a public dev URL.
- Buyer replies can be associated with an order thread only when the message context allows it.

## Project Structure

```text
cmd/
  api/                 Application entrypoint and wiring
internal/
  config/              Environment config loading
  domain/              Business entities, DTOs, contracts
  infrastructure/      Telegram, storage, QR, Sheets integrations
  repository/
    models/            GORM models
    pg/                PostgreSQL repositories
  service/             Business logic and orchestration
  transport/
    http/
      middleware/      Auth middleware
      v1/              HTTP handlers
pkg/
  database/            PostgreSQL bootstrapping
  jwt/                 JWT utilities
docs/                  Generated Swagger assets
tma-test/              Local HTML test page for Mini App experiments
```

## Workflow Notes

### Order Messaging
Order communication is modeled as a thread attached to `order_id`, not as a generic free-form chat.

That decision keeps conversations operationally clean:
- one buyer can have multiple orders,
- each order keeps its own message history,
- staff communication does not get mixed across unrelated orders.

### Auditability
Operational changes are designed to be inspectable via audit logs and admin notifications. This is useful for warehouse environments where state changes should remain traceable.

### Snapshot-Based Order Items
Order items preserve product snapshots so historical orders remain readable even if the source lot later changes or disappears from the active catalog.

## Testing

Run the full test suite:
```bash
go test ./...
```

## What This Repository Is Optimized For

This backend is a strong fit for:
- a commercial pet project that should look and behave like a real internal platform,
- a portfolio project demonstrating end-to-end product thinking,
- a Telegram-native commerce workflow with warehouse operations,
- a compact back office with practical admin tooling.

It is not positioned as a generic marketplace framework. It is intentionally specialized around automotive retail and operations.

## License

This repository currently includes a `LICENSE` file. Use that file as the source of truth for licensing terms.
