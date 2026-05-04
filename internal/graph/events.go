package graph

import (
	"time"
)

type EventType string

const (
	EventScanStarted   EventType = "SCAN_STARTED"
	EventNodeUpserted  EventType = "NODE_UPSERTED"
	EventEdgeCreated   EventType = "EDGE_CREATED"
	EventScanCompleted EventType = "SCAN_COMPLETED"
)

type GraphEvent struct {
	Type    EventType   `json:"type"`
	Payload interface{} `json:"payload"` // Node or Edge
	Time    time.Time   `json:"timestamp"`
}

type Observer interface {
	Notify(event GraphEvent)
}
