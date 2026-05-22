Re-run the team-harness installer to pull the latest agents, skills, and hooks from the most recent GitHub Release. This is a standalone utility — does NOT route through the orchestrator.

Analyze the input: $ARGUMENTS

---

## What this skill does

1. Detect the host OS.
2. Run the team-harness installer one-liner for that OS, passing `--force` through when the operator supplied it.
3. Capture the installer output (categorised as `installed` / `updated` / `unchanged` / `conflicts`).
4. Render a concise summary back to the operator with the file counts, the conflict list (if any) with their absolute paths, and the resolution path.
5. End with the literal restart reminder.

The installer is idempotent and non-destructive. Conflicts are reported, never overwritten. The `MEMORY_MCP_URL`, `MEMORY_MCP_BEARER`, and `CONTEXT7_API_KEY` values are preserved from the existing `~/.claude.json` unless `--force` is set. This skill never prompts the operator interactively.

---

## Argument parsing

Parse `$ARGUMENTS` for the single supported flag:

- `--force` — pass `--force` through to the installer so it overwrites every conflict.

No other flags are accepted. If the operator passed something else, print the supported usage and stop without running the installer.

```
Usage: /th-update [--force]
  --force   Overwrite conflicting files in ~/.claude/.
```

---

## OS detection

Run this detection first. The result picks the right one-liner and the right `--force` propagation pattern.

```bash
if [ -f /proc/version ] && grep -qi microsoft /proc/version 2>/dev/null; then
  echo "ENV:WSL"
elif [ "$(uname -s)" = "Linux" ] || [ "$(uname -s)" = "Darwin" ]; then
  echo "ENV:NIX"
elif command -v wsl.exe >/dev/null 2>&1 || [ "$OS" = "Windows_NT" ]; then
  echo "ENV:WIN"
else
  echo "ENV:NIX"
fi
```

| Result | Installer one-liner |
|---|---|
| `ENV:NIX` or `ENV:WSL` | bash via curl pipe |
| `ENV:WIN` | PowerShell via `irm`/`iex` (scriptblock form when forwarding `--force`) |

---

## Running the installer

### Unix, macOS, WSL (`ENV:NIX` / `ENV:WSL`)

Without `--force`:
```bash
curl -fsSL https://valianx.github.io/team-harness/install.sh | bash
```

With `--force`:
```bash
curl -fsSL https://valianx.github.io/team-harness/install.sh | bash -s -- --force
```

### Windows PowerShell (`ENV:WIN`)

Without `--force`:
```powershell
irm https://valianx.github.io/team-harness/install.ps1 | iex
```

With `--force` (scriptblock form — `iex` does not forward arguments, so the installer is read into a scriptblock and invoked with `--force` as a positional argument; the bootstrap forwards it to the embedded Go installer via `& $InstallerPath @args`):
```powershell
& ([scriptblock]::Create((irm https://valianx.github.io/team-harness/install.ps1))) --force
```

Capture stdout and stderr from the chosen invocation.

---

## Parsing the installer output

The installer prints a `Summary:` block with four counted buckets:

```
Summary:
  installed: N
  updated:   N
  unchanged: N
  conflicts: N
```

If `conflicts: N` is greater than zero, the installer follows the summary with a `Conflicts (on-disk differs from what this install mode would produce):` block listing one absolute path per line, prefixed by `  - `.

Extract:

- The four integer counts from the `Summary:` block.
- Every absolute path from the conflicts list (one per line under the conflicts block).

---

## Operator-facing summary

Render the following block to the operator. Use declarative facts, no emoji, no enthusiasm.

```
team-harness update
-------------------
installed: N
updated:   N
unchanged: N
conflicts: N
```

If `conflicts > 0`, append:

```
Conflicts:
  - {absolute-path-1}
  - {absolute-path-2}
  ...

Resolution: delete the conflicting file under ~/.claude/ and re-run /th-update,
or re-run with /th-update --force to overwrite every conflict in place.
```

If the installer exited non-zero (download failure, network error, missing release), surface the literal error from stderr instead of the summary block:

```
team-harness update failed
--------------------------
{stderr verbatim}
```

---

## Closing line (mandatory)

After the summary block, emit this exact sentence as the final line of the response:

```
Restart Claude Code to load the new agents and skills.
```

No emoji, no leading marker, no rephrasing.

---

## Important

- This skill does NOT route through the orchestrator.
- This skill does NOT prompt the operator interactively — the installer's preservation logic carries the existing `MEMORY_MCP_URL`, `MEMORY_MCP_BEARER`, and `CONTEXT7_API_KEY` values across updates.
- This skill does NOT compare versions or print a "you have X, latest is Y" line — the installer itself reports unchanged vs updated counts; that is the source of truth.
- This skill does NOT write to `session-docs/`.
- The only operator-visible state mutation comes from the installer itself, which writes under `~/.claude/`.
