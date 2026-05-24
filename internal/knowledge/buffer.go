package knowledge

import (
	"sort"
	"sync"
	"time"
)

type EventType string

const (
	EventDecision   EventType = "decision"
	EventError      EventType = "error"
	EventPattern    EventType = "pattern"
	EventFileChange EventType = "file_change"
	EventCommand    EventType = "command"
	EventMetric     EventType = "metric"
)

type SessionEvent struct {
	Timestamp time.Time
	Type      EventType
	Domain    string
	Summary   string
	Detail    string
	File      string
	Tags      []string
}

type EventBuffer struct {
	mu     sync.RWMutex
	events []SessionEvent
	max    int
	head   int
	size   int
}

func NewEventBuffer(maxSize int) *EventBuffer {
	if maxSize < 1 {
		maxSize = 1000
	}
	return &EventBuffer{
		events: make([]SessionEvent, maxSize),
		max:    maxSize,
	}
}

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

func (b *EventBuffer) ByDomain(domain string) []SessionEvent {
	return b.filter(func(e SessionEvent) bool {
		return e.Domain == domain
	})
}

func (b *EventBuffer) ByType(typ EventType) []SessionEvent {
	return b.filter(func(e SessionEvent) bool {
		return e.Type == typ
	})
}

func (b *EventBuffer) Patterns() []SessionEvent {
	return b.ByType(EventPattern)
}

func (b *EventBuffer) Decisions() []SessionEvent {
	return b.ByType(EventDecision)
}

func (b *EventBuffer) Errors() []SessionEvent {
	return b.ByType(EventError)
}

func (b *EventBuffer) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.size
}

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
