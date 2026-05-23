import { useEffect } from 'react';
import { useStatusStore } from '../stores';
import './StatusHUD.css';

const STATUS_COLORS: Record<string, string> = {
  PENDING:     '#666',
  IN_PROGRESS: '#22c55e',
  AUDITING:    '#eab308',
  DONE:        '#3b82f6',
  FAILED:      '#ef4444',
};

/**
 * Renders a compact colored badge for a task status.
 *
 * Looks up the background color from `STATUS_COLORS` and uses `#555` if the status is unknown.
 *
 * @param status - The status label to display (e.g., `PENDING`, `DONE`)
 * @returns A styled `<span>` element containing the `status` text
 */
function StatusBadge({ status }: { status: string }) {
  const color = STATUS_COLORS[status] ?? '#555';
  return (
    <span
      style={{
        display: 'inline-block',
        background: color,
        color: '#fff',
        padding: '1px 8px',
        borderRadius: 3,
        fontSize: 11,
        fontWeight: 600,
        marginLeft: 8,
      }}
    >
      {status}
    </span>
  );
}

/**
 * Displays a compact HUD showing the current task status from the status store.
 *
 * Renders one of four views: a loading indicator, an error message, an idle message when no task is active, or an active task view showing description, status badge, optional tier, localized creation time, and optional verification command. On mount it triggers an immediate poll and continues polling the store every 2 seconds for `${window.location.hostname}:8080`; the polling interval is cleared on unmount.
 *
 * @returns A JSX element containing the status HUD.
 */
export function StatusHUD() {
  const task    = useStatusStore(s => s.task);
  const loading = useStatusStore(s => s.loading);
  const error   = useStatusStore(s => s.error);
  const poll    = useStatusStore(s => s.poll);

  useEffect(() => {
    const baseUrl = `${window.location.hostname}:8080`;
    poll(baseUrl);
    const id = setInterval(() => poll(baseUrl), 2000);
    return () => clearInterval(id);
  }, [poll]);

  if (loading) {
    return (
      <div className="status-hud">
        <span className="status-dot" />
        Loading...
      </div>
    );
  }

  if (error) {
    return (
      <div className="status-hud status-hud--error">
        ⚠ Failed to fetch status: {error}
      </div>
    );
  }

  if (!task) {
    return (
      <div className="status-hud status-hud--idle">
        No active task
      </div>
    );
  }

  const time = task.created_at
    ? new Date(task.created_at).toLocaleTimeString()
    : '';

  return (
    <div className="status-hud">
      <span className="status-hud__label">{task.description}</span>
      <StatusBadge status={task.status} />
      {task.tier && (
        <span className="status-hud__tier">{task.tier}</span>
      )}
      {time && (
        <span className="status-hud__time">{time}</span>
      )}
      {task.verification && (
        <div className="status-hud__cmd">{task.verification}</div>
      )}
    </div>
  );
}
