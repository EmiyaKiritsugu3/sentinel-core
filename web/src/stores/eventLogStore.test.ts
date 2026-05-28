import { describe, it, expect, beforeEach } from 'vitest';
import { useEventLogStore } from './eventLogStore';
import type { GraphEvent } from './types';

describe('eventLogStore', () => {
  beforeEach(() => {
    useEventLogStore.getState().clear();
  });

  it('adds an event and generates a stable _id', () => {
    const event: GraphEvent = { type: 'TEST', payload: {}, timestamp: '2026-05-24T10:00:00Z' };
    useEventLogStore.getState().addEvent(event);

    const events = useEventLogStore.getState().events;
    expect(events.length).toBe(1);
    expect(events[0]._id).toBeDefined();
    expect(events[0].type).toBe('TEST');
  });

  it('prepends events and respects max events', () => {
    for (let i = 0; i < 505; i++) {
      useEventLogStore.getState().addEvent({ type: `TEST_${i}`, payload: {}, timestamp: '2026-05-24T10:00:00Z' });
    }

    const events = useEventLogStore.getState().events;
    expect(events.length).toBe(500);
    // The last added event is TEST_504, which should be at the front
    expect(events[0].type).toBe('TEST_504');
  });

  it('clears events', () => {
    useEventLogStore.getState().addEvent({ type: 'TEST', payload: {}, timestamp: '2026-05-24T10:00:00Z' });
    expect(useEventLogStore.getState().events.length).toBe(1);

    useEventLogStore.getState().clear();
    expect(useEventLogStore.getState().events.length).toBe(0);
  });
});
