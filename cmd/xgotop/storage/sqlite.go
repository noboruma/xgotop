package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStore implements EventStore using SQLite
type SQLiteStore struct {
	db         *sql.DB
	session    *Session
	mu         sync.RWMutex
	eventCount int64
	baseDir    string
}

const schema = `
CREATE TABLE IF NOT EXISTS events (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	timestamp INTEGER NOT NULL,
	event_type INTEGER NOT NULL,
	goroutine INTEGER NOT NULL,
	parent_goroutine INTEGER NOT NULL,
	attr0 INTEGER NOT NULL,
	attr1 INTEGER NOT NULL,
	attr2 INTEGER NOT NULL,
	attr3 INTEGER NOT NULL,
	attr4 INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_goroutine ON events(goroutine);
CREATE INDEX IF NOT EXISTS idx_timestamp ON events(timestamp);
CREATE INDEX IF NOT EXISTS idx_event_type ON events(event_type);
`

// NewSQLiteStore creates a new SQLite event store
func NewSQLiteStore(baseDir string, session *Session) (*SQLiteStore, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("create base directory: %w", err)
	}

	sessionDir := filepath.Join(baseDir, session.ID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return nil, fmt.Errorf("create session directory: %w", err)
	}

	dbPath := filepath.Join(sessionDir, "events.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	// Create tables
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}

	store := &SQLiteStore{
		db:      db,
		session: session,
		baseDir: baseDir,
	}

	return store, nil
}

// OpenSQLiteStore opens an existing SQLite store for reading
func OpenSQLiteStore(baseDir string, sessionID string) (*SQLiteStore, error) {
	sessionDir := filepath.Join(baseDir, sessionID)
	dbPath := filepath.Join(sessionDir, "events.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	store := &SQLiteStore{
		db:      db,
		baseDir: baseDir,
	}

	// Load session metadata
	session, err := loadSessionMetadata(sessionDir)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("load session metadata: %w", err)
	}
	store.session = session

	return store, nil
}

func (s *SQLiteStore) WriteEvent(event *Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `INSERT INTO events (timestamp, event_type, goroutine, parent_goroutine, attr0, attr1, attr2, attr3, attr4)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query,
		event.Timestamp,
		event.EventType,
		event.Goroutine,
		event.ParentGoroutine,
		event.Attributes[0],
		event.Attributes[1],
		event.Attributes[2],
		event.Attributes[3],
		event.Attributes[4],
	)

	if err != nil {
		return fmt.Errorf("insert event: %w", err)
	}

	s.eventCount++
	return nil
}

func (s *SQLiteStore) WriteBatch(events []*Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO events (timestamp, event_type, goroutine, parent_goroutine, attr0, attr1, attr2, attr3, attr4)
							 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, event := range events {
		_, err := stmt.Exec(
			event.Timestamp,
			event.EventType,
			event.Goroutine,
			event.ParentGoroutine,
			event.Attributes[0],
			event.Attributes[1],
			event.Attributes[2],
			event.Attributes[3],
			event.Attributes[4],
		)
		if err != nil {
			return fmt.Errorf("insert event: %w", err)
		}
		s.eventCount++
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (s *SQLiteStore) ReadEvents(ctx context.Context, filter *EventFilter) ([]*Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := "SELECT timestamp, event_type, goroutine, parent_goroutine, attr0, attr1, attr2, attr3, attr4 FROM events WHERE 1=1"
	args := []interface{}{}

	if filter != nil {
		if filter.Goroutine != nil {
			query += " AND goroutine = ?"
			args = append(args, *filter.Goroutine)
		}
		if filter.EventType != nil {
			query += " AND event_type = ?"
			args = append(args, *filter.EventType)
		}
		if filter.StartTime != nil {
			query += " AND timestamp >= ?"
			args = append(args, *filter.StartTime)
		}
		if filter.EndTime != nil {
			query += " AND timestamp <= ?"
			args = append(args, *filter.EndTime)
		}
		if filter.Limit > 0 {
			query += " LIMIT ?"
			args = append(args, filter.Limit)
		}
		if filter.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	}

	query += " ORDER BY timestamp ASC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query events: %w", err)
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		var event Event
		err := rows.Scan(
			&event.Timestamp,
			&event.EventType,
			&event.Goroutine,
			&event.ParentGoroutine,
			&event.Attributes[0],
			&event.Attributes[1],
			&event.Attributes[2],
			&event.Attributes[3],
			&event.Attributes[4],
		)
		if err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return events, nil
}

func (s *SQLiteStore) GetGoroutines(ctx context.Context) ([]uint32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.QueryContext(ctx, "SELECT DISTINCT goroutine FROM events ORDER BY goroutine")
	if err != nil {
		return nil, fmt.Errorf("query goroutines: %w", err)
	}
	defer rows.Close()

	var goroutines []uint32
	for rows.Next() {
		var gid uint32
		if err := rows.Scan(&gid); err != nil {
			return nil, fmt.Errorf("scan goroutine: %w", err)
		}
		goroutines = append(goroutines, gid)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return goroutines, nil
}

func (s *SQLiteStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *SQLiteStore) GetSession() *Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.session
}

func (s *SQLiteStore) UpdateSession(session *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.session = session
	sessionDir := filepath.Join(s.baseDir, session.ID)
	return saveSessionMetadata(sessionDir, session)
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
