# TacticBoard

TacticBoard is a Go-backed collaborative tactics whiteboard with WebSocket fanout, capability tokens, and optional Postgres persistence. The server is built with Gin, Fx dependency injection, Zap logging, and a pluggable storage layer (in-memory by default).

## Features

- REST API for room lifecycle (`/api/rooms`, `/api/rooms/:id`, `/api/rooms/:id/share`, `/api/health`)
- WebSocket hub with ordered operation broadcast, ping/pong heartbeats, and capability-based authorization
- In-memory store with hooks for snapshots and op history
- HMAC capability tokens for view/edit roles
- Configurable per-IP rate limiting and CORS allowlisting for REST endpoints
- Fx-wired modules for config, logging, store, HTTP, and WebSocket hub

## Getting Started

1. Copy the example environment file and update secrets:
   ```bash
   cp .env.example .env
   ```
2. Ensure `JWT_SECRET` is at least 16 characters.
3. (Optional) Set `APP_ALLOWED_ORIGINS` to your SPA origin (e.g., `http://localhost:5173`).
4. Adjust `API_RATE_RPS` / `API_RATE_BURST` if you need different REST limits.
5. Run the server locally:
   ```bash
   make dev
   ```

The API listens on the host/port configured in `.env` (defaults to `0.0.0.0:8080`).

### REST Overview

- `POST /api/rooms` – create a new room and receive view/edit capability tokens
- `GET /api/rooms/:id` – fetch room metadata and latest snapshot (if available)
- `POST /api/rooms/:id/share` – mint an additional capability token for a role
- `GET /api/health` – lightweight health probe

### WebSocket Flow

1. Connect to `/ws/room/:id` and immediately send a `hello` message:
   ```json
   {"type":"hello","roomId":"abc123","cap":"edit","since":0,"token":"<capability-token>"}
   ```
2. The server responds with the latest snapshot (if any) and any deltas since `since`.
3. Editors can send ordered op batches:
   ```json
   {"type":"op","roomId":"abc123","seq":42,"ops":[{"k":"move","id":"n1","x":120,"y":180}]}
   ```
4. All clients receive delta broadcasts and heartbeat `ping`/`pong` frames every ~20 seconds.

## Development Scripts

- `make dev` – run the server in development mode
- `make build` – build the binary to `./bin/tacticboard`
- `make test` – execute Go tests
- `make lint` – run `golangci-lint` if installed

## Docker

A multi-stage Dockerfile is provided:

```bash
docker build -t tacticboard .
docker run --env-file .env -p 8080:8080 tacticboard
```

## Configuration Reference

- `APP_ALLOWED_ORIGINS` – comma-delimited list of origins allowed by CORS (required in production)
- `API_RATE_RPS` / `API_RATE_BURST` – per-IP REST rate limiting (default 5 rps / burst 10)
- `DB_ENABLE` + `DB_DSN` – enable Postgres-backed storage via GORM (default in-memory)
- `WS_WRITE_BUFFER`, `WS_READ_LIMIT` – tune WebSocket buffers and max payload sizes

## Migrations

Starter SQL lives in `migrations/`. Integrate with your migration tool of choice (e.g., Goose, Atlas) before enabling Postgres persistence.

## Testing

Execute unit tests with:

```bash
make test
```

Add integration tests (WebSocket, REST) as flows mature. The store layer currently includes coverage for sequencing and snapshots.

## License

MIT

## Frontend

The React/Vite UI lives in `ui/`. Use `pnpm` to install dependencies, run `pnpm dev` for local development, and `pnpm build` to produce assets for the Go server's `web/` directory. See `ui/README.md` for full details on the canvas layout, state store, and testing commands.
