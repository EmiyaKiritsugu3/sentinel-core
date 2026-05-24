package knowledge

import (
	"sort"
	"sync"
	"time"
)

// EventType classifies a session event for debrief categorization.
type EventType string

const (
	// EventDecision records an architectural or implementation decision.
	EventDecision EventType = "decision"
	// EventError records a captured error or warning.
	EventError EventType = "error"
	// EventPattern records a detected architectural pattern.
	EventPattern EventType = "pattern"
	// EventFileChange records a file modification event.
	EventFileChange EventType = "file_change"
	// EventCommand records a shell command executed during the session.
	EventCommand EventType = "command"
	// EventMetric records a numeric measurement (e.g., token count, latency).
	EventMetric EventType = "metric"
)

// SessionEvent represents a single captured event during a sentinel session.
type SessionEvent struct {
	Timestamp time.Time
	Type      EventType
	Domain    string
	Summary   string
	Detail    string
	File      string
	Tags      []string
}

// EventBuffer is a thread-safe ring buffer that collects session events for later debrief generation.
type EventBuffer struct {
	mu     sync.RWMutex
	events []SessionEvent
	max    int
	head   int
	size   int
}

// NewEventBuffer creates a ring buffer for SessionEvent values with the specified maximum capacity.
// If maxSize is less than 1, the buffer is created with a default capacity of 1000.
func NewEventBuffer(maxSize int) *EventBuffer {
	if maxSize < 1 {
		maxSize = 1000
	}
	return &EventBuffer{
		events: make([]SessionEvent, maxSize),
		max:    maxSize,
	}
}

// Record appends an event to the buffer. Thread-safe. Oldest events are overwritten when the buffer is full.
func (b *EventBuffer) Record(event SessionEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	b.events[b.head] = event
	b.head = (b.head + 1) % b.max
	if b.size < b.max {
		b.size++
	}
}

// Snapshot returns all events in chronological order (oldest first).
func (b *EventBuffer) Snapshot() []SessionEvent {
	b.mu.RLock()
	defer b.mu.RUnlock()
	result := make([]SessionEvent, b.size)
	if b.size == 0 {
		return result
	}
	start := (b.head - b.size + b.max) % b.max
	for i := 0; i < b.size; i++ {
		idx := (start + i) % b.max
		result[i] = b.events[idx]
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.Before(result[j].Timestamp)
	})
	return result
}

// ByDomain returns all events matching the given domain.
func (b *EventBuffer) ByDomain(domain string) []SessionEvent {
	return b.filter(func(e SessionEvent) bool {
		return e.Domain == domain
	})
}

// ByType returns all events matching the given event type.
func (b *EventBuffer) ByType(typ EventType) []SessionEvent {
	return b.filter(func(e SessionEvent) bool {
		return e.Type == typ
	})
}

// Patterns returns all events of type EventPattern.
func (b *EventBuffer) Patterns() []SessionEvent {
	return b.ByType(EventPattern)
}

// Decisions returns all events of type EventDecision.
func (b *EventBuffer) Decisions() []SessionEvent {
	return b.ByType(EventDecision)
}

// Errors returns all events of type EventError.
func (b *EventBuffer) Errors() []SessionEvent {
	return b.ByType(EventError)
}

// Len returns the current number of events in the buffer.
func (b *EventBuffer) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.size
}

// GlobalBuffer is the process-wide singleton event buffer. All sentinel subsystems record events here during a session.
var GlobalBuffer = NewEventBuffer(1000)

func (b *EventBuffer) filter(pred func(SessionEvent) bool) []SessionEvent {
	b.mu.RLock()
	defer b.mu.RUnlock()
	var result []SessionEvent
	if b.size == 0 {
		return result
	}
	start := (b.head - b.size + b.max) % b.max
	for i := 0; i < b.size; i++ {
		idx := (start + i) % b.max
		if pred(b.events[idx]) {
			result = append(result, b.events[idx])
		}
	}
	return result
}
