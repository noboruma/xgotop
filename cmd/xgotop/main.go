package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
	"github.com/google/uuid"

	"go.sazak.io/xgotop/cmd/xgotop/api"
	"go.sazak.io/xgotop/cmd/xgotop/storage"
)

var (
	binaryPath     = flag.String("b", "", "Path to binary file to attach the eBPF programs to")
	pid            = flag.Int("pid", 0, "PID of the running process to attach the eBPF programs to")
	readWorkers    = flag.Int("rw", 3, "Number of perf event buffer read workers")
	processWorkers = flag.Int("pw", 5, "Number of event processing workers")

	// Web mode flags
	webMode       = flag.Bool("web", false, "Enable web mode with API server and WebSocket")
	webPort       = flag.Int("web-port", 8080, "Port for web API server")
	storageFormat = flag.String("storage-format", "binary", "Storage format: binary, jsonl, or sqlite")
	storageDir    = flag.String("storage-dir", "./sessions", "Directory for storing session data")

	silent           = flag.Bool("s", false, "Enable silent mode")
	metricFileSuffix = flag.String("mfs", "", "Suffix for metric file name")
)

const (
	symbolCasgstatus = "runtime.casgstatus"
	symbolMakeslice  = "runtime.makeslice"
	symbolMakemap    = "runtime.makemap"
	symbolNewobject  = "runtime.newobject"
	symbolNewproc1   = "runtime.newproc1"
	symbolGoexit1    = "runtime.goexit1"

	statsInterval = 1000 * time.Millisecond
)

var (
	// Global storage and API server for web mode
	eventStore storage.EventStore
	apiServer  *api.Server
)

func main() {
	log.SetPrefix("xgotop: ")
	log.SetFlags(log.Ltime)

	flag.Parse()

	if *readWorkers <= 0 {
		log.Fatal("-rw must be positive")
	}

	if *processWorkers <= 0 {
		log.Fatal("-pw must be positive")
	}

	// Validate that either -b or -pid is provided
	if *binaryPath == "" && *pid == 0 {
		log.Fatal("either -b or -pid must be provided")
	}

	if *binaryPath != "" && *pid != 0 {
		log.Fatal("only one of -b or -pid can be provided")
	}

	// Determine the executable path
	var executablePath string
	if *pid != 0 {
		// Read the executable path from /proc/<pid>/exe
		exePath := fmt.Sprintf("/proc/%d/exe", *pid)
		resolvedPath, err := os.Readlink(exePath)
		must(err, fmt.Sprintf("reading executable path for PID %d", *pid))
		executablePath = resolvedPath
		log.Printf("Attaching to PID %d (executable: %s)", *pid, executablePath)
	} else {
		executablePath = *binaryPath
		log.Printf("Attaching to executable: %s", executablePath)
	}

	// Initialize web mode if enabled
	if *webMode {
		manager, err := storage.NewManager(*storageDir)
		must(err, "creating storage manager")

		session := &storage.Session{
			ID:         uuid.New().String(),
			StartTime:  time.Now(),
			PID:        *pid,
			BinaryPath: executablePath,
		}

		eventStore, err = manager.CreateSession(context.Background(), session, *storageFormat)
		must(err, "creating event store")
		defer eventStore.Close()

		apiServer = api.NewServer(manager, *webPort)
		go func() {
			if err := apiServer.Start(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("API server error: %v", err)
			}
		}()
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := apiServer.Stop(ctx); err != nil {
				log.Printf("Error stopping API server: %v", err)
			}
		}()

		// Update session on exit
		defer func() {
			endTime := time.Now()
			session.EndTime = &endTime
			session.EventCount = eventStore.GetSession().EventCount
			if err := eventStore.UpdateSession(session); err != nil {
				log.Printf("Error updating session: %v", err)
			}
		}()

		log.Printf("Web mode enabled: http://localhost:%d", *webPort)
		log.Printf("Session ID: %s", session.ID)
		log.Printf("Storage format: %s", *storageFormat)
	}

	// Subscribe to signals for terminating the program.
	stopper := make(chan os.Signal, 1)
	signal.Notify(stopper, os.Interrupt, syscall.SIGTERM)

	// Allow the current process to lock memory for eBPF resources.
	err := rlimit.RemoveMemlock()
	must(err, "locking memory")

	// Load pre-compiled programs and maps into the kernel.
	objs := ebpfObjects{}
	err = loadEbpfObjects(&objs, nil)
	must(err, "loading objects")
	defer objs.Close()

	// Open an ELF binary and read its symbols.
	ex, err := link.OpenExecutable(executablePath)
	must(err, "opening executable")

	probes := map[string]*ebpf.Program{
		symbolCasgstatus: objs.UprobeCasgstatus,
		symbolMakeslice:  objs.UprobeMakeslice,
		symbolMakemap:    objs.UprobeMakemap,
		symbolNewobject:  objs.UprobeNewobject,
		symbolNewproc1:   objs.UprobeNewproc1,
		symbolGoexit1:    objs.UprobeGoexit1,
	}

	// Configure uprobe options based on whether we're attaching to a PID
	uprobeOpts := &link.UprobeOptions{}
	if *pid != 0 {
		uprobeOpts.PID = *pid
		log.Printf("Attaching uprobes to PID %d only", *pid)
	}

	probesAttachedAt := time.Now()

	for symbol, probe := range probes {
		uprobe, err := ex.Uprobe(symbol, probe, uprobeOpts)
		must(err, "attaching uprobe at "+symbol)
		defer uprobe.Close()
	}

	rd, err := ringbuf.NewReader(objs.Events)
	must(err, "creating events ringbuf reader")
	defer rd.Close()

	eventCh := make(chan *ebpfGoRuntimeEventT, 1_000_000)

	var eventCount atomic.Int64
	var lastEventCount atomic.Int64
	eventCount.Store(0)
	lastEventCount.Store(0)

	var readEventCount atomic.Uint64
	var procEventCount atomic.Uint64
	readEventCount.Store(0)
	procEventCount.Store(0)

	// [lock] {
	// 	probeDurationNsSum u64
	// 	probeDurationNsCount u64
	// }
	var probeDurationNsSum atomic.Int64
	var probeDurationNsCount atomic.Int64
	probeDurationNsSum.Store(0)
	probeDurationNsCount.Store(0)

	readersStopped := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	var readWg, processWg sync.WaitGroup

	readWg.Add(*readWorkers)
	processWg.Add(*processWorkers)

	// Metrics
	metricRPS := make([]float64, 0, 1_000)
	metricPPS := make([]float64, 0, 1_000)
	metricEWP := make([]float64, 0, 1_000)
	metricLAT := make([]float64, 0, 1_000)
	metricTimestamps := make([]float64, 0, 1_000)

	// Monitor for stop signals and close the ringbuffer reader to unblock all readers
	go func() {
		<-stopper
		log.Printf("[Main] Received stop signal, closing ringbuffer reader")
		if err := rd.Close(); err != nil {
			log.Printf("[Main] Error closing ringbuffer reader: %v", err)
		}
		cancel()
	}()

	go func(stopped chan struct{}) {
		t := time.NewTicker(statsInterval)
		defer t.Stop()

		for {
			select {
			case <-stopped:
				return
			case <-t.C:
				// Reader/processor stats
				readEvs := readEventCount.Swap(0)
				procEvs := procEventCount.Swap(0)
				rps := float64(readEvs) * float64(time.Second) / float64(statsInterval)
				pps := float64(procEvs) * float64(time.Second) / float64(statsInterval)
				log.Printf("[Stats] RPS: %.2f ev/sec (%.2f ev/sec per worker)", rps, rps/float64(*readWorkers))
				log.Printf("[Stats] PPS: %.2f ev/sec (%.2f ev/sec per worker)", pps, pps/float64(*processWorkers))

				// Queue stats
				ec := eventCount.Load()
				lec := lastEventCount.Swap(ec)
				ediff := ec - lec
				sign := "+"
				if ediff < 0 {
					sign = ""
				}
				log.Printf("[Stats] EWP: %d (%s%d)", ec, sign, ediff)

				// Latency stats
				var lat float64
				latCnt := probeDurationNsCount.Load()
				if latCnt != 0 {
					latSum := probeDurationNsSum.Load()
					latAvg := latSum / latCnt
					latPerc := float64(latSum) / float64(time.Since(probesAttachedAt).Nanoseconds()) * 100.0
					lat = float64(latAvg)

					var prog any
					if *binaryPath == "" {
						prog = *pid
					} else {
						prog = *binaryPath
					}

					log.Printf("[Stats] LAT: %d (%.2f%% of %v)\n\n", latAvg, latPerc, prog)
				} else {
					log.Printf("[Stats] LAT: NaN\n\n")
					lat = 0
				}

				metricRPS = append(metricRPS, rps)
				metricPPS = append(metricPPS, pps)
				metricEWP = append(metricEWP, float64(ec))
				metricLAT = append(metricLAT, lat)
				metricTimestamps = append(metricTimestamps, float64(time.Now().UTC().UnixNano()))
			}
		}
	}(readersStopped)

	for i := range *readWorkers {
		go func(ctx context.Context, id int, wg *sync.WaitGroup, rd *ringbuf.Reader) {
			defer func() {
				wg.Done()
				log.Printf("[RW-%d] I'm done!", i)
			}()
			log.Printf("[RW-%d] I'm alive!", i)

			for {
				event, err := reader(rd)
				if err != nil {
					if errors.Is(err, ringbuf.ErrClosed) {
						log.Printf("[RW-%d] Ringbuffer closed, exiting", i)
						return
					}

					log.Printf("[RW-%d] Read error: %v", i, err)
					continue
				}

				eventCh <- event
				eventCount.Add(1)
				readEventCount.Add(1)
			}
		}(ctx, i, &readWg, rd)
	}

	for i := range *processWorkers {
		go func(id int, wg *sync.WaitGroup, eventCh chan *ebpfGoRuntimeEventT, readersStopped chan struct{}) {
			defer func() {
				wg.Done()
				log.Printf("[PW-%d] I'm done!", i)
			}()
			log.Printf("[PW-%d] I'm alive!", i)

			for {
				select {
				case <-readersStopped:
					log.Printf("[PW-%d] Context is cancelled, ctx err: %v, draining events channel", i, ctx.Err())
					for event := range eventCh {
						eventCount.Add(-1)
						procEventCount.Add(1)
						probeDurationNsCount.Add(1)
						probeDurationNsSum.Add(int64(event.ProbeDurationNs))
						processEvent(i, event)
					}
					log.Printf("[PW-%d] Draining events channel complete", i)
					return
				case event, ok := <-eventCh: // ', ok' idiom is used to prevent race condition
					if !ok {
						// Channel is closed, no more events
						return
					}
					eventCount.Add(-1)
					procEventCount.Add(1)
					probeDurationNsCount.Add(1)
					probeDurationNsSum.Add(int64(event.ProbeDurationNs))
					processEvent(i, event)
				}
			}
		}(i, &processWg, eventCh, readersStopped)
	}

	log.Printf("All readers are alive")

	readWg.Wait()

	log.Printf("All readers are done")
	close(readersStopped) // signal to processors that no more events will be coming
	close(eventCh)

	processWg.Wait()
	log.Printf("All processors are done")

	metrics := struct {
		Rps []float64 `json:"rps"`
		Pps []float64 `json:"pps"`
		Ewp []float64 `json:"ewp"`
		Lat []float64 `json:"lat"`
		Ts  []float64 `json:"ts"`
	}{
		Rps: metricRPS,
		Pps: metricPPS,
		Ewp: metricEWP,
		Lat: metricLAT,
		Ts:  metricTimestamps,
	}
	b, err := json.MarshalIndent(metrics, "", "  ")
	must(err, "marshaling metric data")
	if *metricFileSuffix != "" {
		*metricFileSuffix = "_" + *metricFileSuffix
	}
	err = os.WriteFile(
		fmt.Sprintf("metrics_%s%s.json", time.Now().UTC().Format("2006-01-02-15-04-05"), *metricFileSuffix),
		b, 0666,
	)
	must(err, "writing metrics")
}

func must(err error, op string) {
	if err != nil {
		log.Fatalf("%s: %v", op, err)
	}
}

//go:inline
func reader(rd *ringbuf.Reader) (*ebpfGoRuntimeEventT, error) {
	record, err := rd.Read()
	if err != nil {
		return nil, err
	}
	var event ebpfGoRuntimeEventT
	if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &event); err != nil {
		return nil, fmt.Errorf("parsing event: %v", err)
	}

	return &event, nil
}

//go:inline
func processEvent(id int, event *ebpfGoRuntimeEventT) {
	// Convert to storage event format
	storageEvent := &storage.Event{
		Timestamp:       event.Timestamp,
		EventType:       storage.EventType(event.EventType),
		Goroutine:       event.Goroutine,
		ParentGoroutine: event.ParentGoroutine,
		Attributes:      event.Attributes,
	}

	// Write to storage if web mode is enabled
	if *webMode && eventStore != nil {
		if err := eventStore.WriteEvent(storageEvent); err != nil {
			log.Printf("[PW-%d] Failed to write event to storage: %v", id, err)
		}

		// Broadcast to WebSocket clients
		if apiServer != nil {
			apiServer.BroadcastEvent(storageEvent)
		}
	}

	// Log events (can be disabled in web mode if desired)
	if !*webMode && !*silent {
		switch event.EventType {
		case 0:
			log.Printf("[PW-%d] [ts:%d,lat:%d] goroutine %d state %d -> %d", id, event.Timestamp, event.ProbeDurationNs, event.Attributes[2], event.Attributes[0], event.Attributes[1])
		case 1:
			log.Printf("[PW-%d] [ts:%d,lat:%d] goroutine %d allocated slice []%s with length %d and capacity %d", id, event.Timestamp, event.ProbeDurationNs, event.Goroutine, kindToString(Kind(event.Attributes[1])), event.Attributes[2], event.Attributes[3])
		case 2:
			log.Printf("[PW-%d] [ts:%d,lat:%d] goroutine %d allocated map[%s]%s with initial capacity %d", id, event.Timestamp, event.ProbeDurationNs, event.Goroutine, kindToString(Kind(event.Attributes[1])), kindToString(Kind(event.Attributes[2])), event.Attributes[3])
		case 3:
			log.Printf("[PW-%d] [ts:%d,lat:%d] goroutine %d allocated object of size %d and kind %s", id, event.Timestamp, event.ProbeDurationNs, event.Goroutine, event.Attributes[0], kindToString(Kind(event.Attributes[1])))
		case 4:
			log.Printf("[PW-%d] [ts:%d,lat:%d] goroutine %d created new goroutine %d", id, event.Timestamp, event.ProbeDurationNs, event.Attributes[0], event.Attributes[1])
		case 5:
			log.Printf("[PW-%d] [ts:%d,lat:%d] goroutine %d exited", id, event.Timestamp, event.ProbeDurationNs, event.Attributes[0])
		default:
			log.Printf("[PW-%d] UNKNOWN EVENT TYPE: %d", id, event.EventType)
		}
	}
}

//go:inline
func kindToString(kind Kind) string {
	switch kind {
	case Invalid:
		return "INVALID"
	case Bool:
		return "bool"
	case Int:
		return "int"
	case Int8:
		return "int8"
	case Int16:
		return "int16"
	case Int32:
		return "int32"
	case Int64:
		return "int64"
	case Uint:
		return "uint"
	case Uint8:
		return "uint8"
	case Uint16:
		return "uint16"
	case Uint32:
		return "uint32"
	case Uint64:
		return "uint64"
	case Uintptr:
		return "uintptr"
	case Float32:
		return "float32"
	case Float64:
		return "float64"
	case Complex64:
		return "complex64"
	case Complex128:
		return "complex128"
	case Array:
		return "ARRAY"
	case Chan:
		return "chan T"
	case Func:
		return "func"
	case Interface:
		return "interface{}"
	case Map:
		return "map[K]V"
	case Pointer:
		return "*T"
	case Slice:
		return "[]T"
	case String:
		return "string"
	case Struct:
		return "struct{}"
	case UnsafePointer:
		return "unsafe.Pointer"
	default:
		return fmt.Sprintf("Unknown kind: %d", kind)
	}
}
