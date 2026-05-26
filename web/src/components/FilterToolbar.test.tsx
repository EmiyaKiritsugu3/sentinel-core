import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { FilterToolbar } from './FilterToolbar';

vi.mock('../stores', () => {
  const state: {
    enabledTypes: Set<string>;
    searchText: string;
    selectedPackage: string | null;
    toggleType: () => void;
    setSearchText: () => void;
    setSelectedPackage: () => void;
    reset: () => void;
  } = {
    enabledTypes: new Set() as Set<string>,
    searchText: '',
    selectedPackage: null,
    toggleType: vi.fn(),
    setSearchText: vi.fn(),
    setSelectedPackage: vi.fn(),
    reset: vi.fn(),
  };
  return {
    useFilterStore: (selector?: (s: typeof state) => unknown) => {
      if (typeof selector === 'function') return selector(state);
      return state;
    },
  };
});

import { useFilterStore } from '../stores';

describe('FilterToolbar', () => {
  it('renders all 5 type checkboxes', () => {
    render(<FilterToolbar packages={['internal', 'pkg']} />);
    expect(screen.getByText('function')).toBeDefined();
    expect(screen.getByText('struct')).toBeDefined();
  });

  it('calls toggleType on checkbox change', () => {
    const store = useFilterStore();
    render(<FilterToolbar packages={[]} />);
    const cb = screen.getByText('function').closest('label')?.querySelector('input');
    fireEvent.click(cb!);
    expect(store.toggleType).toHaveBeenCalledWith('function');
  });

  it('calls setSearchText on input change', () => {
    const store = useFilterStore();
    render(<FilterToolbar packages={[]} />);
    const input = screen.getByPlaceholderText('Search nodes...');
    fireEvent.change(input, { target: { value: 'auth' } });
    expect(store.setSearchText).toHaveBeenCalledWith('auth');
  });

  it('calls reset on button click', () => {
    const store = useFilterStore();
    render(<FilterToolbar packages={[]} />);
    fireEvent.click(screen.getByText('Reset'));
    expect(store.reset).toHaveBeenCalled();
  });
});
