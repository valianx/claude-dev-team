# agents/

System prompts for the subagents of the `claude-dev-team` system. Each `.md` file is a single agent.

## File convention

Every agent file is Markdown with YAML frontmatter:

```md
---
name: orchestrator
description: Central hub that coordinates the pipeline.
model: opus
effort: high
color: blue
---

# Agent body (system prompt)
...
```

**Frontmatter keys.**
- `name` — agent identifier (matches the filename).
- `description` — one-line summary used by the invoker to decide when to route to this agent.
- `model` — `opus` for agents whose work is **analysis or coordination** (cannot fail); `sonnet` for agents whose work is **execution following a plan** (write code, tests, diagrams, commits, docs).
- `effort` — reasoning level when the agent is active. Allowed: `medium` | `high` | `xhigh` | `max`. **`low` is forbidden by project policy** (the floor is `medium`). Tune per agent based on how much judgement the role demands; the matrix in the Roster below is canonical.
- `color` — arbitrary colour label for display.

## Roster

The combination of `model` + `effort` below is the canonical matrix for this repo. `/lint` enforces it — any drift between an agent's frontmatter and this table fails the check.

| Agent | Model | Effort | Role |
|---|---|---|---|
| `orchestrator` | opus | `high` | Central hub. Coordinates the pipeline and routes to all other agents. |
| `architect` | opus | `max` | Architecture design, research, planning, audits. |
| `agent-builder` | opus | `max` | Create / improve agents and skills. |
| `security` | opus | `max` | OWASP / CWE / ASVS audits. |
| `reviewer` | opus | `max` | GitHub PR review. |
| `qa` | opus | `high` | Acceptance criteria definition and validation. |
| `gcp-cost-analyzer` | opus | `high` | GCP cost / resource inventory reports. |
| `init` | opus | `medium` | Bootstrap `CLAUDE.md` in any repo. |
| `implementer` | sonnet | `high` | Production code following the architect's Work Plan. |
| `tester` | sonnet | `medium` | Test suites with factory mocks. |
| `diagrammer` | sonnet | `medium` | Excalidraw diagrams (render-validate loop). |
| `likec4-diagrammer` | sonnet | `medium` | LikeC4 diagrams (architecture-as-code). |
| `d2-diagrammer` | sonnet | `medium` | D2 diagrams. |
| `translator` | sonnet | `medium` | i18n discovery, glossary, translation. |
| `delivery` | sonnet | `medium` | Docs, changelog, version, branch, commit, PR. |

Plus reference files (`ref-direct-modes.md`, `ref-special-flows.md`) loaded on-demand by the orchestrator. They are not invocable subagents — their `model` field is vestigial and not enforced by `/lint`.

## Earn the model AND the effort

Two principles drive the matrix above:

1. **Model by nature of the work.** Agents that do **analysis or coordination** (architect, security, reviewer, qa, gcp-cost-analyzer, agent-builder, init, orchestrator) run on `opus` — a wrong call here cascades through the whole pipeline. Agents that do **execution against a finished plan** (implementer, tester, delivery, diagrammers, translator) run on `sonnet` — the heavy thinking has already been done upstream.
2. **Effort by depth of judgement required.** `max` for irreversible analysis (architecture, security audits, PR reviews, agent design). `high` for solid analytical work that doesn't need exhaustive exploration (orchestrator routing, qa validation, FinOps prioritisation, implementer following a Work Plan). `medium` for everything else, **including the most mechanical tasks** — the floor is `medium`, never `low`.

## Adding or modifying an agent

Per the top-level `CLAUDE.md`, agent changes route through the `architect` subagent first, and the `agent-builder` agent writes the prompt. After editing:

1. Run `./bin/install.sh` (or `uv run bin/install.py`) to propagate into your own `~/.claude/`.
2. Add a `CHANGELOG.md` entry under `[Unreleased]`.
3. Open a PR.

## Notes

- `README.md` in this folder is contributor documentation; the installer does **not** copy it to `~/.claude/agents/`.
- Keep one concern per file. One `.md` = one agent.
- Agent prompts communicate with each other through files in `session-docs/{feature-name}/`, never through return values.
