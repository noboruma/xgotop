package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"go.sazak.io/xgotop/cmd/xgotop/storage"
)

// Config represents the timeline configuration
type Config struct {
	NanosecondsPerPixel float64           `json:"nanoseconds_per_pixel"`
	StateColors         map[string]string `json:"state_colors"`
	TypeColors          map[string]string `json:"type_colors"`
}

// Metrics represents the real-time performance metrics
type Metrics struct {
	RPS float64 `json:"rps"` // Read events per second
	PPS float64 `json:"pps"` // Processed events per second
	EWP int64   `json:"ewp"` // Events waiting processing
	LAT float64 `json:"lat"` // Average latency in nanoseconds
	PRC int64   `json:"prc"` // Processing time in nanoseconds
	BFL float64 `json:"bfl"` // Batch flush latency
	QWL float64 `json:"qwl"` // Queue wait latency
}

// Server is the HTTP API server
type Server struct {
	manager    *storage.Manager
	config     *Config
	configMu   sync.RWMutex
	hub        *Hub
	httpServer *http.Server
	metrics    *Metrics
	metricsMu  sync.RWMutex
}

// NewServer creates a new API server
func NewServer(manager *storage.Manager, port int) *Server {
	server := &Server{
		manager: manager,
		metrics: &Metrics{},
		config: &Config{
			NanosecondsPerPixel: 1000000.0, // 1ms per pixel by default
			StateColors: map[string]string{
				"0": "#22c55e", // Gidle -> green
				"1": "#3b82f6", // Grunnable -> blue
				"2": "#eab308", // Grunning -> yellow
				"3": "#f97316", // Gsyscall -> orange
				"4": "#ef4444", // Gwaiting -> red
				"5": "#a855f7", // Gmoribund_unused -> purple
				"6": "#64748b", // Gdead -> gray
				"7": "#ec4899", // Genqueue_unused -> pink
				"8": "#14b8a6", // Gcopystack -> teal
				"9": "#f59e0b", // Gpreempted -> amber
			},
			TypeColors: map[string]string{
				"makeslice": "#3b82f6",
				"makemap":   "#8b5cf6",
				"newobject": "#06b6d4",
			},
		},
		hub: NewHub(),
	}

	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/sessions", server.handleSessions)
	mux.HandleFunc("/api/sessions/", server.handleSession)
	mux.HandleFunc("/api/config", server.handleConfig)
	mux.HandleFunc("/api/metrics", server.handleMetrics)

	// WebSocket endpoint
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWs(server.hub, w, r)
	})

	// CORS middleware
	handler := corsMiddleware(mux)

	server.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
	}

	return server
}

// Start starts the API server
func (s *Server) Start() error {
	// Start the WebSocket hub
	go s.hub.Run()

	log.Printf("API server listening on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Stop stops the API server
func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// BroadcastEvent broadcasts an event to all connected WebSocket clients
func (s *Server) BroadcastEvent(event *storage.Event) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return
	}

	s.hub.Broadcast(data)
}

// BroadcastBatch broadcasts a batch of events to all connected WebSocket clients
func (s *Server) BroadcastBatch(events []*storage.Event) {
	// Send as an array for efficient client processing
	data, err := json.Marshal(map[string]interface{}{
		"type": "batch",
		"events": events,
	})
	if err != nil {
		log.Printf("Failed to marshal event batch: %v", err)
		return
	}

	s.hub.Broadcast(data)
}

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.listSessions(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := s.manager.ListSessions(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
	// Parse session ID from path
	path := r.URL.Path
	sessionID := path[len("/api/sessions/"):]

	// Remove any sub-paths
	if idx := len(sessionID); idx > 0 {
		for i, c := range sessionID {
			if c == '/' {
				idx = i
				break
			}
		}
		subPath := ""
		if idx < len(sessionID) {
			subPath = sessionID[idx:]
			sessionID = sessionID[:idx]
		}

		if subPath == "/events" {
			s.getEvents(w, r, sessionID)
			return
		} else if subPath == "/goroutines" {
			s.getGoroutines(w, r, sessionID)
			return
		}
	}

	// Get session metadata
	session, err := s.manager.GetSession(r.Context(), sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

func (s *Server) getEvents(w http.ResponseWriter, r *http.Request, sessionID string) {
	store, err := s.manager.OpenSession(r.Context(), sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer store.Close()

	// Parse query parameters
	filter := &storage.EventFilter{}

	if goroutineStr := r.URL.Query().Get("goroutine"); goroutineStr != "" {
		if gid, err := strconv.ParseUint(goroutineStr, 10, 32); err == nil {
			gid32 := uint32(gid)
			filter.Goroutine = &gid32
		}
	}

	if eventTypeStr := r.URL.Query().Get("event_type"); eventTypeStr != "" {
		if et, err := strconv.ParseUint(eventTypeStr, 10, 64); err == nil {
			eventType := storage.EventType(et)
			filter.EventType = &eventType
		}
	}

	if startTimeStr := r.URL.Query().Get("start_time"); startTimeStr != "" {
		if st, err := strconv.ParseUint(startTimeStr, 10, 64); err == nil {
			filter.StartTime = &st
		}
	}

	if endTimeStr := r.URL.Query().Get("end_time"); endTimeStr != "" {
		if et, err := strconv.ParseUint(endTimeStr, 10, 64); err == nil {
			filter.EndTime = &et
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	events, err := store.ReadEvents(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

func (s *Server) getGoroutines(w http.ResponseWriter, r *http.Request, sessionID string) {
	store, err := s.manager.OpenSession(r.Context(), sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer store.Close()

	goroutines, err := store.GetGoroutines(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goroutines)
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.configMu.RLock()
		config := s.config
		s.configMu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config)

	case http.MethodPost:
		var config Config
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		s.configMu.Lock()
		s.config = &config
		s.configMu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// UpdateMetrics updates the server's metrics
func (s *Server) UpdateMetrics(metrics *Metrics) {
	s.metricsMu.Lock()
	s.metrics = metrics
	s.metricsMu.Unlock()
}

// handleMetrics handles the /api/metrics endpoint
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.metricsMu.RLock()
	metrics := s.metrics
	s.metricsMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
