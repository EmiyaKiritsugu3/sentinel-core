import { create } from 'zustand';

/** Node type labels as stored in the SQLite nodes.type column. */
export type NodeType = 'function' | 'struct' | 'interface' | 'unresolved_import' | 'file';

export interface FilterState {
  /** Which node types are currently visible. Empty set = show all. */
  enabledTypes: Set<NodeType>;
  /** Search text for node name / file path filtering (instant, no debounce). */
  searchText: string;
  /** Selected package prefix for file_path filtering. null = no package filter. */
  selectedPackage: string | null;
}

export interface FilterActions {
  toggleType: (t: NodeType) => void;
  setSearchText: (text: string) => void;
  setSelectedPackage: (pkg: string | null) => void;
  reset: () => void;
}

export const useFilterStore = create<FilterState & FilterActions>((set) => ({
  enabledTypes: new Set(),
  searchText: '',
  selectedPackage: null,

  toggleType: (t) =>
    set((state) => {
      const next = new Set(state.enabledTypes);
      if (next.has(t)) {
        next.delete(t);
      } else {
        next.add(t);
      }
      return { enabledTypes: next };
    }),

  setSearchText: (text) => set({ searchText: text }),

  setSelectedPackage: (pkg) => set({ selectedPackage: pkg }),

  reset: () =>
    set({
      enabledTypes: new Set(),
      searchText: '',
      selectedPackage: null,
    }),
}));
