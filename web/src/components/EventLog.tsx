import { useEffect, useRef } from 'react';
import { useEventLogStore } from '../stores';
import type { GraphEvent } from '../stores/types';
import './EventLog.css';

/**
 * Renders a scrollable event log of GraphEvents received via WebSocket.
 * Events are displayed newest-first with type, timestamp, and payload summary.
 * Auto-scrolls to top on new events unless user has scrolled away.
 */
export function EventLog() {
  const events = useEventLogStore(s => s.events);
  const clear  = useEventLogStore(s => s.clear);
  const listRef = useRef<HTMLDivElement>(null);
  const userScrolled = useRef(false);

  // Auto-scroll to top on new events (newest first)
  useEffect(() => {
    if (!userScrolled.current && listRef.current) {
      listRef.current.scrollTop = 0;
    }
  }, [events.length]);

  const handleScroll = () => {
    if (!listRef.current) return;
    // If user scrolls down more than 40px from top, stop auto-scroll
    userScrolled.current = listRef.current.scrollTop > 40;
  };

  return (
    <div className="event-log">
      <div className="event-log__header">
        <span className="event-log__title">Event Log</span>
        <span className="event-log__count">{events.length}</span>
        {events.length > 0 && (
          <button className="event-log__clear" aria-label="Clear event log" onClick={clear} title="Clear events">
            ✕
          </button>
        )}
      </div>
      <div
        className="event-log__list"
        ref={listRef}
        onScroll={handleScroll}
      >
        {events.length === 0 ? (
          <div className="event-log__empty">No events yet</div>
        ) : (
          events.map((event, i) => (
            <EventRow key={`${event.timestamp}-${i}`} event={event} />
          ))
        )}
      </div>
    </div>
  );
}

function EventRow({ event }: { event: GraphEvent }) {
  const time = event.timestamp
    ? new Date(event.timestamp).toLocaleTimeString()
    : '';
  const summary = summarizePayload(event);

  return (
    <div className="event-log__row" data-event-type={event.type}>
      <div className="event-log__row-header">
        <span className={`event-log__type event-log__type--${event.type.toLowerCase()}`}>
          {event.type}
        </span>
        <span className="event-log__time">{time}</span>
      </div>
      {summary && (
        <div className="event-log__summary">{summary}</div>
      )}
    </div>
  );
}

/** Produces a concise one-line summary from the event payload. */
function summarizePayload(event: GraphEvent): string {
  const p = event.payload;
  if (!p || typeof p !== 'object') return '';

  switch (event.type) {
    case 'NODE_UPSERTED':
      return (p as Record<string, unknown>).Name
        ? `Node: ${(p as Record<string, unknown>).Name}`
        : `Node: ${(p as Record<string, unknown>).ID ?? '?'}`;
    case 'EDGE_CREATED':
      return `Edge: ${(p as Record<string, unknown>).From ?? '?'} → ${(p as Record<string, unknown>).To ?? '?'}`;
    default:
      return Object.keys(p as Record<string, unknown>).slice(0, 2).join(', ');
  }
}
