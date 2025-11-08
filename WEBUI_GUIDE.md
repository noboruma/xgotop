# Web UI Guide

## Quick Start

### 1. Build Everything

```bash
# Build backend
make compile

# Install frontend dependencies
make web-install

# Build frontend
make web-build
```

### 2. Start the Backend

```bash
# Terminal 1: Start xgotop in web mode
sudo ./xgotop -b ./testserver -web -web-port 8080 -storage-dir ./sessions
```

### 3. Start the Frontend

```bash
# Terminal 2: Start the React dev server
cd web && npm run dev
```

### 4. Open the UI

Navigate to http://localhost:5173 in your browser.

## Features

### Timeline View

The main view shows a timeline of all goroutines with:
- **Top row per goroutine**: State transitions (colored boxes)
- **Middle row**: Allocations (makeslice, makemap)
- **Bottom row**: newobject calls

### Controls

- **Zoom**: Use +/- buttons or Ctrl+Scroll
- **Pan**: Scroll horizontally
- **NS/PX**: Adjust nanoseconds-per-pixel ratio
- **Live Mode**: Real-time monitoring
- **Replay Mode**: Load past sessions

### Color Coding

#### States
- Green: Idle/Runnable
- Yellow: Running
- Orange: Syscall
- Red: Waiting
- Gray: Dead
- Purple: Moribund
- Pink: Enqueue
- Teal: CopyStack
- Amber: Preempted

#### Allocations
- Blue: makeslice
- Purple: makemap
- Cyan: newobject

## API Usage

### List Sessions

```bash
curl http://localhost:8080/api/sessions
```

### Get Events

```bash
curl "http://localhost:8080/api/sessions/SESSION_ID/events?limit=100"
```

### Filter by Goroutine

```bash
curl "http://localhost:8080/api/sessions/SESSION_ID/events?goroutine=5"
```

### WebSocket Connection

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');
ws.onmessage = (event) => {
  const goroutineEvent = JSON.parse(event.data);
  console.log(goroutineEvent);
};
```

## Storage Formats

### Binary (default)
- Most compact
- Fastest writes
- Use for high-frequency events

### JSONL
- Human-readable
- Easy to parse with standard tools
- Good for debugging

### SQLite
- Queryable
- Supports complex filters
- Best for analysis

## Troubleshooting

### Backend won't start
- Check you have sudo/root privileges
- Verify eBPF is supported: `ls /sys/kernel/btf/vmlinux`

### Frontend won't connect
- Check backend is running on port 8080
- Verify WebSocket URL in `.env.local`
- Check browser console for errors

### No events appearing
- Ensure target binary/PID is actively running
- Check that uprobes attached successfully (look for error logs)
- Verify the binary is a Go binary with runtime symbols

## Tips

1. **Performance**: Use binary format for production, JSONL for debugging
2. **Zoom**: Start zoomed out to see overall patterns, zoom in for details
3. **Filtering**: Use API filters to reduce data transfer for large sessions
4. **Sessions**: Sessions are automatically created when starting in web mode


