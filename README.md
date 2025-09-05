# HTTP/1.1 Server in Go

## Things Learned

- Small buffers (8 bytes) teach you about streaming and state machines
- CRLF (`\r\n`) matters everywhere in HTTP — request lines, headers, chunks
- Case-insensitive headers with comma-combining for duplicates (RFC 9110)
- State machines keep request parsing and response writing in order
- Closing write side prevents "connection reset" errors
- Chunked encoding enables streaming when content length is unknown
- Trailer headers provide metadata after the body

A small learning project that implements a streaming HTTP/1.1 request parser (request-line → headers → body) and a minimal TCP server that replies with `200 OK` and headers. Built around RFC 9110/9112 concepts.

## Quick Start

Prerequisites: Go 1.22+ (go.mod lists 1.24.2; use your latest stable Go).

Run the HTTP server:

```bash
go run cmd/httpserver/main.go
```

The server listens on `:42069` and shuts down cleanly with Ctrl+C.

## Development Commands

- Run TCP listener (logs parsed requests):
  ```bash
  go run cmd/tcplistener/main.go
  ```
- Run UDP sender (stdin → UDP datagrams):
  ```bash
  go run cmd/udpsender/main.go
  ```
- Run tests:
  ```bash
  go test ./...
  ```
- Build all binaries:
  ```bash
  go build ./cmd/...
  ```
- Format / vet:
  ```bash
  go fmt ./... && go vet ./...
  ```

## Project Layout

- `cmd/httpserver/` — Minimal HTTP server with graceful shutdown.
- `cmd/tcplistener/` — Raw TCP listener that parses and prints requests.
- `cmd/udpsender/` — Interactive UDP client for manual testing.
- `internal/request/` — Streaming parser (request-line, headers, body via Content-Length).
- `internal/headers/` — Header parsing, case-insensitive keys, duplicate combining.
- `internal/response/` — Helpers to write status lines and headers.
- `internal/server/` — TCP server that returns `200 OK` with headers.

## Features

- Streaming request parsing with small-buffer growth (request-line → headers → body).
- HTTP/1.1 only; validates request-line format and version.
- Headers: case-insensitive keys; duplicate fields combined with commas.
- Body: `Content-Length` only (reads exactly N bytes; ignores extra).
- Server: responds with routing, chunked encoding, trailers, and proxy support.
- Graceful shutdown; closes write side to avoid resets.

Notes: Header output order is map-based (non-deterministic).

## Roadmap

- MVP completion
  - Return a small response body and basic routes (e.g., `/`, `/health`).
  - Proper error responses (400/500) and consistent default headers.
- Protocol hardening
  - Input limits and read deadlines; stricter header token validation.
  - Optional: `Transfer-Encoding: chunked` support; consider keep-alive.
- Testing & tooling
  - Unit tests for `internal/response` and `internal/server`.
  - Integration test that exercises end-to-end parsing/response.
- Developer experience
  - Stabilize header output order; Makefile and CI (fmt/vet/test).
  - Align `go.mod` Go version with the toolchain in use.

## Phases

See `progress.md` for phase-by-phase development notes and commit references.
