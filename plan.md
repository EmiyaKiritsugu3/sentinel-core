1. Modify `web/src/components/StatusHUD.tsx`
   - Use `replace_with_git_merge_diff` to add `role="alert"` and `aria-live="assertive"` to the error state div.
   - Also add `role="status"` and `aria-live="polite"` to the loading, idle, and active task divs.
   - Concrete diff:
```
<<<<<<< SEARCH
  if (error) {
    return (
      <div className="status-hud status-hud--error">
        ⚠ Failed to fetch status: {error}
      </div>
    );
  }

  if (!task) {
    if (loading) {
      return (
        <div className="status-hud">
          <span className="status-dot" />
          Loading...
        </div>
      );
    }
    return (
      <div className="status-hud status-hud--idle">
        No active task
      </div>
    );
  }

  let time = '';
  if (task.created_at) {
    const date = new Date(task.created_at);
    if (!isNaN(date.getTime())) {
      time = date.toLocaleTimeString();
    }
  }

  return (
    <div className="status-hud">
      <span className="status-hud__label">{task.description}</span>
=======
  if (error) {
    return (
      <div className="status-hud status-hud--error" role="alert" aria-live="assertive">
        ⚠ Failed to fetch status: {error}
      </div>
    );
  }

  if (!task) {
    if (loading) {
      return (
        <div className="status-hud" role="status" aria-live="polite">
          <span className="status-dot" />
          Loading...
        </div>
      );
    }
    return (
      <div className="status-hud status-hud--idle" role="status" aria-live="polite">
        No active task
      </div>
    );
  }

  let time = '';
  if (task.created_at) {
    const date = new Date(task.created_at);
    if (!isNaN(date.getTime())) {
      time = date.toLocaleTimeString();
    }
  }

  return (
    <div className="status-hud" role="status" aria-live="polite">
      <span className="status-hud__label">{task.description}</span>
>>>>>>> REPLACE
```
2. Verify modification
   - Run `cat web/src/components/StatusHUD.tsx` to ensure the diff was applied successfully.
3. Append learning to `.jules/palette.md`
   - Use `mkdir -p .jules` and then append a learning about dynamic status HUDs and `aria-live` attributes to the journal file using:
```bash
cat << 'EOF' >> .jules/palette.md
## 2024-05-15 - Dynamic HUD Accessibility
**Learning:** Dynamic status updates (like StatusHUD) are visually apparent but completely invisible to screen readers without specific ARIA roles.
**Action:** Always assign `role="status"` and `aria-live="polite"` to dynamically updating status containers, and `role="alert"` / `aria-live="assertive"` to error states to ensure immediate screen reader announcements.
