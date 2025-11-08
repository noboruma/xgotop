# xgotop - Real-time Go Runtime Event Visualization

A powerful eBPF-based tool for monitoring and visualizing Go runtime events with a beautiful brutalist web UI.

## Features

- **Real-time monitoring** of Go runtime events via eBPF uprobes
- **Web UI** with timeline visualization showing goroutine states and allocations
- **Multiple storage formats**: Binary (default), JSONL, and SQLite
- **Session replay** to analyze past recordings
- **Zoom and pan controls** for detailed timeline exploration
- **PID or binary attachment** for flexible monitoring

## Architecture

### Backend (Go + eBPF)
- eBPF uprobes attached to Go runtime functions
- HTTP API server with REST endpoints
- WebSocket support for real-time event streaming
- Three storage formats:
  - **Binary**: Compact, fast (default)
  - **JSONL**: Human-readable, streamable
  - **SQLite**: Queryable, supports complex filters

### Frontend (React + TypeScript)
- Brutalist UI design with bold borders and stark contrasts
- Canvas-based timeline rendering for performance
- Configurable nanoseconds-per-pixel ratio
- Live and replay modes
- Zustand state management

## Quick Start

### Prerequisites

- Go 1.25+
- Node.js 18+
- Linux with eBPF support (kernel 5.8+)
- sudo/root access (for eBPF)

### Build

```bash
# Build backend
make compile

# Install frontend dependencies
make web-install
```

### Usage

#### CLI Mode (No Web UI)

```bash
# Monitor by binary path
sudo ./xgotop -b /path/to/binary

# Monitor by PID
sudo ./xgotop -pid 12345
```

#### Web Mode

```bash
# Start xgotop with web UI
make run-web

# Or manually:
sudo ./xgotop -b ./testserver -web -web-port 8080 -storage-dir ./sessions

# In another terminal, start the frontend
make web-dev
```

Then open http://localhost:5173 in your browser.

## Command Line Options

```
-b string
    Path to binary file to attach the eBPF programs to
-pid int
    PID of the running process to attach the eBPF programs to
-rw int
    Number of perf event buffer read workers (default 3)
-pw int
    Number of event processing workers (default 5)

Web Mode Options:
-web
    Enable web mode with API server and WebSocket
-web-port int
    Port for web API server (default 8080)
-storage-format string
    Storage format: binary, jsonl, or sqlite (default "binary")
-storage-dir string
    Directory for storing session data (default "./sessions")
```

## API Endpoints

- `GET /api/sessions` - List all recorded sessions
- `GET /api/sessions/:id` - Get session metadata
- `GET /api/sessions/:id/events` - Query events with filters
- `GET /api/sessions/:id/goroutines` - Get list of goroutines
- `GET /api/config` - Get timeline configuration
- `POST /api/config` - Update timeline configuration
- `GET /ws` - WebSocket endpoint for real-time events

### Query Parameters for Events

- `goroutine` - Filter by goroutine ID
- `event_type` - Filter by event type (0-5)
- `start_time` - Filter by start timestamp (nanoseconds)
- `end_time` - Filter by end timestamp (nanoseconds)
- `limit` - Limit number of results
- `offset` - Offset for pagination

## Event Types

0. **CasGStatus** - Goroutine state changes
1. **MakeSlice** - Slice allocations
2. **MakeMap** - Map allocations
3. **NewObject** - Object allocations
4. **NewGoroutine** - Goroutine creation
5. **GoExit** - Goroutine exit

## Web UI Controls

- **Zoom**: Use `+`/`-` buttons or `Ctrl+Scroll`
- **Pan**: Scroll horizontally
- **NS/PX**: Adjust nanoseconds-per-pixel ratio for time scale
- **Live Mode**: Real-time monitoring with WebSocket
- **Replay Mode**: Load and analyze past sessions

## Development

### Backend

```bash
# Generate eBPF code
make gen

# Build
make compile

# Run tests
go test ./...
```

### Frontend

```bash
cd web

# Install dependencies
npm install

# Development server
npm run dev

# Build for production
npm run build
```

## Troubleshooting

### eBPF Issues

- Ensure you have kernel 5.8+ with eBPF support
- Run with sudo/root privileges
- Check that BTF is enabled: `ls /sys/kernel/btf/vmlinux`

### WebSocket Connection Issues

- Verify the API server is running on the correct port
- Check CORS settings if accessing from a different origin
- Update `.env.local` in the web directory with correct URLs

## License

See LICENSE file for details.

## Contributing

Contributions are welcome! Please open an issue or pull request.
