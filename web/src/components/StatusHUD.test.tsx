import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { StatusHUD } from './StatusHUD';

// Mock Zustand store
const mockPoll = vi.fn();
type MockState = {
  task: Record<string, unknown> | null;
  loading: boolean;
  error: string | null;
  poll: typeof mockPoll;
};
let mockState: MockState = { task: null, loading: false, error: null, poll: mockPoll };

vi.mock('../stores', () => {
  return {
    useStatusStore: (selector?: (s: MockState) => unknown) => {
      if (typeof selector === 'function') return selector(mockState);
      return mockState;
    },
  };
});

describe('StatusHUD', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockState = { task: null, loading: false, error: null, poll: mockPoll };
  });

  it('shows idle state when no task', () => {
    render(<StatusHUD />);
    expect(screen.getByText('No active task')).toBeDefined();
  });

  it('shows loading state with dot', () => {
    mockState.loading = true;
    render(<StatusHUD />);
    expect(screen.getByText('Loading...')).toBeDefined();
    expect(document.querySelector('.status-dot')).toBeDefined();
  });

  it('shows error state', () => {
    mockState.error = 'Network error';
    render(<StatusHUD />);
    expect(screen.getByText(/Failed to fetch/)).toBeDefined();
  });

  it('shows task details when task loaded', () => {
    mockState.task = {
      id: '1',
      description: 'Add Auth Service',
      status: 'IN_PROGRESS',
      tier: 'T1',
      verification: 'go test ./...',
      created_at: '2026-05-23T14:30:00Z',
    };
    render(<StatusHUD />);
    expect(screen.getByText('Add Auth Service')).toBeDefined();
    expect(screen.getByText('IN_PROGRESS')).toBeDefined();
    expect(screen.getByText('T1')).toBeDefined();
  });
});
