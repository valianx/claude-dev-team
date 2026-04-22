# hooks/

OS-native notification scripts plus the `config.json` template that wires them into Claude Code.

## Files

| File | Purpose |
|---|---|
| `notify-windows.sh` | Windows toast notification via PowerShell. |
| `notify-mac.sh` | macOS notification via `osascript`. |
| `notify-linux.sh` | Linux desktop notification via `notify-send` (libnotify). |
| `config.json` | Per-OS hook template — copy the section for your OS into `~/.claude/settings.json`. |

All scripts are Bash and cross-platform:
- Windows runs them via Git Bash.
- macOS and Linux run them natively.

## Script contract

Each script reads the Claude Code hook payload from stdin (JSON), extracts `last_assistant_message` and `cwd`, and fires an OS-native notification. Scripts exit silently on errors so they never block Claude Code.

**Required runtime dependencies:**
- `python3` — used to parse the JSON payload (preinstalled on macOS and most Linux distros; on Windows + Git Bash, requires a Python install).
- Windows: `powershell.exe` (included in Windows).
- macOS: `osascript` (built-in).
- Linux: `notify-send` (package `libnotify-bin` on Debian/Ubuntu).

## Enabling hooks after install

The installer copies these scripts into `~/.claude/hooks/` but does **not** modify `~/.claude/settings.json`. To activate them:

1. Open `config.json` in this folder.
2. Copy the `hooks` object under your OS key (`windows`, `macos`, or `linux`).
3. Merge it into `~/.claude/settings.json` under the top-level `"hooks"` key.
4. Restart Claude Code.

## Events covered

Each OS section in `config.json` binds three events:

| Event | When it fires |
|---|---|
| `Stop` | Claude finished its turn. |
| `Notification` | Matcher `idle_prompt|permission_prompt` — Claude is waiting for input. |
| `PostToolUseFailure` | A tool invocation failed. |

## Adding support for another OS

1. Add `notify-<os>.sh` following the existing pattern (read stdin, parse JSON, fire native notification, exit silently on failure).
2. Add a matching section to `config.json` under the new OS key.
3. Update the platform label map in `bin/install.py` if needed.
4. Document the new OS's requirements in this README.

## Notes

- `README.md` in this folder is contributor documentation; the installer does **not** copy it to `~/.claude/hooks/`.
- Hooks must stay **generic and portable**. No personal tokens, private URLs, or OpenClaw-style integrations in this folder — those belong in each developer's local `~/.claude/hooks/`.
