package webapi

import (
	"strings"
	"sync"
	"time"
)

type LogEntry struct {
	Time string `json:"time"`
	Text string `json:"text"`
}

type LogBuffer struct {
	mu      sync.RWMutex
	entries []LogEntry
	maxSize int
}

func NewLogBuffer(maxSize int) *LogBuffer {
	return &LogBuffer{
		entries: make([]LogEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

func (lb *LogBuffer) Append(text string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	entry := LogEntry{
		Time: time.Now().Format("15:04:05"),
		Text: text,
	}

	if len(lb.entries) >= lb.maxSize {
		lb.entries = lb.entries[1:]
	}
	lb.entries = append(lb.entries, entry)
}

func (lb *LogBuffer) GetEntries(filter string) []LogEntry {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	if filter == "" {
		result := make([]LogEntry, len(lb.entries))
		copy(result, lb.entries)
		return result
	}

	result := make([]LogEntry, 0)
	lowerFilter := strings.ToLower(filter)
	for _, entry := range lb.entries {
		if strings.Contains(strings.ToLower(entry.Text), lowerFilter) {
			result = append(result, entry)
		}
	}
	return result
}

func (lb *LogBuffer) GetAll() []LogEntry {
	return lb.GetEntries("")
}

func (lb *LogBuffer) Clear() {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.entries = lb.entries[:0]
}
