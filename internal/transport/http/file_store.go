package http

import (
	"sync"
	"time"
)

// fileStore — временное хранилище сгенерированных файлов в памяти.
// SSE отправляет только метаданные, клиент забирает файлы через GET endpoint.
// Записи автоматически удаляются через TTL.
type fileStore struct {
	mu      sync.RWMutex
	entries map[string]*fileEntry
}

type fileEntry struct {
	Files      map[string]string // filename → content
	Complete   bool              // true when generation finished (all files stored)
	LastStatus string            // latest orchestrator status message for polling clients
	CreatedAt  time.Time
	UpdatedAt  time.Time // refreshed on every write; TTL counts from here
}

// fileTTL — время жизни записи после завершения (30 минут).
// Incomplete entries are never expired by TTL — only completed ones count down.
const fileTTL = 30 * time.Minute

var globalFileStore = newFileStore()

func newFileStore() *fileStore {
	fs := &fileStore{
		entries: make(map[string]*fileEntry),
	}
	// Background cleanup goroutine
	go fs.cleanup()
	return fs
}

// Store saves files for a session. Merges into existing entry if present.
func (fs *fileStore) Store(sessionID string, files map[string]string) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	entry, ok := fs.entries[sessionID]
	if !ok {
		entry = &fileEntry{
			Files:     make(map[string]string),
			CreatedAt: time.Now(),
		}
		fs.entries[sessionID] = entry
	}
	entry.UpdatedAt = time.Now() // refresh TTL on every write
	// Merge: final result overwrites individual file entries
	for k, v := range files {
		entry.Files[k] = v
	}
}

// Append adds a single file to an existing session entry (or creates new).
func (fs *fileStore) Append(sessionID, filename, content string) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	entry, ok := fs.entries[sessionID]
	if !ok {
		entry = &fileEntry{
			Files:     make(map[string]string),
			CreatedAt: time.Now(),
		}
		fs.entries[sessionID] = entry
	}
	entry.UpdatedAt = time.Now() // refresh TTL on every write
	entry.Files[filename] = content
}

// Get returns stored files for a session, or nil if not found.
// Does NOT check TTL — cleanup goroutine handles expiry.
func (fs *fileStore) Get(sessionID string) map[string]string {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	entry, ok := fs.entries[sessionID]
	if !ok {
		return nil
	}
	return entry.Files
}

// MarkComplete marks a session's files as complete (generation finished).
// Creates an empty entry if none exists (ensures complete=true even on error).
// Resets UpdatedAt so TTL countdown starts from completion moment.
func (fs *fileStore) MarkComplete(sessionID string) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	entry, ok := fs.entries[sessionID]
	if !ok {
		entry = &fileEntry{
			Files:     make(map[string]string),
			CreatedAt: time.Now(),
		}
		fs.entries[sessionID] = entry
	}
	entry.Complete = true
	entry.UpdatedAt = time.Now() // TTL counts from completion
}

// IsComplete returns true if the session's files are marked as complete.
func (fs *fileStore) IsComplete(sessionID string) bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	entry, ok := fs.entries[sessionID]
	if !ok {
		return false
	}
	return entry.Complete
}

// FileCount returns number of stored files for a session.
func (fs *fileStore) FileCount(sessionID string) int {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	entry, ok := fs.entries[sessionID]
	if !ok {
		return 0
	}
	return len(entry.Files)
}

// UpdateStatus updates the latest orchestrator status message for a session.
func (fs *fileStore) UpdateStatus(sessionID, status string) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	entry, ok := fs.entries[sessionID]
	if !ok {
		entry = &fileEntry{
			Files:     make(map[string]string),
			CreatedAt: time.Now(),
		}
		fs.entries[sessionID] = entry
	}
	entry.LastStatus = status
	entry.UpdatedAt = time.Now()
}

// GetStatus returns the latest orchestrator status message for a session.
func (fs *fileStore) GetStatus(sessionID string) string {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	entry, ok := fs.entries[sessionID]
	if !ok {
		return ""
	}
	return entry.LastStatus
}

// Delete removes files for a session.
func (fs *fileStore) Delete(sessionID string) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	delete(fs.entries, sessionID)
}

// cleanup periodically removes expired entries.
// Only removes COMPLETED entries older than TTL (based on UpdatedAt).
// Incomplete entries are never removed — they represent running generations.
func (fs *fileStore) cleanup() {
	ticker := time.NewTicker(2 * time.Minute)
	for range ticker.C {
		fs.mu.Lock()
		now := time.Now()
		for id, entry := range fs.entries {
			// Never expire incomplete entries (generation still running)
			if !entry.Complete {
				continue
			}
			// Use UpdatedAt for TTL (set on MarkComplete)
			age := now.Sub(entry.UpdatedAt)
			if age > fileTTL {
				delete(fs.entries, id)
			}
		}
		fs.mu.Unlock()
	}
}
