import { create } from 'zustand';
import type { GraphEvent } from './types';

const MAX_EVENTS = 500;

export interface EventLogState {
  /** Ring buffer of recent events, newest first. */
  events: GraphEvent[];
  addEvent: (event: GraphEvent) => void;
  clear: () => void;
}

export const useEventLogStore = create<EventLogState>((set) => ({
  events: [],

  addEvent: (event) =>
    set((state) => {
      const next = [event, ...state.events];
      if (next.length > MAX_EVENTS) {
        next.length = MAX_EVENTS;
      }
      return { events: next };
    }),

  clear: () => set({ events: [] }),
}));
