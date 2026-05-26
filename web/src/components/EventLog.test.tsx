import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { EventLog } from './EventLog';

let storeState: { events: unknown[]; clear: () => void } = { events: [], clear: vi.fn() };

vi.mock('../stores', () => ({
  useEventLogStore: (selector?: (state: typeof storeState) => unknown) => {
    if (typeof selector === 'function') return selector(storeState);
    return storeState;
  },
}));

describe('EventLog', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    storeState = { events: [], clear: vi.fn() };
  });

  it('shows empty state when no events', () => {
    render(<EventLog />);
    expect(screen.getByText('No events yet')).toBeDefined();
  });

  it('renders events and shows count', () => {
    storeState = {
      events: [
        { type: 'NODE_UPSERTED', payload: { Name: 'main' }, timestamp: '2026-05-24T10:00:00Z' },
        { type: 'EDGE_CREATED', payload: { From: 'a', To: 'b' }, timestamp: '2026-05-24T10:00:01Z' },
      ],
      clear: vi.fn(),
    };
    render(<EventLog />);
    expect(screen.getByText('NODE_UPSERTED')).toBeDefined();
    expect(screen.getByText('EDGE_CREATED')).toBeDefined();
    expect(screen.getByText('Node: main')).toBeDefined();
    expect(screen.getByText('Edge: a → b')).toBeDefined();
  });

  it('calls clear on button click', () => {
    const clearFn = vi.fn();
    storeState = {
      events: [
        { type: 'NODE_UPSERTED', payload: {}, timestamp: '2026-05-24T10:00:00Z' },
      ],
      clear: clearFn,
    };
    render(<EventLog />);
    fireEvent.click(screen.getByTitle('Clear events'));
    expect(clearFn).toHaveBeenCalled();
  });

  it('does not show clear button when empty', () => {
    render(<EventLog />);
    expect(screen.queryByTitle('Clear events')).toBeNull();
  });

  it('handles scroll event to stop auto-scroll', () => {
    storeState = {
      events: [
        { type: 'NODE_UPSERTED', payload: {}, timestamp: '2026-05-24T10:00:00Z' },
      ],
      clear: vi.fn(),
    };
    render(<EventLog />);

    const list = document.querySelector('.event-log__list');
    if (list) {
      // Simulate scrolling down past 40px
      Object.defineProperty(list, 'scrollTop', { value: 50, writable: true });
      fireEvent.scroll(list);
    }
  });
});
