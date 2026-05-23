import { create } from 'zustand';
import type { TaskStatus } from './types';

export interface StatusState {
  task: TaskStatus | null;
  loading: boolean;
  error: string | null;
  poll: (baseUrl: string) => Promise<void>;
}

export const useStatusStore = create<StatusState>((set) => ({
  task: null,
  loading: false,
  error: null,

  poll: async (baseUrl: string) => {
    set({ loading: true, error: null });
    try {
      const res = await fetch(`${baseUrl}/api/status`);
      if (!res.ok) {
        set({ loading: false, error: `HTTP ${res.status}` });
        return;
      }
      const task = (await res.json()) as TaskStatus;
      if (task.id || task.status) {
        set({ task, loading: false });
      } else {
        set({ task: null, loading: false });
      }
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      console.error('[StatusStore] poll failed:', msg, e);
      set({ loading: false, error: msg });
    }
  },
}));
