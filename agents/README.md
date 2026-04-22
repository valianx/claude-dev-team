# agents/

System prompts for the subagents of the `claude-dev-team` system. Each `.md` file is a single agent.

## File convention

Every agent file is Markdown with YAML frontmatter:

```md
---
name: orchestrator
description: Central hub that coordinates the pipeline.
model: opus
color: blue
---

# Agent body (system prompt)
...
```

**Frontmatter keys.**
- `name` — agent identifier (matches the filename).
- `description` — one-line summary used by the invoker to decide when to route to this agent.
- `model` — `opus` for agents that cannot fail (`orchestrator`, `architect`, `agent-builder`, `init`), `sonnet` for the rest.
- `color` — arbitrary colour label for display.

## Roster

| Agent | Model | Role |
|---|---|---|
| `orchestrator` | opus | Central hub. Coordinates the pipeline and routes to all other agents. |
| `architect` | opus | Architecture design, research, planning, audits. |
| `implementer` | sonnet | Production code. |
| `tester` | sonnet | Test suites with factory mocks. |
| `qa` | sonnet | Acceptance criteria definition and validation. |
| `delivery` | sonnet | Docs, changelog, version, branch, commit, PR. |
| `reviewer` | sonnet | GitHub PR review. |
| `security` | sonnet | OWASP / CWE / ASVS audits. |
| `init` | opus | Bootstrap `CLAUDE.md` in any repo. |
| `agent-builder` | opus | Create / improve agents and skills. |
| `diagrammer` | sonnet | Excalidraw diagrams (render-validate loop). |
| `likec4-diagrammer` | sonnet | LikeC4 diagrams (architecture-as-code). |
| `d2-diagrammer` | sonnet | D2 diagrams. |
| `translator` | sonnet | i18n discovery, glossary, translation. |
| `gcp-cost-analyzer` | sonnet | GCP cost / resource inventory reports. |

Plus reference files (`ref-direct-modes.md`, `ref-special-flows.md`) loaded on-demand by the orchestrator.

## Adding or modifying an agent

Per the top-level `CLAUDE.md`, agent changes route through the `architect` subagent first, and the `agent-builder` agent writes the prompt. After editing:

1. Run `./bin/install.sh` (or `uv run bin/install.py`) to propagate into your own `~/.claude/`.
2. Add a `CHANGELOG.md` entry under `[Unreleased]`.
3. Open a PR.

## Notes

- `README.md` in this folder is contributor documentation; the installer does **not** copy it to `~/.claude/agents/`.
- Keep one concern per file. One `.md` = one agent.
- Agent prompts communicate with each other through files in `session-docs/{feature-name}/`, never through return values.
