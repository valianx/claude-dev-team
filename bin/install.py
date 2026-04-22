#!/usr/bin/env python3
# /// script
# requires-python = ">=3.11"
# dependencies = []
# ///
"""claude-dev-team installer.

Installs agents, skills, hooks, and the ChromaDB MCP server into ~/.claude/,
and registers the `memory` + `context7` MCP servers in ~/.claude.json.

Non-destructive for files that already exist and were customized locally.
Creates a timestamped backup of ~/.claude.json before modifying it.
"""
from __future__ import annotations

import hashlib
import json
import os
import shutil
import stat
import sys
from datetime import datetime
from pathlib import Path

__version__ = "0.1.0"

# ---------------------------------------------------------------------------
# Paths
# ---------------------------------------------------------------------------
SCRIPT_DIR = Path(__file__).resolve().parent
REPO_ROOT = SCRIPT_DIR.parent
HOME = Path.home()
CLAUDE_DIR = HOME / ".claude"
CLAUDE_JSON = HOME / ".claude.json"

# Names to skip when recursing (runtime state, caches, venvs)
SKIP_NAMES = {".venv", "__pycache__", ".server.pid", "server.log"}

_stats: dict[str, list[str]] = {
    "installed": [],
    "unchanged": [],
    "conflicts": [],
}


# ---------------------------------------------------------------------------
# Copy helpers
# ---------------------------------------------------------------------------
def hash_file(p: Path) -> str:
    return hashlib.sha256(p.read_bytes()).hexdigest()


def ensure_dir(p: Path) -> None:
    p.mkdir(parents=True, exist_ok=True)


def should_skip(name: str) -> bool:
    return name in SKIP_NAMES or name.endswith(".pyc")


def copy_file(src: Path, dest: Path, *, executable: bool = False) -> None:
    ensure_dir(dest.parent)

    if dest.exists():
        if hash_file(src) == hash_file(dest):
            _stats["unchanged"].append(str(dest))
        else:
            _stats["conflicts"].append(str(dest))
        return

    shutil.copy2(src, dest)
    if executable and sys.platform != "win32":
        dest.chmod(dest.stat().st_mode | stat.S_IXUSR | stat.S_IXGRP | stat.S_IXOTH)
    _stats["installed"].append(str(dest))


def copy_dir_flat(
    src_dir: Path,
    dest_dir: Path,
    *,
    suffix: str | None = None,
    executable: bool = False,
) -> None:
    if not src_dir.exists():
        return
    for entry in sorted(src_dir.iterdir()):
        if not entry.is_file() or should_skip(entry.name):
            continue
        if suffix and not entry.name.endswith(suffix):
            continue
        copy_file(entry, dest_dir / entry.name, executable=executable)


def copy_dir_recursive(
    src_dir: Path,
    dest_dir: Path,
    *,
    executable_ext: str | None = None,
) -> None:
    if not src_dir.exists():
        return
    for entry in sorted(src_dir.iterdir()):
        if should_skip(entry.name):
            continue
        if entry.is_dir():
            copy_dir_recursive(entry, dest_dir / entry.name, executable_ext=executable_ext)
        elif entry.is_file():
            is_exec = bool(executable_ext and entry.name.endswith(executable_ext))
            copy_file(entry, dest_dir / entry.name, executable=is_exec)


# ---------------------------------------------------------------------------
# Dependency detection
# ---------------------------------------------------------------------------
def require_cli(cmd: str, hint: str) -> None:
    if shutil.which(cmd) is None:
        print(f"Error: required CLI '{cmd}' not found in PATH.", file=sys.stderr)
        print(f"  {hint}", file=sys.stderr)
        sys.exit(1)


def check_dependencies() -> None:
    require_cli("uv", "Install: https://docs.astral.sh/uv/getting-started/installation/")
    require_cli("gh", "Install GitHub CLI: https://cli.github.com/")


# ---------------------------------------------------------------------------
# context7 API key
# ---------------------------------------------------------------------------
def get_context7_api_key() -> str:
    env_key = os.environ.get("CONTEXT7_API_KEY", "").strip()
    if env_key:
        print("  context7 API key: loaded from CONTEXT7_API_KEY env var")
        return env_key

    if not sys.stdin.isatty():
        print("Error: CONTEXT7_API_KEY not set and stdin is not interactive.", file=sys.stderr)
        print("  Export CONTEXT7_API_KEY and re-run.", file=sys.stderr)
        sys.exit(1)

    print("  context7 API key required (get one at https://context7.com/).")
    key = input("  Paste your CONTEXT7_API_KEY: ").strip()
    if not key:
        print("Error: empty API key.", file=sys.stderr)
        sys.exit(1)
    return key


# ---------------------------------------------------------------------------
# ~/.claude.json merge (mcpServers only)
# ---------------------------------------------------------------------------
def backup_claude_json() -> Path | None:
    if not CLAUDE_JSON.exists():
        return None
    timestamp = datetime.now().strftime("%Y%m%d-%H%M%S")
    backup = CLAUDE_JSON.with_name(f"{CLAUDE_JSON.name}.bak-{timestamp}")
    shutil.copy2(CLAUDE_JSON, backup)
    return backup


def register_mcp_servers(context7_key: str) -> Path | None:
    data: dict = {}
    if CLAUDE_JSON.exists():
        with CLAUDE_JSON.open("r", encoding="utf-8") as f:
            data = json.load(f)

    backup = backup_claude_json()

    mcp_dir_posix = (CLAUDE_DIR / "chromadb-mcp").as_posix()

    mcp_servers = data.setdefault("mcpServers", {})
    mcp_servers["memory"] = {
        "type": "stdio",
        "command": "uv",
        "args": ["run", "--directory", mcp_dir_posix, "python", "-m", "server"],
        "env": {},
    }
    mcp_servers["context7"] = {
        "type": "http",
        "url": "https://mcp.context7.com/mcp",
        "headers": {"CONTEXT7_API_KEY": context7_key},
    }

    CLAUDE_JSON.write_text(
        json.dumps(data, indent=2, ensure_ascii=False) + "\n",
        encoding="utf-8",
    )
    return backup


# ---------------------------------------------------------------------------
# Install phases
# ---------------------------------------------------------------------------
def install_agents() -> None:
    copy_dir_flat(REPO_ROOT / "agents", CLAUDE_DIR / "agents", suffix=".md")


def install_skills() -> None:
    skills_src = REPO_ROOT / "skills"

    # Flat .md skills → ~/.claude/commands/
    copy_dir_flat(skills_src, CLAUDE_DIR / "commands", suffix=".md")

    # Complex skills (subdirs) → ~/.claude/skills/<name>/
    if not skills_src.exists():
        return
    for entry in sorted(skills_src.iterdir()):
        if entry.is_dir() and not should_skip(entry.name):
            copy_dir_recursive(entry, CLAUDE_DIR / "skills" / entry.name)


def install_hooks() -> None:
    copy_dir_flat(
        REPO_ROOT / "hooks",
        CLAUDE_DIR / "hooks",
        suffix=".sh",
        executable=True,
    )


def install_chromadb_mcp() -> None:
    copy_dir_recursive(
        REPO_ROOT / "chromadb-mcp",
        CLAUDE_DIR / "chromadb-mcp",
        executable_ext=".sh",
    )


# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
def print_summary(claude_json_backup: Path | None) -> None:
    os_label = {
        "win32": "windows",
        "darwin": "macos",
        "linux": "linux",
    }.get(sys.platform, sys.platform)

    print()
    print("Summary:")
    print(f"  installed: {len(_stats['installed'])}")
    print(f"  unchanged: {len(_stats['unchanged'])}")
    print(f"  conflicts: {len(_stats['conflicts'])}")

    if _stats["conflicts"]:
        print()
        print("Conflicts (left untouched — delete manually and re-run to overwrite):")
        for c in _stats["conflicts"]:
            print(f"  - {c}")

    print()
    print("MCP servers registered in ~/.claude.json:")
    print("  - memory   (ChromaDB-backed knowledge graph)")
    print("  - context7 (library docs)")
    if claude_json_backup:
        print(f"  backup: {claude_json_backup}")

    print()
    print("Next steps:")
    print("  1. Restart Claude Code so it picks up the new MCP servers.")
    print(f'  2. To enable notification hooks, open hooks-config.json in this repo,')
    print(f'     copy the "{os_label}" section, and merge it into')
    print(f'     ~/.claude/settings.json under the "hooks" key.')


# ---------------------------------------------------------------------------
# Entry
# ---------------------------------------------------------------------------
def main() -> None:
    print(f"claude-dev-team installer v{__version__}")
    print(f"  source:   {REPO_ROOT}")
    print(f"  target:   {CLAUDE_DIR}")
    print(f"  platform: {sys.platform}")
    print()

    print("Checking dependencies...")
    check_dependencies()
    print("  uv: ok")
    print("  gh: ok")
    print()

    print("context7 setup:")
    context7_key = get_context7_api_key()
    print()

    print("Installing files...")
    ensure_dir(CLAUDE_DIR)
    install_agents()
    install_skills()
    install_hooks()
    install_chromadb_mcp()

    print("Registering MCP servers in ~/.claude.json...")
    claude_json_backup = register_mcp_servers(context7_key)

    print_summary(claude_json_backup)


if __name__ == "__main__":
    main()
