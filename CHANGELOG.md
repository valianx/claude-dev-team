# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **MIT License** (`LICENSE`) — repo is now under MIT, copyright 2026 Mario Gutierrez. `README.md` updated accordingly.
- Contributor README in each top-level system folder: `agents/`, `hooks/`, `skills/`, `chromadb-mcp/`. Each describes the file conventions, how to add or modify artifacts, and routing / runtime details. These READMEs are **not** copied into `~/.claude/` — the installer skips them.

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
