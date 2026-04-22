Show the current state of all pipelines in session-docs. This is a standalone utility — does NOT route through the orchestrator.

Analyze the input: $ARGUMENTS

---

## What to scan

1. Use Glob to find all `session-docs/*/00-state.md` files
2. For each found, read the file and extract:
   - Feature name (from folder name)
   - Current phase
   - Status (in_progress, waiting, iterating, complete)
   - Iteration count
   - Last completed phase
   - Next action
   - Last updated timestamp
3. Also check for `session-docs/batch-progress.md` — if found, read and include batch status
4. **Scan worktrees** — run `git worktree list` to find active worktrees. For each worktree path, check if `session-docs/*/00-state.md` exists inside and extract the same fields
5. **Verify live processes** — run `tmux list-sessions 2>/dev/null` (via WSL if on Windows: `wsl -e tmux list-sessions 2>/dev/null`). Map tmux session names to worktree/task names to determine which tasks have a live Claude Code process running

---

## Display format

### If pipelines found

```
Pipeline Status
===============

| Feature | Phase | Status | Iter | Process | Last Updated | Next Action |
|---------|-------|--------|------|---------|-------------|-------------|
| auth-module | 2-implement | in_progress | 0/3 | LIVE | 2026-03-08 14:30 | implementer working |
| payments | complete | complete | 1/3 | — | 2026-03-07 18:00 | none |
```

**Process column values:**
- `LIVE` — tmux session found, Claude Code is actively running
- `DEAD` — worktree exists but no tmux session (process crashed or terminal closed)
- `—` — not a worktree task (running in main session)

### If batch found

```
Batch Status
============

| # | Task | Round | Phase | Status | Process | Branch | PR |
|---|------|-------|-------|--------|---------|--------|----|
| 1 | jwt-setup | 1 | — | DONE | — | feature/101-jwt | #15 |
| 2 | token-service | 2 | — | DONE | — | feature/102-token | #16 |
| 3 | login-endpoint | 3 | 3-verify | RUNNING | LIVE | feature/103-login | — |
| 4 | refresh-flow | 3 | 2-implement | RUNNING | LIVE | feature/104-refresh | — |
| 5 | middleware | 3 | — | RUNNING | DEAD | feature/105-mw | — |

Progress: 2/5 DONE | 2 LIVE | 1 DEAD (needs /recover --batch)
```

Highlight:
- `DEAD` process — needs recovery, suggest `/recover --batch`
- `iterating` status — needs attention
- `complete` / `DONE` status — done
- Stale pipelines (last updated > 1h ago with status != complete) — mark as "stale?"

### If no pipelines found

```
No active pipelines in session-docs/.
```

---

## How to detect live processes

### Step 1 — List worktrees
```bash
git worktree list --porcelain
```
Parse output to get worktree paths and branch names.

### Step 2 — List tmux sessions
```bash
# On WSL/Linux/macOS:
tmux list-sessions -F '#{session_name}:#{session_activity}' 2>/dev/null

# On Windows (via WSL):
wsl -e tmux list-sessions -F '#{session_name}:#{session_activity}' 2>/dev/null
```
If tmux is not available or returns error, skip process detection and show `?` in the Process column.

### Step 3 — Match sessions to tasks
Claude Code worktree sessions typically use the worktree name as part of the tmux session name. Match by checking if the task/feature name appears in the session name.

### Step 4 — Read state from worktrees
For each worktree path, check:
```
{worktree-path}/session-docs/*/00-state.md
```
If found, extract the same fields as regular session-docs.

---

## Actions (optional arguments)

- **No args or `list`** — show the tables above (pipelines + batch + process status)
- **`<feature-name>`** — show detailed state for one feature: read full `00-state.md` including hot context and recovery instructions
- **`--batch`** — show only batch status with process verification
- **`clean`** — list completed pipelines and ask user which to delete (also offers to remove completed worktrees)

---

## Important

- This skill does NOT route through the orchestrator
- Read-only — never modifies session-docs
- Works even if no `.gitignore` or CLAUDE.md exists
- If `00-state.md` is missing but session-docs folder exists, report the folder as "no state file (legacy?)"
- If tmux is not available, skip process detection gracefully — show `?` instead of LIVE/DEAD
