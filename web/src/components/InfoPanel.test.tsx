import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { InfoPanel } from './InfoPanel';

// Mock fetch globally
globalThis.fetch = vi.fn();

describe('InfoPanel', () => {
  const mockNode = {
    id: 'node-1',
    label: 'TestNode',
    type: 'function',
    file_path: 'pkg/test/main.go',
    start_line: 1,
    end_line: 5,
  };

  const baseUrl = 'http://localhost:8080';

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders node basic info and calls onClose when close button clicked', () => {
    (globalThis.fetch as ReturnType<typeof vi.fn>).mockImplementation(() =>
      Promise.resolve({
        ok: true,
        json: () => Promise.resolve({ lines: [], adrs: [] }),
      })
    );

    const onClose = vi.fn();
    render(<InfoPanel node={mockNode} baseUrl={baseUrl} onClose={onClose} />);

    expect(screen.getByText('node-1')).toBeDefined();
    expect(screen.getByText('function')).toBeDefined();
    expect(screen.getByText('pkg/test/main.go')).toBeDefined();
    expect(screen.getByText('L1–5')).toBeDefined();

    const closeBtn = screen.getByLabelText('Close info panel');
    fireEvent.click(closeBtn);
    expect(onClose).toHaveBeenCalled();
  });

  it('fetches code and ADR list on mount', async () => {
    (globalThis.fetch as ReturnType<typeof vi.fn>).mockImplementation((url: string) => {
      if (url.includes('/api/code')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ lines: ['func Test() {', '}'] }),
        });
      }
      if (url.includes('/api/adr')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ adrs: [{ id: '1', title: 'ADR 1', filename: 'adr-1.md' }] }),
        });
      }
      return Promise.reject(new Error('not found'));
    });

    render(<InfoPanel node={mockNode} baseUrl={baseUrl} onClose={vi.fn()} />);

    // Wait for the mocked fetch requests to resolve
    await waitFor(() => {
      expect(screen.getByText('func Test() {')).toBeDefined();
    });
  });
});
