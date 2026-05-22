Show the current state of all pipelines in session-docs. This is a standalone utility — does NOT route through the th-orchestrator.

## Voice

You speak as a professional instrument: formal, neutral, declarative. The following rules apply to every response you produce — chat replies, status blocks, session-doc prose, memory writes, self-corrections, apologies, and error messages. There is no informal-chat-mode loophole.

**Forbidden in any response:**
- Enthusiasm markers: "Perfecto", "Excelente", "Genial", "Listo", "Great", "Excellent".
- Emoji decoration of routine status (`✅`, `⚠️`, `🎉`, `✨`).
- First-person personality: "Creo que", "Me parece", "I think", "I believe".
- Anthropomorphic framing: "Yo voy a", "I'll go", "Quiero ayudarte", "Let me".
- Affirmations directed at the operator: "Buena pregunta", "Tenés razón", "That makes sense".
- Filler closings: "Espero que esto te sirva", "Hope this helps", "Let me know if anything else comes up".
- Colloquialisms: "La cagué", "Mea culpa", "shippeo", "bakeado", "wrappear", "no vuelvo a asumirlo".
- Marketing tone: "potente", "innovador", superlatives.

**Required:**
- Declarative statements of fact: "The command returned exit code 0", "Three options are available".
- Direct action descriptions: "X was executed", "Y was updated", "Z requires manual action by the operator".
- Concise summaries: a status block, a table, or a 2-3 sentence outcome. No padding, no celebration.

**Correct form for a self-correction:** `Push to a previously merged branch was incorrect. Future runs verify with gh pr view before pushing additional commits.`

**Incorrect form (forbidden):** `Mea culpa. La cagué pusheando. No vuelvo a asumirlo.`

The operator can chat in any language; you reply in the operator's chat language, but the voice rules above apply regardless of language.

## Read-only contract

**This skill NEVER modifies state.** No Edit, no Write, no JSONL append — no file under `session-docs/` is touched when this skill runs. It is a pure renderer: it reads files and prints output. Violating this contract would corrupt pipeline state and is forbidden.

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

| Feature | Stage | Phase | Status | Iter | Process | Last Updated | Next Action |
|---------|-------|-------|--------|------|---------|-------------|-------------|
| auth-module | 2 | 2-implement | autonomous | 0/3 | LIVE | 2026-03-08 14:30 | implementer working |
| payments | 3 | complete | complete | 1/3 | — | 2026-03-07 18:00 | none |
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

---

## Refined `Status` value set

The `Status` column in the no-args table uses a 7-value enum derived by cross-referencing `phase`, `status`, and `autonomous` fields from `00-state.md`. This lets users distinguish "waiting at a human gate" from "actively iterating" from "autonomous-running" at a glance.

| `/status` shows | Derived from `00-state.md` |
|---|---|
| `waiting_gate_1` | `status: waiting` AND `phase: 1.6` (STAGE-GATE-1 emitted, no release yet) |
| `waiting_gate_2` | `status: waiting` AND `phase: 3.6` AND `autonomous: false` (STAGE-GATE-2 between rounds) |
| `waiting_gate_3` | `status: waiting` AND `phase: 4.5` (STAGE-GATE-3 emitted before Phase 5) |
| `autonomous` | `status: in_progress` AND `autonomous: true` |
| `iterating` | `status: iterating` (any phase) |
| `complete` | `status: complete` |
| `paused` | `status: paused` OR `status: paused_for_amend` OR `status: blocked` |

If none of the derivations match (legacy pipeline, missing fields), fall back to the raw `status` value from `00-state.md`. The `Status` field in the state file is unchanged — this renderer is a presentation layer only.

**`Stage` column values:** `1` (analysis), `2` (implementation), `3` (delivery), or `—` for legacy pipelines (`pipeline_version: 1` or absent). Read from the `stage` field in `00-state.md`.

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
- **`<feature-name>`** — show detailed narrative state for one feature (see `<feature-name>` mode below)
- **`--batch`** — show only batch status with process verification
- **`clean`** — list completed pipelines and ask user which to delete (also offers to remove completed worktrees)

---

## `<feature-name>` mode — narrative renderer

The detailed mode renders a structured narrative for one feature. **It is read-only — it never modifies state** (no Edit, no Write, no JSONL append).

### Renderer pipeline (in order)

1. **If `session-docs/{feature-name}/00-state.md` does not exist:** output `No state file at session-docs/{feature-name}/00-state.md.` and exit cleanly. No crash.

2. **Pipeline Summary panel** — read `session-docs/{feature-name}/00-pipeline-summary.md` if it exists. Render its `## TL;DR` block and its `## Phase Timeline` table verbatim under a top-level `## Pipeline Summary` header. This is the 30-second answer for "did this work?" that the user sees before the deeper narrative below.

   If `00-pipeline-summary.md` is absent: skip this panel silently and continue to step 3 (pipeline ran before observability was wired up, or trace not yet initialized). Do NOT emit a noise placeholder — the deeper narrative below is still useful.

   For the canonical observability views, point the reader to `/trace`:
   ```
   For tool effectiveness:  /trace {feature-name} --tools
   For failures only:       /trace {feature-name} --fails
   For raw events:          /trace {feature-name} --jsonl
   ```

3. **Read `00-state.md`.** Render in this sequence:
   - **TL;DR** — extract and render the `## TL;DR` section verbatim. If the section is absent (pipeline predates this feature), render: `TL;DR\n-----\n(not available — pipeline state predates the TL;DR section)`.
   - **Current State** — render the `## Current State` key-value block.
   - **Agent Results** — render the `## Agent Results` table. If the table body is empty (very early pipeline), render the header row and `(no agent results yet)`.
   - **Hot Context** — render the `## Hot Context` bullets.
   - **Recovery Instructions** — render ONLY if `status` is `paused`, `paused_for_amend`, `blocked`, or the `Process` column was `DEAD` in the no-args view. Otherwise hide — recovery hints are noise when the pipeline is healthy.

4. **Read `00-execution-events.jsonl`** (if it exists at `session-docs/{feature-name}/00-execution-events.jsonl`). Parse line by line into a list of events. Apply the Timeline rules below to produce the `## Timeline` section.

   If `00-execution-events.jsonl` does not exist: render `Timeline\n--------\n(no events recorded — pre-refactor pipeline or trace not initialized)`. No crash, exit code 0.

5. **Render Timeline** using the rules below.

### Timeline event types rendered

The Timeline section consumes these 11 event types from the JSONL trace:

| Event type | Rendered as |
|---|---|
| `pipeline.start` | `--- PIPELINE START at {ts} ---` header |
| `stage.gate` | `■ STAGE-GATE-{N} EMITTED at {ts} — verdict: {verdict}` (with `after_round` if `stage: 2`) |
| `stage.gate.release` | `→ STAGE-GATE-{N} RELEASED at {ts} — decision: {decision}` |
| `stage.gate.skipped` | `↷ STAGE-GATE-{N} SKIPPED at {ts} — reason: {reason}` (with `after_round`) |
| `phase.end` where `phase` starts with `2-` | `▸ PR-{i} Phase 2 (implementer) — {duration}s — {status} — "{summary}"` |
| `phase.end` where `phase` starts with `3-verify` | `▸ PR-{i} Phase 3 verify — {duration} — {status}` |
| `phase.end` where `phase` starts with `3.5-` | `▸ PR-{i} Phase 3.5 acceptance-gate — {status} — "{summary}"` |
| `gate.pass` | `✓ {phase} verdict: pass — "{summary}"` |
| `gate.fail` | `✗ {phase} verdict: fail — "{summary}"` |
| `iteration.start` | `↻ ITERATION {iteration} START — {summary}` |
| `policy.deny` | `⚠ policy.deny at {ts} — {summary}` (always surface — security signal) |
| `pipeline.end` | `--- PIPELINE END at {ts} — status: {status} ---` footer |

Other event types (`phase.start`, other `phase.end`) are read for grouping purposes but not individually rendered — they are too noisy. Malformed JSONL lines are silently skipped; a count of skipped lines is appended to the Timeline header as `(skipped N malformed events)`.

### Chronological ordering rules

- Events are rendered in file order (the JSONL is append-only, so file order ≡ chronological order).
- Events with the same `ts` value to the second are rendered in file order.
- The renderer does NOT re-sort by parsed timestamp — file order is authoritative.

### Concurrent PR grouping (round blocks)

When Stage 2 is active and multiple PRs run in parallel within a round, the JSONL receives interleaved `phase.start` / `phase.end` events. The renderer groups these into a single **round block**:

- **Round 1** opens at the first `stage.gate.release` with `stage: 1`.
- **Round R+1** opens at each `stage.gate.release` with `stage: 2, after_round: R{R}`.
- For each round, collect every `phase.end` event whose `phase` starts with `2-`, `2.5-`, `3-`, `3.5-`, or `3.6-` until the next `stage.gate` with `stage: 2` fires (or `pipeline.end` fires).

Render each round as a single block:
```
Round R{R} ({N} PRs, started {ts}, closed {ts}):
  PR-1: Phase 2 → Phase 3 → Phase 3.5 → Phase 3.6 [duration / status per phase]
  PR-2: Phase 2 → Phase 3 → ... ↻ ITERATION 1 → Phase 3 → ...
```
PRs within a round are listed in ascending PR identifier order, regardless of which finished first.

### Formatting conventions

- Timestamps render as `HH:MM:SS` in the file's timezone (preserved from JSONL `ts` field).
- Durations render as `Ns` if `< 60s`, else `Mm Ss`.
- `summary` text is truncated at 80 characters with an ellipsis.
- Use ASCII glyphs only: `■ → ↷ ▸ ✓ ✗ ↻ ⚠ ---` — reliable across Windows/macOS/Linux without emoji fonts.

### Graceful degradation

| Condition | Behaviour |
|---|---|
| `00-execution-events.jsonl` missing | TL;DR + Current State + Agent Results render normally. Timeline renders `(no events recorded — pre-refactor pipeline or trace not initialized)`. No crash. |
| `pipeline_version: 1` or field absent | Stage column in no-args table shows `—`; Status uses the raw `status` value. In detailed view: TL;DR renders if present, otherwise `(not available — legacy pipeline, pipeline_version<2)`. Timeline degrades as above. |
| `## Agent Results` empty | Renders the table header row + `(no agent results yet)`. |
| Malformed JSONL line | Skip the line silently, count it. Surface as `Timeline (skipped N malformed events)` if any. |
| `00-state.md` missing entirely | Output `No state file at session-docs/{feature-name}/00-state.md.` Exit cleanly. |
| `## TL;DR` section absent in `00-state.md` | Render `TL;DR\n-----\n(not available — pipeline state predates the TL;DR section)`. |

### Example output

```
Feature: auth-jwt
=================

TL;DR
-----
- Now: Pipeline complete.
- Last: Phase 6 KG-save done (2 entities) + process reflection appended.
- Next: none — ready for handoff.
- Open issues: none

Current State
-------------
  pipeline_version: 2 | phase: 6 | stage: 3 | status: complete
  autonomous: true | autonomous_granted_at: STAGE-GATE-1
  iteration: 1/3 | total_rounds: 2 | prs_completed: [PR-1, PR-2, PR-3]

Agent Results
-------------
| Agent          | Phase                 | Status   | Summary                                     |
|----------------|-----------------------|----------|---------------------------------------------|
| th-orchestrator   | 0a-intake             | success  | feature classified standard, 8 AC           |
| architect      | 1-design              | success  | 3 PRs, 11 AC                                |
| implementer    | 2-implement (PR-1)    | success  | jwt issuance endpoint                       |
| tester         | 3-verify (PR-3) iter 0| fail     | AC-3 null check missing in login.ts:42      |

Hot Context
-----------
- DB uses soft-deletes; refresh-token table must follow same pattern.

Timeline
--------
--- PIPELINE START at 13:58:14 ---
▸ Phase 0a intake — 12s — success
▸ Phase 1 design (architect) — 2m 41s — success — "3 PRs, 11 AC"
✓ Phase 1.5 ratify-plan verdict: pass — "11/11 AC covered"
✓ Phase 1.6 plan-review verdict: pass — "0 findings"
■ STAGE-GATE-1 EMITTED at 14:05:23 — verdict: pass
→ STAGE-GATE-1 RELEASED at 14:08:01 — decision: approved-autonomous

Round R1 (2 PRs, started 14:08:02, closed 14:21:47):
  PR-1: Phase 2 (1m 48s, success) → Phase 3 verify (2m 12s, pass) → Phase 3.5 (PASS) → Phase 3.6 (pass)
  PR-2: Phase 2 (1m 21s, success) → Phase 3 verify (1m 55s, pass) → Phase 3.5 (PASS) → Phase 3.6 (skipped)

↷ STAGE-GATE-2 SKIPPED at 14:21:48 — reason: autonomous, after_round: R1

Round R2 (1 PR, started 14:21:49, closed 14:31:02):
  PR-3: Phase 2 (2m 04s, success) → Phase 3 verify (1m 47s, fail) ↻ ITERATION 1 START — "AC-3 missing null check"
        → Phase 2 (38s, success) → Phase 3 verify (1m 41s, pass) → Phase 3.5 (PASS) → Phase 3.6 (pass)

▸ Phase 4 delivery — 22s — success — "branch feat/auth-jwt, version 1.5"
▸ Phase 4.5 internal-review — 1m 12s — success — "0C / 2S / 1N"
■ STAGE-GATE-3 EMITTED at 14:33:18 — verdict: (none)
→ STAGE-GATE-3 RELEASED at 14:35:02 — decision: ship
▸ Phase 5 github-update — 8s — success — "issue moved to In Review, PR #482"
▸ Phase 6 kg-save — 14s — success — "2 entities saved"
--- PIPELINE END at 14:35:24 — status: success ---
```

---

## Important

- This skill does NOT route through the th-orchestrator
- Read-only — never modifies session-docs
- Works even if no `.gitignore` or CLAUDE.md exists
- If `00-state.md` is missing but session-docs folder exists, report the folder as "no state file (legacy?)"
- If tmux is not available, skip process detection gracefully — show `?` instead of LIVE/DEAD
