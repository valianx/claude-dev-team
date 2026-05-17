#!/usr/bin/env python3
# tests/test_agent_frontmatter.py
# Validates that every agent's YAML frontmatter parses as valid YAML.
#
# Catches the failure mode where an unquoted ": " (colon-space) inside a
# description silently breaks YAML parsing. Claude Code's harness handles
# this by silently dropping the agent from the registered `subagent_type`
# list — there is NO error surfaced to the user; the agent just vanishes.
# That happened to `plan-reviewer.md` and stalled Phase 1.6 until found.
#
# Usage:
#   uv run --with PyYAML python tests/test_agent_frontmatter.py
# Exit code:
#   0 if every agent frontmatter parses, 1 otherwise.

from __future__ import annotations

import sys
from pathlib import Path

import yaml

REPO_ROOT = Path(__file__).resolve().parent.parent
AGENTS_DIR = REPO_ROOT / "agents"

REQUIRED_KEYS = ("name", "description", "model")

failures: list[tuple[str, str]] = []
passed = 0

print("=== Agent frontmatter YAML validity ===")
for md_path in sorted(AGENTS_DIR.glob("*.md")):
    if md_path.name == "README.md":
        continue
    content = md_path.read_text(encoding="utf-8")
    if not content.startswith("---"):
        failures.append((md_path.name, "no frontmatter"))
        print(f"  [FAIL] {md_path.name} — no frontmatter")
        continue
    end = content.find("\n---", 3)
    if end < 0:
        failures.append((md_path.name, "frontmatter not closed"))
        print(f"  [FAIL] {md_path.name} — frontmatter not closed")
        continue
    fm = content[3:end].strip()
    try:
        parsed = yaml.safe_load(fm)
    except yaml.YAMLError as e:
        msg = str(e).replace("\n", " | ")
        failures.append((md_path.name, msg))
        print(f"  [FAIL] {md_path.name} — {msg}")
        continue
    if not isinstance(parsed, dict):
        failures.append((md_path.name, f"frontmatter is not a mapping (got {type(parsed).__name__})"))
        print(f"  [FAIL] {md_path.name} — not a mapping")
        continue
    missing = [k for k in REQUIRED_KEYS if k not in parsed]
    if missing:
        msg = f"missing required keys: {missing}"
        failures.append((md_path.name, msg))
        print(f"  [FAIL] {md_path.name} — {msg}")
        continue
    passed += 1
    print(f"  [PASS] {md_path.name}")

print()
print("=" * 60)
total = passed + len(failures)
print(f"  agent frontmatter tests: {passed} passed / {total} total")
print("=" * 60)
if failures:
    print()
    print("Failures:")
    for name, msg in failures:
        print(f"  - {name}: {msg}")
    sys.exit(1)
sys.exit(0)
