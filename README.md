# simple-bar-server-go

This is go port of Jean-Tinland/simple-bar-server using only go's standard library.

Jean Tinland's `simple-bar[-server]` is great software and I really like it.
The only reason I want a go version is because I am scared of running npm on my personal laptop due to recent events.

It seems to work for me. Please file issues if you find any bugs or the upstream behavior has changed.

It listens on the same loopback ports as the upstream server:

- HTTP commands: `127.0.0.1:7776`
- WebSocket clients: `127.0.0.1:7777`

## Build

```bash
go build -o simple-bar-server .
```

## Run

```bash
./simple-bar-server
```

The server prints:

```text
simple-bar-server running at http://localhost:7776
```

On macOS it also asks Uebersicht to refresh on startup, matching the upstream
server behavior.

## macOS LaunchAgent

To run the server as a per-user macOS service:

```bash
./scripts/install-launch-agent.sh
```

To unload and remove the LaunchAgent:

```bash
./scripts/uninstall-launch-agent.sh
```

Logs are written to `~/Library/Logs/simple-bar-server-go.out.log` and
`~/Library/Logs/simple-bar-server-go.err.log`.

## Compatibility

The HTTP and WebSocket API mirrors upstream `simple-bar-server`:

- refresh, toggle, enable, or disable simple-bar widgets
- refresh yabai spaces, windows, and displays widgets
- refresh the skhd mode indicator
- refresh the AeroSpace spaces widget, including `space` payloads
- push missive notifications to connected simple-bar clients
