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
