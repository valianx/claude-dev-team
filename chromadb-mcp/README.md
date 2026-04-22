# chromadb-mcp/

ChromaDB-backed knowledge-graph MCP server. Gives Claude Code semantic memory across projects.

This folder gets installed to `~/.claude/chromadb-mcp/`. The data itself (entities, relations, embeddings) lives in `~/.claude/chromadb/`, which **this repo never touches**. Each developer builds their own local KG as they work.

## Files

| File | Purpose |
|---|---|
| `server.py` | FastMCP server exposing the KG tools (create / search / delete entities, relations, etc.). Runs via `uv run`. |
| `pyproject.toml` | Python project metadata. Requires Python 3.10–3.13. Deps: `chromadb`, `mcp`, `uvicorn`. |
| `uv.lock` | Reproducible lock file for the above. |
| `manage-server.sh` | Optional helper for running the server in **SSE mode** on a fixed port (useful when sharing a KG between Windows and WSL). Default Claude Code usage is **stdio**, which does not need this. |
| `viewer/app.py` | Standalone web UI for inspecting the KG. Run with `uv run viewer/app.py`. |
| `migrate_knowledge.py` | One-shot migration tool from the legacy Memory MCP `knowledge.json` format into ChromaDB. |
| `export.py` | Export the local KG to a portable JSON file (default: `<hostname>-<date>.json`). |
| `import.py` | Merge an exported JSON into the local KG. Non-destructive: existing entities get new observations appended (deduped); local data is never deleted. |

## Runtime contract

The MCP server is registered in `~/.claude.json` by the installer as:

```json
"mcpServers": {
  "memory": {
    "type": "stdio",
    "command": "uv",
    "args": ["run", "--directory", "~/.claude/chromadb-mcp", "python", "-m", "server"],
    "env": {}
  }
}
```

Claude Code spawns the server on demand; no long-running process is required.

## Storage

- **Database path**: `$CHROMADB_PATH` (default `~/.claude/chromadb`).
- Persistent SQLite + HNSW embeddings.
- Embedding model: `all-MiniLM-L6-v2` (~80 MB, downloaded on first run).

## Content policy

The KG is technical memory meant to be shareable between developers. Content rules (what may be stored, what must be filtered out) are defined in [`docs/kg-content-policy.md`](../docs/kg-content-policy.md) and enforced at write time by the orchestrator's Phase 6.

## Sharing a KG

Dev A exports, Dev B imports. No automatic sync, no shared database.

```bash
# Dev A — export
uv run --directory chromadb-mcp/ python export.py --out shared-knowledge/<name>-<date>.json

# Commit and push the file (PR into shared-knowledge/). Review focuses on content.

# Dev B — after pull
uv run --directory chromadb-mcp/ python import.py shared-knowledge/<name>-<date>.json
```

See [`shared-knowledge/README.md`](../shared-knowledge/README.md) for the full workflow.

## Running the web viewer

```bash
uv run --directory chromadb-mcp/ python viewer/app.py
# Opens http://localhost:8420 by default.
```

Useful for curation, policy audits, or inspecting what an agent actually saved.

## SSE mode (advanced)

For setups where both Windows and WSL Claude Code instances should share a single KG, run the server as a long-lived SSE process:

```bash
./chromadb-mcp/manage-server.sh start     # status | stop | restart
```

This listens on `http://localhost:8421/sse`. Update `~/.claude.json` to point the `memory` MCP to that URL if you go this route. Optional — most developers don't need it.

## Notes

- `README.md` in this folder is contributor documentation; the installer does **not** copy it to `~/.claude/chromadb-mcp/`.
- Runtime state (`.venv/`, `__pycache__/`, `.server.pid`, `server.log`) is gitignored and skipped by the installer.
