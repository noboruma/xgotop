![](./images/logo.png)

# xgotop - Realtime Go Runtime Visualization

A powerful eBPF-based tool for monitoring and visualizing Goroutine events in realtime with a beautiful web UI.

## Features

- **Realtime monitoring** of Go runtime events via eBPF uprobes
- **Web UI** with timeline visualization and goroutine memory allocations
- **Multiple storage formats**: Protobuf (default) and JSONL
- **Session replay** to replay past observations
- **Watch by binary or PID** of the target Go program

> [!NOTE]
> `xgotop` only supports `arm64` architecture for now, we will rollout support for `amd64` soon!

## Usage

Running `xgotop` is relatively straightforward. Get a Go binary ready, and run each of the commands below in a separate terminal:

1. Web UI

```bash
make web-dev
```

2. `xgotop`

Compile the program:

```bash
make compile
```

Run `xgotop` with a binary path:

```bash
sudo ./xgotop -b <GO_BINARY_PATH> -web -web-port 8080
```

or with a process ID:

```bash
sudo ./xgotop -pid <PROCESS_ID> -web -web-port 8080
```

Then open [localhost:5173](http://localhost:5173), and you will see this:

![](./images/ui.png)

## How Does it Work?

![](./images/systemdesign.png)

## Runtime Metrics

![](./images/r16_p1_w1_protobuf.png)