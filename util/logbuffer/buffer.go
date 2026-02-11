package logbuffer

import (
	"sync"
	"time"
)

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
}

type Buffer struct {
	entries []LogEntry
	size    int
	mu      sync.Mutex
}

var DefaultBuffer = NewBuffer(1000)

func NewBuffer(size int) *Buffer {
	return &Buffer{
		entries: make([]LogEntry, 0, size),
		size:    size,
	}
}

func (b *Buffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	entry := LogEntry{
		Timestamp: time.Now(),
		Message:   string(p),
	}

	if len(b.entries) >= b.size {
		b.entries = b.entries[1:]
	}
	b.entries = append(b.entries, entry)

	return len(p), nil
}

func (b *Buffer) GetEntries() []LogEntry {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Return a copy to avoid race conditions
	copyEntries := make([]LogEntry, len(b.entries))
	copy(copyEntries, b.entries)
	return copyEntries
}
