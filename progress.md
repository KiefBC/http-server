# Phase 1

## Resources

None.

## Steps

1. We implemented a simple 8 byte reader with a local file.
2. We realized reading 8 bytes at a time would make the output unreadable so we implemented reading lines seperated by a `\n` and building up any lines that ended without said `\n`.
3. We implemented a goroutine and channel to read and print out to the stdout as the lines are parsed, mocking receiving data via TCP.

# Phase 2

## Resources

[RFC 9112](https://datatracker.ietf.org/doc/html/rfc9112)
[RFC 9110](https://datatracker.ietf.org/doc/html/rfc9110)

## Steps

1. We transitioned from reading local files to accepting data and connections via TCP.
2. Took what we did in our TCP and copied it into a new file and setup our first UDP server.

# Phase 3

## Resouces

[testify](https://github.com/stretchr/testify)

## Steps

1. We implemented a request-line (aka start-line which is a response) parser for structured data parsing. Example: "GET /coffee HTTP/1.1"
2. Updated our request-line parser to continually read chunks of data
3. Implemented a state-machine to handle our reading and parsing
4. Connected our TCP Listener to the new parsing by removing the current channel.

# Phase 4

## Resources

[RFC 9112](https://datatracker.ietf.org/doc/html/rfc9112)
[RFC 9110](https://datatracker.ietf.org/doc/html/rfc9110)

## Steps

1. Implemented Header parsing to go along with our request-line and start-line. Example: "Host: localhost:42069\r\n\r\n"
2. Added logic for checking valid characters based on rules from [RFC 9110 5.1](https://datatracker.ietf.org/doc/html/rfc9110) and [RFC 9110 5.6.2](https://datatracker.ietf.org/doc/html/rfc9110#name-tokens)
3. Added logic for checking for multiple headers [RFC 9110 5.2](https://datatracker.ietf.org/doc/html/rfc9110#name-field-lines-and-combined-fi). Example: "Set-Person: lane-loves-go, prime-loves-zig, tj-loves-ocaml"
4. Added Header parsing to our existing request-line parsing to our `internal/request/request.go`
