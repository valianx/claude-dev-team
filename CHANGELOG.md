# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **Auto-verification gates before delivery.** The pipeline now has two redundant acceptance gates that prevent shipping when AC are not fully covered:
  - `agents/orchestrator.md` new **Phase 3.5 — Acceptance Gate**: between Verify and Delivery, the orchestrator re-reads `00-task-intake.md`, `03-testing.md`, `04-validation.md`, `04-security.md` and confirms every AC has both a passing test (in tester's AC Coverage) and a `PASS` (in qa's validation). On mismatch, routes back to implementer (still bounded by max 3 iterations) or aborts with `status: blocked` if AC counts diverge between docs.
  - `agents/delivery.md` new **Step 0 — Acceptance Gate**: re-verifies the same artifacts before any branch / commit. Aborts with `status: failed` if any AC is missing PASS, missing a test, or if security has unresolved Critical / High findings.
- **Definition of Done in delivery.** New `agents/delivery.md` Step 9b runs the project's quality gates (lint, typecheck, tests, build — discovered from CLAUDE.md or the project manifest) before staging files. Any failure aborts delivery.
- **Acceptance traceability matrix.** New `agents/delivery.md` Step 9c writes `session-docs/{feature-name}/acceptance-matrix.md` with one row per AC mapping to test (file:line), QA evidence (file:line) and security status. The matrix is embedded in the PR body (Step 11.2 / 11.3) so reviewers see acceptance coverage at a glance.

### Changed

- **Iteration routing now reads only `failure-brief.md`.** When Phase 3 verify fails, the orchestrator no longer re-reads `03-testing.md`, `04-validation.md`, or `04-security.md` in full (5-15K tokens each). Instead, `tester` / `qa` / `security` append a compact iteration entry to `session-docs/{feature-name}/failure-brief.md` as part of their Return Protocol when they fail. The orchestrator reads ONLY the brief to decide routing (Case A / B / C / D). Full session-docs remain available for debugging when the brief is unclear, but happy-path iteration touches only the brief.
- **Batch worktrees emit one-line events instead of copying `00-state.md`.** The Stop hook now writes `{task}|{status}|{summary}` (≤300 bytes) and PostToolUse writes `{task}|{phase}` (~50 bytes) to `/tmp/batch-results/`. Previously each event copied the entire `00-state.md` (5-15K tokens). The parent orchestrator's context now scales linearly at ~300 bytes per task instead of multiple kilobytes; if it needs more detail it opens the worktree's `00-state.md` on demand.

### Fixed

- **Phase 6 token-cost anti-pattern removed.** `agents/orchestrator.md` no longer calls `read_graph` from the Knowledge Save phase. The previous "Auto-consolidate check" loaded the entire knowledge graph (often 100K+ tokens) on every pipeline just to count entities — token cost scaled linearly with KG size. Dedup now relies exclusively on the targeted `search_nodes` call already done in step 2 (vector search, top-N, cheap regardless of graph size). Periodic whole-graph consolidation is surfaced to the user via `/memory consolidate` instead of running automatically.

### Added (earlier in this cycle)

- **Stack guardrails distilled from the knowledge graph.** Recurring pitfalls observed across past pipelines are now codified into the agent prompts so they are caught at design / implementation time, not at runtime:
  - `agents/implementer.md` Phase 0: NestJS + OpenTelemetry guardrail (SDK before `NestFactory.create()`, align the `@opentelemetry/*` family on upgrades, smoke-test runtime after major bumps, `Resource` removal in `@opentelemetry/resources` v2.x).
  - `agents/implementer.md` Phase 2 Frontend: Next.js + shadcn/ui + React guardrails (shadcn v3 vs v4 `asChild` → `render`, Next.js 16 `middleware.ts` deprecation, auto-fetching hook initial state, `next/dynamic({ ssr: false })` skeleton sizing, App Router `loading.tsx` per detail segment, Zustand selector reactivity).
  - `agents/tester.md` new "Common Testing Pitfalls (NestJS / Node)" section: TypeORM entity coverage cap, `setImmediate` mocking pattern, `error?.message || String(error)` branch coverage, env vars before `require()` in Koa/Express controller mocks, fake timers for `moment.utc()` and date-range boundary tests.
  - `agents/architect.md` new "Domain Heuristics" subsection: PostgreSQL high-volume time-series partitioning rules (no `synchronize: true`, partition key in every unique constraint, summary tables for full-history aggregations, TypeORM decimal transformer) and multi-currency / multi-country financial aggregation contract.
  - `agents/delivery.md` new Step 8c: API gateway re-sync notice for services behind Apigee / Kong / AWS API Gateway.

### Changed

- **Knowledge graph write policy hardened.** `agents/orchestrator.md` Phase 6 (Knowledge Save), `skills/memory.md` (consolidate / create paths), and `docs/kg-content-policy.md` now spell out concrete redaction rules with examples drawn from real past violations: no absolute paths with a user identifier (`C:/Users/<name>/...`), no PR / issue numbers, no developer names, `[project]` entities must be named after the bare repo. Each agent that can write to the KG runs a short pre-write checklist before calling `create_entities` / `add_observations`.
- **Earn the model AND the effort.** Reassigned the 15 agents along two dials: `model` (opus for analysis/coordination, sonnet for execution-against-plan) and `effort` (`max` for irreversible analysis, `high` for solid analytical work, `medium` as floor for everything else — `low` is forbidden by policy). Seven agents move to `sonnet` (`implementer`, `tester`, `delivery`, `diagrammer`, `d2-diagrammer`, `likec4-diagrammer`, `translator`); the other eight stay on `opus` with explicit effort levels. The canonical matrix lives in `agents/README.md` and is enforced by a new `/lint` check (Check 7).
- **`/lint` Check 7 added.** Validates that every agent's `model` + `effort` frontmatter matches the canonical matrix in `agents/README.md`, fails on `effort: low`, and warns on unknown agents.
- **`agent-builder.md` "Earn the model" section** rewritten to cover both `model` and `effort` dials and reference the canonical matrix.
- **Notifications default is now quiet.** Removed the `Stop` event from `hooks/config.json` (it fires on every Claude response and creates notification fatigue during active work). The default set is now `Notification` (idle / permission prompts) + `PostToolUseFailure` only. Developers who want a ping when long runs finish can opt in by following the "Opt-in: notify when Claude finishes a turn" section in `hooks/README.md`.

### Added

- **MIT License** (`LICENSE`) — repo is now under MIT, copyright 2026 Mario Gutierrez. `README.md` updated accordingly.
- Contributor README in each top-level system folder: `agents/`, `hooks/`, `skills/`, `chromadb-mcp/`. Each describes the file conventions, how to add or modify artifacts, and routing / runtime details. These READMEs are **not** copied into `~/.claude/` — the installer skips them.
- `chromadb-mcp/README.md` is now the **canonical reference** for every KG operation (view, edit, share, run the server, migrate), replacing scattered docs. Top-level `README.md` points to it.

### Changed

- `bin/install.py` now skips `README.md` files when copying, so contributor docs can live alongside the artifacts without polluting a developer's `~/.claude/`.

## [0.2.0] — 2026-04-22

### Added

- **Manifest-based safe updates.** The installer now writes `~/.claude/.claude-dev-team-manifest.json` tracking which files it installed and their hashes. On re-run, files whose current hash matches the manifest are safely overwritten with the new version (this is a clean update). Files modified locally are still reported as conflicts and left untouched. Adds an `updated` counter to the summary.
- **UTF-8 stdout** forced in `bin/install.py` so Unicode characters (em-dashes, etc.) render correctly in Windows terminals.

### Changed

- **Repo structure simplified.** Moved `hooks-config.json` → `hooks/config.json` (cohesion: all hooks material in one place). Removed `diagram.excalidraw` / `diagram_preview.png` (outdated visuals, will be redone in a future release). Removed `settings.json` from the repo (was personal to the original maintainer).
- **`README.md` rewritten** with installation instructions at the top, target OS and dependency requirements, and a tight overview of what the system ships. English throughout.
- **`docs/kg-content-policy.md` translated to English** to match the system-wide documentation convention.
- **`agents/README` removed** — redundant with `README.md` and out of date.

### Removed

- `settings.json` (personal) — also purged from git history.
- `diagram.excalidraw` and `diagram_preview.png` (obsolete).
- `agents/README` (redundant).

## [0.1.0] — 2026-04-22

Initial release of the `claude-dev-team` agent system distribution.

### Added

- **Installer.** Cross-platform Python installer at `bin/install.py` (PEP 723 inline metadata, zero third-party deps) with bootstrap scripts `bin/install.sh` (Unix / macOS) and `bin/install.ps1` (Windows). Copies agents, skills, hooks, and the ChromaDB MCP server into `~/.claude/`. Non-destructive: existing files with different hashes are reported as conflicts and never overwritten.
- **MCP registration.** Installer surgically merges `mcpServers.memory` and `mcpServers.context7` into `~/.claude.json` with a timestamped backup (`~/.claude.json.bak-YYYYMMDD-HHMMSS`). Prompts for `CONTEXT7_API_KEY` interactively or reads it from the environment.
- **Knowledge-graph MCP server** (`chromadb-mcp/`): stdio FastMCP server, optional SSE runner (`manage-server.sh`), web viewer (`viewer/app.py`), legacy migration tool (`migrate_knowledge.py`).
- **KG sharing.** `chromadb-mcp/export.py` dumps the local KG to JSON; `chromadb-mcp/import.py` merges a JSON into the local KG non-destructively (dedup observations, idempotent relations).
- **`shared-knowledge/` folder** as the agreed drop-off location for KG exports, with a README describing the workflow.
- **KG content policy** (`docs/kg-content-policy.md`) — technical memory only; no personal data, credentials, client/stakeholder info, or volatile references.
- **Policy filter in `orchestrator.md`** Phase 6 (Knowledge Save) enforcing the policy on every `create_entities` / `add_observations` call.
- **macOS notification hook** (`hooks/notify-mac.sh`) for parity with existing Linux and Windows hooks. `hooks-config.json` gained a `macos` section.

### Required dependencies

- `uv` — Python toolchain manager (runs installer and MCP server).
- `gh` — GitHub CLI (used by several skills).
- **context7 API key** — for library docs retrieval.
