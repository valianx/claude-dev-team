# Knowledge Graph Content Policy

**Scope.** This policy governs what may be stored in the Knowledge Graph (KG) of the `claude-dev-team` system — in each developer's local KG (`~/.claude/chromadb/`) and in any export shared with the team.

**Guiding principle.** The KG is **technical memory intended to be shareable**. Everything that goes in must be useful to another developer on the team and safe to circulate between machines.

---

## ✅ Allowed

- **Reusable code patterns** (framework conventions, recurring solutions).
- **Gotchas and pitfalls** of libraries, runtimes, or tools (`pnpm`, `Supabase`, `shadcn`, `Drizzle`, etc.).
- **Architectural decisions** with technical rationale (why X over Y, with observed technical constraints).
- **Technical inventories** of a project (service listings, ports, public endpoints, folder structure).
- **Useful commands** specific to a stack (builds, migrations, debugging).
- **Technical conventions** of a project (naming, layout, testing).

## ❌ Forbidden

- **Personal data**: person names (including the developer's own), roles, responsibilities, personal preferences.
- **Personalized feedback**: instructions directed at a specific user ("Mario prefers that…").
- **Credentials and secrets**: tokens, API keys, URLs of private services, internal IPs, absolute paths with user names (`C:/Users/mario/...`).
- **Client / stakeholder data**: client company names, contacts, agreements, contractual information.
- **Volatile temporal references**: ticket numbers, specific PRs, incidents tied to dates, mentions of in-flight releases.
- **Organizational information**: hierarchy, internal policies, non-technical discussions.

## ⚠️ Gray area — requires judgment

- **Business inventories**: internal services described by technical function (OK) vs. described by their relationship to clients or regulation (not OK).
- **Project names**: public or open-source names (OK) vs. confidential internal products (not OK).
- **Metrics**: performance / throughput without business context (OK) vs. revenue, users, KPIs (not OK).

When in doubt, the agent must **omit**. Adding content later is cheap; extracting it from an already-distributed KG is expensive.

---

## How the policy is applied

- **At write time**: the `orchestrator` (and any agent that persists to the KG) must filter content against this policy before calling `create_entities` / `add_observations`. If an observation falls in the forbidden zone, it is omitted silently; if it falls in the gray zone, it is omitted by default.
- **On export**: `export.py` trusts that the local KG already complies — it performs no curation of its own.
- **On import**: `import.py` trusts that the source file already complies — no filtering either.

The filter lives **in one place only**: at write time.

## What to do when a violation is detected

- **In your local KG**: delete the entity / observation via the `/memory` skill or the viewer (`uv run chromadb-mcp/viewer/app.py`).
- **In a shared file**: reject the PR in `shared-knowledge/`, ask the origin to curate and re-export.
- **In an agent's prompt**: open a PR adjusting the agent's prompt to comply.

## Language convention

All KG content is written in **English**, regardless of the conversation language. This aligns with the system's existing convention and makes sharing across teams easier.

---

## Status

**Version**: 0.1 (initial draft, 2026-04-22).

**Implementation status**:
- ✅ Filter wired into `orchestrator.md` Phase 6 (Knowledge Save).
- ✅ `chromadb-mcp/export.py` and `chromadb-mcp/import.py` available.

This policy is **normative for humans and agents**. The orchestrator's filter is the first line of defense at write time; humans curate by reviewing PRs into `shared-knowledge/`.
