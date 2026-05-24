package knowledge

import (
	"sync"
	"testing"
	"time"
)

func TestNewEventBuffer_DefaultsTo1000(t *testing.T) {
	b := NewEventBuffer(0)
	if b.max != 1000 {
		t.Errorf("expected max=1000, got %d", b.max)
	}
	b2 := NewEventBuffer(-5)
	if b2.max != 1000 {
		t.Errorf("expected max=1000, got %d", b2.max)
	}
}

func TestRecordAndSnapshot_Basic(t *testing.T) {
	b := NewEventBuffer(10)
	e1 := SessionEvent{Type: EventDecision, Domain: "auth", Summary: "decided something"}
	e2 := SessionEvent{Type: EventError, Domain: "network", Summary: "connection refused"}

	b.Record(e1)
	b.Record(e2)

	snap := b.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 events, got %d", len(snap))
	}
	if snap[0].Type != EventDecision || snap[1].Type != EventError {
		t.Errorf("unexpected order or types in snapshot")
	}
}

func TestSnapshot_ChronologicalOrder(t *testing.T) {
	b := NewEventBuffer(10)
	now := time.Now()
	b.Record(SessionEvent{Timestamp: now.Add(2 * time.Second), Type: EventDecision})
	b.Record(SessionEvent{Timestamp: now, Type: EventError})
	b.Record(SessionEvent{Timestamp: now.Add(1 * time.Second), Type: EventPattern})

	snap := b.Snapshot()
	if len(snap) != 3 {
		t.Fatalf("expected 3 events, got %d", len(snap))
	}
	for i := 1; i < len(snap); i++ {
		if snap[i-1].Timestamp.After(snap[i].Timestamp) {
			t.Errorf("snapshot not sorted chronologically at index %d", i)
		}
	}
}

func TestRingBuffer_Wraparound(t *testing.T) {
	b := NewEventBuffer(3)
	b.Record(SessionEvent{Type: EventDecision, Summary: "e1"})
	b.Record(SessionEvent{Type: EventError, Summary: "e2"})
	b.Record(SessionEvent{Type: EventPattern, Summary: "e3"})
	b.Record(SessionEvent{Type: EventCommand, Summary: "e4"})

	if b.Len() != 3 {
		t.Fatalf("expected Len=3 after wraparound, got %d", b.Len())
	}

	snap := b.Snapshot()
	if len(snap) != 3 {
		t.Fatalf("expected 3 events in snapshot, got %d", len(snap))
	}
	if snap[0].Summary != "e2" {
		t.Errorf("expected first event e2, got %s", snap[0].Summary)
	}
	if snap[1].Summary != "e3" {
		t.Errorf("expected second event e3, got %s", snap[1].Summary)
	}
	if snap[2].Summary != "e4" {
		t.Errorf("expected third event e4, got %s", snap[2].Summary)
	}
}

func TestByDomain(t *testing.T) {
	b := NewEventBuffer(10)
	b.Record(SessionEvent{Domain: "auth", Type: EventDecision})
	b.Record(SessionEvent{Domain: "network", Type: EventError})
	b.Record(SessionEvent{Domain: "auth", Type: EventPattern})

	authEvents := b.ByDomain("auth")
	if len(authEvents) != 2 {
		t.Fatalf("expected 2 auth events, got %d", len(authEvents))
	}
	for _, e := range authEvents {
		if e.Domain != "auth" {
			t.Errorf("unexpected domain %s in auth filter", e.Domain)
		}
	}

	netEvents := b.ByDomain("network")
	if len(netEvents) != 1 {
		t.Fatalf("expected 1 network event, got %d", len(netEvents))
	}

	emptyEvents := b.ByDomain("nonexistent")
	if len(emptyEvents) != 0 {
		t.Errorf("expected 0 events for nonexistent domain, got %d", len(emptyEvents))
	}
}

func TestByType(t *testing.T) {
	b := NewEventBuffer(10)
	b.Record(SessionEvent{Type: EventDecision, Domain: "a"})
	b.Record(SessionEvent{Type: EventError, Domain: "b"})
	b.Record(SessionEvent{Type: EventDecision, Domain: "c"})

	decisions := b.ByType(EventDecision)
	if len(decisions) != 2 {
		t.Fatalf("expected 2 decisions, got %d", len(decisions))
	}
	for _, e := range decisions {
		if e.Type != EventDecision {
			t.Errorf("unexpected type %s in decision filter", e.Type)
		}
	}

	errors := b.ByType(EventError)
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
}

func TestPatternsDecisionsErrors_Shortcuts(t *testing.T) {
	b := NewEventBuffer(10)
	b.Record(SessionEvent{Type: EventPattern, Summary: "p1"})
	b.Record(SessionEvent{Type: EventDecision, Summary: "d1"})
	b.Record(SessionEvent{Type: EventError, Summary: "e1"})
	b.Record(SessionEvent{Type: EventPattern, Summary: "p2"})

	if len(b.Patterns()) != 2 {
		t.Errorf("expected 2 patterns, got %d", len(b.Patterns()))
	}
	if len(b.Decisions()) != 1 {
		t.Errorf("expected 1 decision, got %d", len(b.Decisions()))
	}
	if len(b.Errors()) != 1 {
		t.Errorf("expected 1 error, got %d", len(b.Errors()))
	}
}

func TestEmptyBuffer(t *testing.T) {
	b := NewEventBuffer(10)

	if b.Len() != 0 {
		t.Errorf("expected Len=0, got %d", b.Len())
	}

	snap := b.Snapshot()
	if len(snap) != 0 {
		t.Errorf("expected empty snapshot, got %d events", len(snap))
	}

	byDomain := b.ByDomain("test")
	if len(byDomain) != 0 {
		t.Errorf("expected empty ByDomain, got %d events", len(byDomain))
	}

	byType := b.ByType(EventDecision)
	if len(byType) != 0 {
		t.Errorf("expected empty ByType, got %d events", len(byType))
	}

	if len(b.Patterns()) != 0 {
		t.Errorf("expected empty Patterns, got %d events", len(b.Patterns()))
	}
}

func TestConcurrentRecordAndRead(t *testing.T) {
	b := NewEventBuffer(500)
	var wg sync.WaitGroup

	numWriters := 10
	numReaders := 5
	eventsPerWriter := 100

	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < eventsPerWriter; j++ {
				b.Record(SessionEvent{
					Type:    EventMetric,
					Domain:  "perf",
					Summary: "writer event",
				})
			}
		}(i)
	}

	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				_ = b.Snapshot()
				_ = b.ByDomain("perf")
				_ = b.Len()
			}
		}(i)
	}

	wg.Wait()

	len := b.Len()
	if len < 1 || len > 500 {
		t.Errorf("unexpected buffer length %d", len)
	}
}

func TestEventBuffer_TagsImmutability(t *testing.T) {
	buf := NewEventBuffer(10)
	tags := []string{"original"}
	buf.Record(SessionEvent{Type: EventDecision, Summary: "test", Tags: tags})
	tags[0] = "mutated"
	snap := buf.Snapshot()
	if snap[0].Tags[0] != "original" {
		t.Errorf("Tags should be immutable: got %q", snap[0].Tags[0])
	}
}
