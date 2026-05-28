---
name: update
description: Refresh the th plugin marketplace catalog, report the available version, and sync managed CLAUDE.md blocks. Reload is operator-driven.
---

Refresh the `team-harness` plugin marketplace catalog, report whether a new `th` release is available, and keep the managed `~/.claude/CLAUDE.md` blocks aligned with the running plugin version. This is a standalone utility — it does NOT route through the orchestrator. It is the repeatable update command; `/th:setup` is the one-time bootstrap and is never part of this flow.

Usage: `/th:update`

Analyze the input: $ARGUMENTS

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

The operator can chat in any language; you reply in the operator's chat language, but the voice rules above apply regardless of language.

---

## Hard limitation — the operator must run `/reload-plugins`

This skill can refresh the marketplace catalog, but it **cannot** reload plugins into the running session. `/reload-plugins` and `/plugin ...` are Claude Code UI commands; no tool, hook, or CLI invocation can drive them from inside an agent. The session keeps running the old plugin bytes until the operator types `/reload-plugins` manually.

This skill therefore does two things and stops: (1) refreshes the catalog, (2) reports the version delta and instructs the operator to reload. Do not claim the new version is active — it is not, until the operator reloads.

**`/plugin update th` is a no-op in this setup.** The GitHub marketplace source pins a commit; `/plugin update` compares the installed version against the same pinned commit and sees no change. The catalog must be refreshed first (`marketplace update`), then the session reloaded (`/reload-plugins`). This skill performs the refresh step.

---

## Contract

1. **Capture the installed version.** Run `claude plugin list`. Parse the block for `th@team-harness-marketplace` and extract its `Version:` value. If the plugin is not listed, report `th plugin not installed via team-harness-marketplace.` and stop — direct the operator to `/plugin install th@team-harness-marketplace`.

2. **Refresh the marketplace catalog.** Run `claude plugin marketplace update team-harness-marketplace`. Surface any error verbatim and stop on failure — do not proceed to the version comparison with stale data.

3. **Read the latest available version.** Read `~/.claude/plugins/marketplaces/team-harness-marketplace/.claude-plugin/marketplace.json` (refreshed by step 2). Take the `version` field of the `th` entry under `plugins`. On Windows the path resolves under the operator's home directory — use the Read tool, not a shell `cat`, so the path is portable.

4. **Compare and report.** Compare installed (step 1) vs latest (step 3) using semantic-version ordering:
   - **Update available** (latest > installed): print a status block with both versions and the next action.
   - **Already current** (latest == installed): state that the catalog was refreshed and the installed version is the latest. The operator may still run `/reload-plugins` if they suspect a same-version content drift, but it is not required.
   - **Installed ahead** (installed > latest): unusual; report both versions and note the catalog may not have propagated the latest release yet.

5. **Always close with the reload instruction** when an update is available:
   ```
   Marketplace catalog refreshed.
   Installed: <X>  →  Available: <Y>

   To activate <Y>, run:
       /reload-plugins

   (/plugin update th is a no-op here — the catalog refresh above is the update step.)
   ```

6. **Sync the managed `~/.claude/CLAUDE.md` blocks (always — idempotent).** This is the recurring counterpart to `/th:setup`'s one-time bootstrap: `/th:setup` runs once to configure MCP servers and workspace mode; `/th:update` keeps the managed blocks aligned on every run. Do NOT tell the operator to re-run `/th:setup` for this — `/th:update` owns the recurring sync.
   - **Source of truth.** The two managed blocks are defined verbatim in this plugin's `skills/setup/SKILL.md`. Derive its path from this skill's own base directory: the base directory ends in `…/skills/update`; the setup contract is its sibling `…/skills/setup/SKILL.md`. Reading from the running version's own cache guarantees the blocks match the plugin version that is actually active.
   - **Extract** both blocks, each including its delimiter comments:
     - `<!-- orchestrator-dispatch-rule:start -->` … `<!-- orchestrator-dispatch-rule:end -->`
     - `<!-- nested-dispatch-takeover:start -->` … `<!-- nested-dispatch-takeover:end -->`
   - **Back up** `~/.claude/CLAUDE.md` to `~/.claude/CLAUDE.md.bak-YYYYMMDD-HHMMSS` (UTC) before the first write. If the file does not exist, create it (blocks-only) and skip the backup.
   - **Write each block idempotently:** if both its markers are present in `~/.claude/CLAUDE.md`, replace everything from `:start` to `:end` inclusive with the canonical block; otherwise append the block at the end of the file. Also migrate legacy orchestrator markers (`<!-- th-orchestrator-inline-rule:start -->`, `<!-- th-orchestrator-dispatch-rule:start -->`) by replacing them with the current `orchestrator-dispatch-rule` block.
   - **Never touch anything outside the marker-delimited blocks.** All other content in `~/.claude/CLAUDE.md` is the operator's and is preserved byte-for-byte.
   - **Report** which blocks were synced (`updated` / `inserted` / `already current`). If both were already current, state that no change was needed.

## Error handling

- If `claude` is not on PATH, report `claude CLI not found on PATH; cannot refresh the marketplace.` and stop.
- If the catalog file is missing after a successful `marketplace update`, report the path checked and stop — do not fabricate a version.
- Surface every CLI error verbatim. No silent retries.

## Important

- This skill is for **plugin installations**. For legacy Go-installer installations, file syncing is a different path (deprecated).
- The skill refreshes the marketplace catalog, reports the version delta, and syncs the marker-delimited managed blocks in `~/.claude/CLAUDE.md` to the running plugin version. It never edits repository files, never writes `~/.claude/.team-harness.json` (that config is `/th:setup`'s domain), never touches `~/.claude/CLAUDE.md` content outside the managed-block markers, and never reloads the session — the reload is always operator-driven.
- Division of labour with `/th:setup`: setup is the one-time bootstrap (MCP servers, workspace mode, first write of the managed blocks); update is the repeatable command that keeps the catalog and the managed blocks in sync on every run. Re-running setup is never required as part of the update flow.
