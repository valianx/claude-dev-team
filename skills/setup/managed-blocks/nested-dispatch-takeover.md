<!-- nested-dispatch-takeover:start -->
## nested-dispatch-takeover

**When this fires:** A subagent response contains the phrase `Dispatch handoff — top-level Claude takes over now`, or an existing `workspaces/{feature}/00-state.md` has `status: blocked-no-dispatch`. Cause: the orchestrator was invoked from a nested context (another agent, a chained dispatch) and the Claude Code harness stripped its `Task` tool as an anti-recursion safety measure.

**Minimum actions (do NOT ask the operator for permission — the original task is the authorisation):**

1. Parse the `dispatch_handoff` JSON block embedded in the subagent's response (or read it from the `## Handoff` section of `00-state.md` if `state_ref` is provided).
2. Dispatch the named agent directly via `Task(subagent_type={next_agent}, ...)` from the top-level session. Do NOT re-invoke `@th:orchestrator` or any skill that routes through it — that recreates the nested context and the same Task strip happens again.
3. Read the full flow for the detected task type (`agents/ref-special-flows.md` for `fix`/`hotfix`/`docs`; `agents/orchestrator.md` phase sections for `feature`/`refactor`/`enhancement`). Execute EVERY stage and honor EVERY gate of the detected flow — skipping any is a defect, not a shortcut. See the **Takeover Pipeline Manifest** in `docs/subagent-orchestration.md` for the ordered stage/gate list with per-type annotations.

**Full protocol:** see `docs/subagent-orchestration.md` in the `team-harness` repo (Takeover Pipeline Manifest + 8-step takeover contract, handoff JSON schema, `blocked-manual-push` handling).

**Red herring:** if `~/.claude/agents/` does not exist, this is NOT a failure. Plugin-installed agents live under `~/.claude/plugins/cache/.../th/<version>/agents/`. The `subagent_type` strings are namespaced (`th:architect`, `th:implementer`, etc.) and the harness resolves them from the plugin path.

**Path & name resolution:** all `docs/…` and `agents/…` paths referenced above (`docs/subagent-orchestration.md`, `agents/ref-special-flows.md`, `agents/orchestrator.md`) are repo-relative for contributors with a `team-harness` clone. For plugin installs (no repo clone), the same files live under `~/.claude/plugins/cache/team-harness-marketplace/th/<highest-version>/` — resolve `<highest-version>` to the highest semver directory present (multiple versions may be cached after updates; the newest is canonical). The `dispatch_handoff` JSON stores `next_dispatch.agent` in **prefixed** form (`th:architect`) — use it verbatim for `Task(subagent_type=…)`, but **strip the `th:` prefix** to derive the agent's file path (`th:architect` → `agents/architect.md`); team-harness agents are flat, so a prefix-strip suffices.

**Guard:** if `next_dispatch.agent == th:orchestrator`, the handoff is malformed — dispatch the phase agent from `00-state.md` (or `th:architect` at boot), never `th:orchestrator` itself. See `docs/subagent-orchestration.md § dispatch_handoff Schema` for the canonical schema and `§ Takeover Protocol` step 4 for the full consume-side guard.
<!-- nested-dispatch-takeover:end -->