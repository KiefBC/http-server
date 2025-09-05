# The Phases For Web Server Written In Go

## Phase 1 - File Reading

### Summary

Bootstrapped local file reading with small chunking, line-buffering, and goroutine-based streaming to simulate incremental TCP input.

### Merge/Commit

- Commit: `e60a126` (initial commit), `8e627fd` (added progress log)

### Resources

None.

### Steps

1. We implemented a simple 8 byte reader with a local file.
2. We realized reading 8 bytes at a time would make the output unreadable so we implemented reading lines separated by a `\n` and building up any lines that ended without said `\n`.
3. We implemented a goroutine and channel to read and print out to the stdout as the lines are parsed, mocking receiving data via TCP.

## Phase 2 - TCP -> UDP

### Summary

Moved from files to sockets: added a TCP listener and a simple stdin-driven UDP sender for quick manual testing.

### Merge/Commit

- Commit: `ae2656c` (feat: tcp and udp)

### Resources

[RFC 9112](https://datatracker.ietf.org/doc/html/rfc9112)
[RFC 9110](https://datatracker.ietf.org/doc/html/rfc9110)

### Steps

1. We transitioned from reading local files to accepting data and connections via TCP.
2. Took what we did in our TCP and copied it into a new file and setup our first UDP server.

## Phase 3 - HTTP Parsing

### Summary

Added HTTP request-line parsing with a streaming reader and state machine; integrated into the TCP listener with tests.

### Merge/Commit

- Commit: `66afc7e` (request-line parsing and tests)

### Resources

[testify](https://github.com/stretchr/testify)

### Steps

1. Implemented an HTTP request-line parser for structured data (e.g., "GET /coffee HTTP/1.1").
2. Updated the parser to read in chunks with an expandable buffer.
3. Introduced a state machine to drive reading and parsing.
4. Connected the TCP listener to use the new parser (removed the previous channel approach).
5. Added tests for request-line parsing.

## Phase 4 - HTTP Header Parsing

### Summary

Implemented CRLF-delimited header parsing with validation and duplicate-field combining; integrated with request parsing and tests.

### Merge/Commit

- Merge: `24d4923` (headers parsed, tests, RFC notes)

### Resources

[RFC 9112](https://datatracker.ietf.org/doc/html/rfc9112)
[RFC 9110](https://datatracker.ietf.org/doc/html/rfc9110)

### Steps

1. Implemented HTTP header parsing with CRLF-delimited lines (RFC 9112 §2.2) (e.g., "Host: localhost:42069\r\n").
2. Validated header field-names/tokens and treated names case-insensitively per [RFC 9110 §5.1](https://datatracker.ietf.org/doc/html/rfc9110#section-5.1), [§5.6.2](https://datatracker.ietf.org/doc/html/rfc9110#section-5.6.2).
3. Combined duplicate header fields using comma separation per [RFC 9110 §5.2](https://datatracker.ietf.org/doc/html/rfc9110#section-5.2). Example: "Set-Person: alex-loves-go, kiefer-loves-zig, no-one-loves-ocaml".
4. Integrated header parsing into `internal/request/request.go` and expanded tests for valid/malformed headers.

## Phase 5 - HTTP Body Parsing

### Summary

Extended the parser to read message bodies using Content-Length semantics with thorough edge-case tests.

### Merge/Commit

- Merge: `fea80e2` (http body parsing and tests)

### Resources

None.

### Steps

1. Extended the parser state machine with a body stage and added a `Body` field to `Request`.
2. Implemented `Content-Length` handling per [RFC 9112 §6](https://datatracker.ietf.org/doc/html/rfc9112#section-6): read exactly N bytes into the body; ignore extra data; error if short.
3. Added tests covering normal, zero-length, short, and overlong body scenarios.

## Phase 6 - HTTP Server and Responses

### Summary

Introduced response helpers and a concurrent TCP server that returns 200 OK with default headers; added an httpserver app with graceful shutdown.

### Merge/Commit

- Merge: `4247e54` (server + response helpers + app)

### Resources

### Steps

1. Introduced response helpers to write HTTP/1.1 status lines and default headers.
2. Implemented `internal/server` to accept TCP connections and return `200 OK` with headers (no body), then close the write side to avoid resets.
3. Added `cmd/httpserver` with graceful shutdown on SIGINT/SIGTERM, separate from the raw TCP listener.

## Phase 7 - Chunked Encoding & Trailers

### Summary

Implemented HTTP/1.1 chunked transfer encoding with a response writer that streams chunks, plus optional trailer headers. Added demo endpoints for chunked output and proxy streaming.

### Merge/Commit

- Merge: `7d19489` (chunked encoding)

### Resources

[RFC 9112 §7.1](https://datatracker.ietf.org/doc/html/rfc9112#section-7.1) (Transfer-Encoding: chunked)  
[RFC 9112 §7.1.2](https://datatracker.ietf.org/doc/html/rfc9112#section-7.1.2) (Trailer fields)  
[RFC 9112 §7.1.3](https://datatracker.ietf.org/doc/html/rfc9112#section-7.1.3) (Last chunk and message completion)

### Steps

1. Added `internal/response/writer.go` with `WriteChunkedBody`, `WriteChunkedBodyDone`, `WriteTrailers`, and `WriteTrailersDone`, enforcing correct write order via an internal state machine.
2. Added `internal/response/headers.go` with `GetChunkedHeaders()` to set `Transfer-Encoding: chunked` (and omit `Content-Length`).
3. Implemented `/chunked` endpoint to stream multiple HTML chunks in a single response.
4. Implemented `/httpbin/<path>` proxy that streams upstream content as chunked, preserves `Content-Type`, sets `Trailer: X-Content-SHA256, X-Content-Length`, and computes trailer values after streaming.
5. Updated server handler integration to use the new writer API; ensured proper termination with final chunk and optional trailers.
