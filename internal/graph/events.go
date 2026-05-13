package graph

import (
	"time"
)

// EventType represents the type of a graph event.
type EventType string

// Standard event types emitted by the graph engine.
const (
	EventScanStarted   EventType = "SCAN_STARTED"
	EventNodeUpserted  EventType = "NODE_UPSERTED"
	EventEdgeCreated   EventType = "EDGE_CREATED"
	EventScanCompleted EventType = "SCAN_COMPLETED"
)

// GraphEvent represents an event emitted during graph lifecycle operations.
type GraphEvent struct { //nolint:revive // Intentional name: graph.GraphEvent avoids ambiguity
	Type    EventType   `json:"type"`
	Payload interface{} `json:"payload"` // Node or Edge
	Time    time.Time   `json:"timestamp"`
}

// Observer receives notifications about graph events.
type Observer interface {
	// Notify is called when a graph event occurs.
	Notify(event GraphEvent)
}
