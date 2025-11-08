package storage

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

const (
	binaryMagicNumber = uint32(0x474F5452) // "GOTR" (Go Trace)
	binaryVersion     = uint32(1)
	eventSize         = 64 // size of Event struct in bytes
)

// BinaryStore implements EventStore using a binary format
type BinaryStore struct {
	file       *os.File
	session    *Session
	mu         sync.RWMutex
	eventCount int64
	baseDir    string
}

// NewBinaryStore creates a new binary event store
func NewBinaryStore(baseDir string, session *Session) (*BinaryStore, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("create base directory: %w", err)
	}

	sessionDir := filepath.Join(baseDir, session.ID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return nil, fmt.Errorf("create session directory: %w", err)
	}

	filePath := filepath.Join(sessionDir, "events.bin")
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("open binary file: %w", err)
	}

	store := &BinaryStore{
		file:    file,
		session: session,
		baseDir: baseDir,
	}

	// Write header if file is empty
	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("stat file: %w", err)
	}

	if stat.Size() == 0 {
		if err := store.writeHeader(); err != nil {
			file.Close()
			return nil, fmt.Errorf("write header: %w", err)
		}
	}

	return store, nil
}

// OpenBinaryStore opens an existing binary store for reading
func OpenBinaryStore(baseDir string, sessionID string) (*BinaryStore, error) {
	sessionDir := filepath.Join(baseDir, sessionID)
	filePath := filepath.Join(sessionDir, "events.bin")

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open binary file: %w", err)
	}

	store := &BinaryStore{
		file:    file,
		baseDir: baseDir,
	}

	// Read and validate header
	if err := store.readHeader(); err != nil {
		file.Close()
		return nil, fmt.Errorf("read header: %w", err)
	}

	// Load session metadata
	session, err := loadSessionMetadata(sessionDir)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("load session metadata: %w", err)
	}
	store.session = session

	return store, nil
}

func (s *BinaryStore) writeHeader() error {
	if err := binary.Write(s.file, binary.LittleEndian, binaryMagicNumber); err != nil {
		return err
	}
	if err := binary.Write(s.file, binary.LittleEndian, binaryVersion); err != nil {
		return err
	}
	return nil
}

func (s *BinaryStore) readHeader() error {
	var magic, version uint32
	if err := binary.Read(s.file, binary.LittleEndian, &magic); err != nil {
		return fmt.Errorf("read magic: %w", err)
	}
	if magic != binaryMagicNumber {
		return fmt.Errorf("invalid magic number: %x", magic)
	}
	if err := binary.Read(s.file, binary.LittleEndian, &version); err != nil {
		return fmt.Errorf("read version: %w", err)
	}
	if version != binaryVersion {
		return fmt.Errorf("unsupported version: %d", version)
	}
	return nil
}

func (s *BinaryStore) WriteEvent(event *Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := binary.Write(s.file, binary.LittleEndian, event); err != nil {
		return fmt.Errorf("write event: %w", err)
	}

	s.eventCount++
	return nil
}

func (s *BinaryStore) WriteBatch(events []*Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, event := range events {
		if err := binary.Write(s.file, binary.LittleEndian, event); err != nil {
			return fmt.Errorf("write event: %w", err)
		}
		s.eventCount++
	}

	return nil
}

func (s *BinaryStore) ReadEvents(ctx context.Context, filter *EventFilter) ([]*Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Seek to beginning after header
	if _, err := s.file.Seek(8, io.SeekStart); err != nil {
		return nil, fmt.Errorf("seek to start: %w", err)
	}

	var events []*Event
	count := 0
	skipped := 0

	for {
		select {
		case <-ctx.Done():
			return events, ctx.Err()
		default:
		}

		var event Event
		if err := binary.Read(s.file, binary.LittleEndian, &event); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("read event: %w", err)
		}

		// Apply filters
		if filter != nil {
			if filter.Goroutine != nil && event.Goroutine != *filter.Goroutine {
				continue
			}
			if filter.EventType != nil && event.EventType != *filter.EventType {
				continue
			}
			if filter.StartTime != nil && event.Timestamp < *filter.StartTime {
				continue
			}
			if filter.EndTime != nil && event.Timestamp > *filter.EndTime {
				continue
			}
			if filter.Offset > 0 && skipped < filter.Offset {
				skipped++
				continue
			}
		}

		events = append(events, &event)
		count++

		if filter != nil && filter.Limit > 0 && count >= filter.Limit {
			break
		}
	}

	return events, nil
}

func (s *BinaryStore) GetGoroutines(ctx context.Context) ([]uint32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, err := s.file.Seek(8, io.SeekStart); err != nil {
		return nil, fmt.Errorf("seek to start: %w", err)
	}

	goroutineMap := make(map[uint32]bool)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		var event Event
		if err := binary.Read(s.file, binary.LittleEndian, &event); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("read event: %w", err)
		}

		goroutineMap[event.Goroutine] = true
	}

	goroutines := make([]uint32, 0, len(goroutineMap))
	for gid := range goroutineMap {
		goroutines = append(goroutines, gid)
	}

	return goroutines, nil
}

func (s *BinaryStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.file != nil {
		return s.file.Close()
	}
	return nil
}

func (s *BinaryStore) GetSession() *Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.session
}

func (s *BinaryStore) UpdateSession(session *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.session = session
	sessionDir := filepath.Join(s.baseDir, session.ID)
	return saveSessionMetadata(sessionDir, session)
}
