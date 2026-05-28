---
name: update
description: Update the th plugin — refresh catalog, download the new version, sync managed CLAUDE.md blocks. Reload is operator-driven.
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

## The update flow has THREE steps — the skill does two, the operator does one

A `th` update is not "refresh catalog + reload". It is three distinct steps, and skipping the middle one leaves `/reload-plugins` with nothing new to activate:

1. **Refresh the catalog** — `claude plugin marketplace update team-harness-marketplace`. Updates the marketplace metadata (`marketplace.json`) so the CLI knows a newer version exists. **This does NOT download any plugin files.**
2. **Download the new version** — `claude plugin update th@team-harness-marketplace`. Fetches the new version into the plugin cache (`~/.claude/plugins/cache/.../th/<new-version>/`). The CLI prints `Restart to apply changes`. **This is the step that actually downloads; the catalog refresh alone does not.**
3. **Activate** — `/reload-plugins` (or restart Claude Code). Loads the downloaded version into the running session.

This skill performs steps 1 and 2 via the `claude` CLI (both are runnable from Bash). It **cannot** perform step 3: `/reload-plugins` and `/plugin …` are Claude Code UI commands with no agent-callable tool. So the skill refreshes, downloads, syncs the managed blocks, reports — then stops. Do not claim the new version is active; it is not until the operator reloads.

(Historical note: `claude plugin update` is only a no-op when the installed version already equals the catalog's latest. Once the catalog refresh in step 1 surfaces a newer version, step 2 does real work — it is mandatory, not optional.)

---

## Contract

1. **Capture the installed version.** Run `claude plugin list`. Parse the block for `th@team-harness-marketplace` and extract its `Version:` value. If the plugin is not listed, report `th plugin not installed via team-harness-marketplace.` and stop — direct the operator to `/plugin install th@team-harness-marketplace`.

2. **Refresh the marketplace catalog.** Run `claude plugin marketplace update team-harness-marketplace`. Surface any error verbatim and stop on failure — do not proceed to the version comparison with stale data.

3. **Read the latest available version.** Read `~/.claude/plugins/marketplaces/team-harness-marketplace/.claude-plugin/marketplace.json` (refreshed by step 2). Take the `version` field of the `th` entry under `plugins`. On Windows the path resolves under the operator's home directory — use the Read tool, not a shell `cat`, so the path is portable.

4. **Compare.** Compare installed (step 1) vs latest (step 3) using semantic-version ordering:
   - **Update available** (latest > installed): proceed to step 5 (download).
   - **Already current** (latest == installed): no download needed; skip to step 6 (block sync still runs). State that the plugin is current; no reload required.
   - **Installed ahead** (installed > latest): unusual; report both versions, note the catalog may not have propagated the latest release yet, and skip the download.

5. **Download the new version** (only when an update is available). Run `claude plugin update th@team-harness-marketplace`. This fetches the new version into the plugin cache and prints `… updated from <X> to <Y>. Restart to apply changes.` Surface any error verbatim and stop on failure. Do NOT skip this — the catalog refresh in step 2 does not download files, so without this step `/reload-plugins` has nothing new to activate.

6. **Sync the managed `~/.claude/CLAUDE.md` blocks (always — idempotent).** This is the recurring counterpart to `/th:setup`'s one-time bootstrap: `/th:setup` runs once to configure MCP servers and workspace mode; `/th:update` keeps the managed blocks aligned on every run. Do NOT tell the operator to re-run `/th:setup` for this — `/th:update` owns the recurring sync.
   - **Source of truth.** The two managed blocks are defined verbatim in the plugin's `skills/setup/SKILL.md`. Read it from the **highest version directory** present under `~/.claude/plugins/cache/team-harness-marketplace/th/` (semver-sorted) — after step 5 that is the just-downloaded version, so the synced blocks match the version the operator is about to activate. Path: `~/.claude/plugins/cache/team-harness-marketplace/th/<latest>/skills/setup/SKILL.md`.
   - **Extract** both blocks, each including its delimiter comments:
     - `<!-- orchestrator-dispatch-rule:start -->` … `<!-- orchestrator-dispatch-rule:end -->`
     - `<!-- nested-dispatch-takeover:start -->` … `<!-- nested-dispatch-takeover:end -->`
   - **Back up** `~/.claude/CLAUDE.md` to `~/.claude/CLAUDE.md.bak-YYYYMMDD-HHMMSS` (UTC) before the first write. If the file does not exist, create it (blocks-only) and skip the backup.
   - **Write each block idempotently:** if both its markers are present in `~/.claude/CLAUDE.md`, replace everything from `:start` to `:end` inclusive with the canonical block; otherwise append the block at the end of the file. Also migrate legacy orchestrator markers (`<!-- th-orchestrator-inline-rule:start -->`, `<!-- th-orchestrator-dispatch-rule:start -->`) by replacing them with the current `orchestrator-dispatch-rule` block.
   - **Never touch anything outside the marker-delimited blocks.** All other content in `~/.claude/CLAUDE.md` is the operator's and is preserved byte-for-byte.
   - **Report** which blocks were synced (`updated` / `inserted` / `already current`). If both were already current, state that no change was needed.

7. **Close with the activation instruction** when a download occurred in step 5:
   ```
   Plugin downloaded into cache.
   Active (this session): <X>  →  Downloaded: <Y>
   Managed CLAUDE.md blocks: <synced summary>

   To activate <Y>, run:
       /reload-plugins        (or restart Claude Code)
   ```
   When already current, omit the activation line and state that no reload is needed.

## Error handling

- If `claude` is not on PATH, report `claude CLI not found on PATH; cannot refresh the marketplace.` and stop.
- If the catalog file is missing after a successful `marketplace update`, report the path checked and stop — do not fabricate a version.
- Surface every CLI error verbatim. No silent retries.

## Important

- This skill is for **plugin installations**. For legacy Go-installer installations, file syncing is a different path (deprecated).
- The skill refreshes the marketplace catalog, downloads the new version into the plugin cache (`claude plugin update`), reports the version delta, and syncs the marker-delimited managed blocks in `~/.claude/CLAUDE.md` to the version being activated. It never edits repository files, never writes `~/.claude/.team-harness.json` (that config is `/th:setup`'s domain), never touches `~/.claude/CLAUDE.md` content outside the managed-block markers, and never reloads the session — the reload/restart is always operator-driven.
- Division of labour with `/th:setup`: setup is the one-time bootstrap (MCP servers, workspace mode, first write of the managed blocks); update is the repeatable command that keeps the catalog and the managed blocks in sync on every run. Re-running setup is never required as part of the update flow.
