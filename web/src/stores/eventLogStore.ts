import { create } from 'zustand';
import type { GraphEvent } from './types';

const MAX_EVENTS = 500;

export interface StoredGraphEvent extends GraphEvent {
  _id: number;
}

export interface EventLogState {
  /** Ring buffer of recent events, newest first. */
  events: StoredGraphEvent[];
  addEvent: (event: GraphEvent) => void;
  clear: () => void;
}

let eventCounter = 0;

export const useEventLogStore = create<EventLogState>((set) => ({
  events: [],

  addEvent: (event) =>
    set((state) => {
      const storedEvent: StoredGraphEvent = { ...event, _id: ++eventCounter };
      const next = [storedEvent, ...state.events];
      if (next.length > MAX_EVENTS) {
        next.length = MAX_EVENTS;
      }
      return { events: next };
    }),

  clear: () => set({ events: [] }),
}));
