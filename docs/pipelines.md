# Pipelines reference

This document describes every pipeline the th-orchestrator supports. Each section covers when to use the pipeline, the phases it runs, and the artifacts it produces.

For the day-to-day usage walkthrough, see [`docs/how-it-works.md`](./how-it-works.md). For agent contracts and the full routing table, see [`agents/th-orchestrator.md`](../agents/th-orchestrator.md) and [`agents/ref-special-flows.md`](../agents/ref-special-flows.md).

---

## Feature pipeline (standard SDD flow)

**When to use.** New features, enhancements, API additions, non-trivial refactors, or any work that requires a design decision before implementation. The default pipeline when no special type is detected.

### Phases

| Phase | Agent | Output |
|---|---|---|
| Phase 0a — Classify & Read | th-orchestrator | `00-state.md` initialized, KG session started |
| Phase 0b — Intake | th-orchestrator | `00-task-intake.md` |
| Phase 1 — Design | architect | `01-architecture.md`, `02-task-list.md` |
| Phase 1.5 — Plan Ratification | qa | AC validation against Work Plan |
| Phase 1.6 — Plan Review | plan-reviewer | `01-plan-review.md` — pass/concerns/fail verdict |
| **STAGE-GATE-1** | operator | Approve or approve-autonomous |
| Phase 2.0 — (bug-fix only) | — | — (see Bug-fix pipeline) |
| Phase 2 — Implementation | implementer | code, `02-implementation.md` |
| Phase 2.5 — Constraint Reconciliation | qa | keep/amend/drop decision when a hidden constraint surfaces |
| Phase 3 — Verify | tester, qa, security (parallel) | `03-testing.md`, `04-validation.md`, `04-security.md` |
| Phase 3.5 — Acceptance Gate | th-orchestrator | re-routes to implementer if any AC is missing a passing test |
| Phase 3.6 — Acceptance Check | acceptance-checker | `06-acceptance-check.md` — independent spec-vs-delivery comparison |
| **STAGE-GATE-2** | operator | Per-PR approval (skipped with autonomy) |
| Phase 4 — Delivery | delivery | CHANGELOG entry, version bump, branch, commit |
| Phase 4.5 — Internal Review | reviewer | `05-internal-review.md` — advisory top-3 issues |
| **STAGE-GATE-3** | operator | Final ship/amend/abort |
| Phase 5 — GitHub | delivery | PR opened on GitHub (`Fixes #N`, labels) |
| Phase 6 — KG Capture | th-orchestrator | `process-insight` node written to Memory MCP |

**STAGE-GATE-1** is mandatory and cannot be skipped. **STAGE-GATE-3** is mandatory and cannot be skipped. **STAGE-GATE-2** fires between PR batches and is skipped when the operator granted `approve autonomous` at GATE-1.

### Notable artifacts

- `session-docs/{feature}/01-architecture.md` — design proposal
- `session-docs/{feature}/02-task-list.md` — PR table with Given/When/Then AC per PR and `Status:` field
- `session-docs/{feature}/01-plan-review.md` — plan-reviewer verdict
- `session-docs/{feature}/00-state.md` — live pipeline state (TL;DR + phase + agent results)
- `session-docs/{feature}/00-execution-events.jsonl` — append-only JSONL trace
- `session-docs/{feature}/00-pipeline-summary.md` — human-readable rollup

---

## Bug-fix pipeline (type: fix)

**When to use.** A known bug needs a focused, scoped fix. Triggered when intent signals contain `bug`, `fix`, `solucionar`, `arreglar`, `corregir`, `regresión`, urgency markers, or a GitHub `bug` label. The same 3-stage backbone as the feature pipeline, with type-specific content shifts.

Full specification: [`agents/ref-special-flows.md`](../agents/ref-special-flows.md) § Bug-fix Flow.

### Differences from the feature pipeline

| Stage | Bug-fix difference |
|---|---|
| Stage 1 | architect runs in **root-cause mode** → `01-root-cause.md` (1 page max) instead of `01-architecture.md` |
| Phase 2.0 | tester authors a **failing regression test** in `02-regression-test.md` BEFORE the implementer touches source — mandatory, no fallback |
| Stage 2 — Implementation | implementer runs under **scope discipline**: zero tangential refactors; spotted issues go to `## Follow-ups Spotted` |
| Stage 2 — Verify | `security` agent runs **always** in parallel with `tester` and `qa`, regardless of any other criterion |
| Stage 3 — Delivery | CHANGELOG entry under `### Fixed`; PR title `fix(area): <summary>`; PR body includes mandatory `## Bug Report` section with reproduction steps + root cause + regression test path; `Fixes #N` triggers GitHub auto-close |

### Tier system (1–4)

The bug-fix pipeline is tier-classified at Phase 0a to calibrate ceremony to severity.

| Tier | Name | Phase 1 (root-cause) | Phase 2.0 (regression test) | Phase 3 agents |
|---|---|---|---|---|
| **1** | Docs/Trivial | Skipped — one-sentence prose plan | Conditional skip when no behavior change | tester (no-regress suite) only |
| **2** | Light fix | Architect `mode: light-root-cause`, ≤30 lines | Mandatory | tester + qa |
| **3** | Standard fix (default) | Architect `mode: full-root-cause`, 1 page max | Mandatory | tester + qa + security |
| **4** | Critical/Security | `mode: full-root-cause` + mandatory `mcp__memory__search_nodes` Prior Art query | Mandatory | tester + qa + security (extended analysis) |

**Classification signals.**

- **Signal 1 — Keywords.** Low-tier hints: `typo`, `comment`, `docs`. High-tier triggers: `auth`, `injection`, `token`, `bypass`, `sql`, `xss`, `csrf`, `rce`, `overflow`, `exploit`, `cve`.
- **Signal 2 — File-path patterns.** Tier 1: `*.md`, `docs/**`, `LICENSE`, `CHANGELOG*`. Tier 2: `.github/**`, `scripts/**`, `*.test.*`. Tier 3: `src/**`, `lib/**`, `app/**`, `cmd/**`. Sensitive paths (`auth/**`, `middleware/**`, `api/**`, `db/**`, `security/**`, `crypto/**`, `session/**`) force a minimum of Tier 3, regardless of any other signal.
- **Signal 3 — Operator override.** `[TIER: N]`, `[regression-test: required]`, `[security: required]` markers in the bug report take precedence.

When signals are ambiguous, the default is Tier 3 (conservative). The architect can re-tier mid-flow via `tier_promote` + `tier_promote_rationale` with operator confirmation.

---

## Hotfix sub-flow (type: hotfix)

**When to use.** An urgent single-file or minimal-scope fix that cannot wait for a full root-cause analysis cycle. Triggered by `hotfix` in the request.

Differences from the bug-fix pipeline:

- Phase 1 (architect root-cause) is **skipped entirely**. The th-orchestrator emits a one-sentence prose plan at STAGE-GATE-1 instead.
- Phase 2.0 (regression test) is **still mandatory**. There is no fallback.
- PR title appends `(hotfix)` suffix: `fix(area): <summary> (hotfix)`.

Full specification: [`agents/ref-special-flows.md`](../agents/ref-special-flows.md) § Hotfix sub-flow.

---

## Refactor flow (type: refactor)

**When to use.** Code restructuring with no observable behavior change. Triggered by `refactor`, `rename`, `reorganize`, or similar intent.

Key constraints: the `tester` agent verifies that existing tests still pass (no net-new test coverage is written for a pure refactor). The `qa` agent validates against a `VERIFY-format` AC list (no functional regression). Security runs only if the refactor touches sensitive paths.

---

## Security-sensitive flow

**When to use.** Any change touching `auth/**`, `middleware/**`, `api/**`, `db/**`, `security/**`, `crypto/**`, `session/**`, or flagged by the operator with `[security: required]`.

Security-sensitive flows force `security-sensitive: true` at Phase 0a Step 7, regardless of the pipeline type. The `security` agent runs at Phase 3 in parallel with `tester` and `qa`. For Tier 4 bug-fixes, the security agent runs extended analysis (adjacent-code surface beyond the diff).

---

## Database changes flow

**When to use.** Any feature or fix that adds, removes, or alters database schema, indexes, or migrations.

Migration strategy is mandatory for all database-touching PRs: migrations must be reversible (up + down), follow the project's migration tooling, and be deployed atomically with the code that depends on them. The architect declares the migration strategy in `01-architecture.md` and the plan-reviewer validates it.

---

## Test pipeline (/test-pipeline)

**When to use.** Service-wide test coverage analysis or a structured test pass across multiple components without a feature change. Triggered by `/test-pipeline` or `@th-orchestrator run the test pipeline`.

The `tester` agent runs in coverage mode, reports coverage gaps, and produces a prioritized list of tests to add. No implementation or delivery phases run.

---

## Research / Spike flow (type: research or spike)

**When to use.** Time-boxed investigation of an unknown (technology evaluation, feasibility analysis, performance profiling, cost modeling). No code changes are committed.

The th-orchestrator routes to read-only direct mode: no `implementer`, no `delivery`, no PR. Output is a `01-research.md` spike document with findings, trade-offs, and a recommendation. The operator decides whether to promote to a feature pipeline from there.

---

## Plan flow (type: plan)

**When to use.** Design-only run: the operator wants `01-architecture.md` + `02-task-list.md` but will not immediately implement. Triggered by `/plan`, `/design`, or `@th-orchestrator give me the work plan`.

Runs Stage 1 (Phase 0–1.6 + STAGE-GATE-1) and stops. No implementation dispatched. The operator can resume implementation later via `@th-orchestrator implement it`.

---

## Acceptance gate (Phase 3.5)

**When to use.** Fires automatically between Phase 3 (Verify) and STAGE-GATE-2 for every PR in every pipeline.

Phase 3.5 is the th-orchestrator re-reading the three verify artifacts (`03-testing.md`, `04-validation.md`, `04-security.md`) and the original AC list. If any AC from `02-task-list.md` is missing a passing test or has an unresolved security finding, Phase 3.5 routes back to the `implementer` for a targeted fix before the gate opens. STAGE-GATE-2 never opens on a partial-pass.

---

## gh-fallback graceful degradation

When the `gh` CLI is unavailable or unauthenticated, skills degrade through four tiers rather than failing hard:

| Tier | What it covers |
|---|---|
| A | Read operations via `curl` against the GitHub REST API (requires `$GH_TOKEN` or `$GITHUB_TOKEN`) |
| B | Write operations (PR creation, comments) via `curl` or operator-paste when `curl` write fails |
| C | (reserved) |
| D | Project-board operations skipped silently |

When write via `curl` also fails, `delivery` returns `status: blocked-manual-push`. The th-orchestrator emits a STOP block with the compare URL and `session-docs/{feature}/inputs/pr-body.md`. The operator opens the PR manually, then replies `pr opened #N` to continue.

Full contract: [`agents/_shared/gh-fallback.md`](../agents/_shared/gh-fallback.md).

---

## Multi-reviewer flow (/review-pr --multi)

**When to use.** PRs larger than 1 500 lines or 8 files, or when the operator explicitly requests multiple focused reviews. Triggered via `/review-pr --multi` or `/review-pr --reviewers security,architecture`.

The `reviewer` agent runs 2–3 focused review passes (one per focus: `general`, `security`, `architecture`, `style`). The `reviewer-consolidator` agent then merges the drafts into a single unified PR review, de-duplicates findings, surfaces contradictions, and derives the final verdict. Only the consolidated review is posted to GitHub.

Review policy: if `.team-harness/review-policy.md` exists in the consumer repo, the reviewer reads it and enforces its declared rules. Scaffold via `/init --scaffold-review-policy`.

Re-review automation: optionally scaffold `.github/workflows/team-harness-rereview.yml` via `/init --scaffold-rereview-workflow`. The workflow posts a PR comment when new commits arrive on a PR that already has a team-harness review.
