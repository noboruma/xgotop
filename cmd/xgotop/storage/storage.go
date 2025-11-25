package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// EventType represents the type of runtime event
type EventType uint64

const (
	EventTypeCasGStatus   EventType = 0
	EventTypeMakeSlice    EventType = 1
	EventTypeMakeMap      EventType = 2
	EventTypeNewObject    EventType = 3
	EventTypeNewGoroutine EventType = 4
	EventTypeGoExit       EventType = 5
)

// Event represents a Go runtime event
type Event struct {
	Timestamp       uint64    `json:"timestamp"`
	EventType       EventType `json:"event_type"`
	Goroutine       uint32    `json:"goroutine"`
	ParentGoroutine uint32    `json:"parent_goroutine"`
	Attributes      [5]uint64 `json:"attributes"`
}

// Session represents a recording session
type Session struct {
	ID         string     `json:"id"`
	StartTime  time.Time  `json:"start_time"`
	EndTime    *time.Time `json:"end_time,omitempty"`
	PID        int        `json:"pid,omitempty"`
	BinaryPath string     `json:"binary_path"`
	EventCount int64      `json:"event_count"`
}

// EventFilter represents filters for querying events
type EventFilter struct {
	Goroutine *uint32
	EventType *EventType
	StartTime *uint64
	EndTime   *uint64
	Limit     int
	Offset    int
}

// EventStore is the interface for storing and retrieving events
type EventStore interface {
	// WriteEvent writes a single event to storage
	WriteEvent(event *Event) error

	// WriteBatch writes multiple events to storage
	WriteBatch(events []*Event) error

	// ReadEvents reads events with optional filters
	ReadEvents(ctx context.Context, filter *EventFilter) ([]*Event, error)

	// GetGoroutines returns a list of unique goroutine IDs
	GetGoroutines(ctx context.Context) ([]uint32, error)

	// Close closes the storage
	Close() error

	// GetSession returns session metadata
	GetSession() *Session

	// UpdateSession updates session metadata
	UpdateSession(session *Session) error
}

// SessionStore manages multiple sessions
type SessionStore interface {
	// ListSessions returns all sessions
	ListSessions(ctx context.Context) ([]*Session, error)

	// GetSession returns a specific session
	GetSession(ctx context.Context, id string) (*Session, error)

	// OpenSession opens a session for reading
	OpenSession(ctx context.Context, id string) (EventStore, error)

	// CreateSession creates a new session
	CreateSession(ctx context.Context, session *Session, format string) (EventStore, error)

	// DeleteSession deletes a session
	DeleteSession(ctx context.Context, id string) error

	io.Closer
}

// Helper functions for session metadata
func saveSessionMetadata(sessionDir string, session *Session) error {
	metadataPath := filepath.Join(sessionDir, "metadata.json")
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal session metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("write session metadata: %w", err)
	}

	return nil
}

func loadSessionMetadata(sessionDir string) (*Session, error) {
	metadataPath := filepath.Join(sessionDir, "metadata.json")
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("read session metadata: %w", err)
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("unmarshal session metadata: %w", err)
	}

	return &session, nil
}
