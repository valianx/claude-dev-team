# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
