# internal/liveview

WebSocket-based live graph viewer with REST API endpoints and C4 architecture diagram serving.

## Overview

The liveview package provides a real-time visualization server for the Sentinel dependency graph. It implements `graph.Observer` to receive scan events and broadcasts them to connected WebSocket clients. A bundled REST API serves graph snapshots, task status, source code, and ADR documents.

## Key Types

### `Server`
WebSocket hub that implements `graph.Observer`. Manages client connections with registration, broadcast, and keepalive (ping/pong).
- `NewServer()` — creates hub with buffered broadcast channel (256 events)
- `Run(ctx)` — blocking hub event loop (select on register/unregister/broadcast)
- `Notify(GraphEvent)` — non-blocking push to broadcast channel (implements `graph.Observer`)
- `StartHTTP(port, db)` — registers HTTP routes and starts blocking server

Connection management: single `writePump` goroutine per client (Gorilla WebSocket concurrency model), `readPump` monitors pong timeouts, origin validation restricts connections to `localhost`/`127.0.0.1`.

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/ws` | WebSocket upgrade for live graph events |
| `GET` | `/api/graph` | Full graph snapshot (all nodes + edges as JSON) |
| `GET` | `/api/status` | Latest task status from SQLite |
| `GET` | `/api/code?path=&start=&end=` | Read file content with line range |
| `GET` | `/api/adr` | List ADR files from `docs/architecture/adr/` |
| `GET` | `/api/adr/{filename}` | Read specific ADR file content |
| `GET` | `/` | Static file server for `web/dist/` (Vite build) |

All endpoints set CORS headers for local development. The code endpoint enforces path traversal protection.

## Dependencies

- `internal/graph` — observer interface, Node/Edge types
- `pkg/sqlite` — DB validation
- `github.com/gorilla/websocket` — WebSocket implementation

## Usage

```go
db, _ := sqlite.Init()
server := liveview.NewServer()
graphEngine.RegisterObserver(server) // server implements graph.Observer
go server.Run(context.Background())
server.StartHTTP(8080, db)
```
