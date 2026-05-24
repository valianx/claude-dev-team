"""Tests for the mcpServers preservation logic in bin/install.py.

Run with: python -m unittest tests/test_installer_preservation.py -v

These tests use a temporary directory for CLAUDE_DIR / CLAUDE_JSON so they
never touch the real ~/.claude/.
"""
from __future__ import annotations

import importlib
import importlib.util
import json
import os
import sys
import tempfile
import unittest
from pathlib import Path
from types import ModuleType
from typing import Any


# ---------------------------------------------------------------------------
# Loader: import bin/install.py as a module without executing main()
# ---------------------------------------------------------------------------
_REPO_ROOT = Path(__file__).resolve().parent.parent


def _load_installer() -> ModuleType:
    """Load bin/install.py as a fresh module (does NOT call main())."""
    spec = importlib.util.spec_from_file_location(
        "install",
        _REPO_ROOT / "bin" / "install.py",
    )
    assert spec is not None and spec.loader is not None
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)  # type: ignore[arg-type]
    return module


# ---------------------------------------------------------------------------
# Base test case: provides a temp dir and patches module-level path constants
# ---------------------------------------------------------------------------
class InstallerTestCase(unittest.TestCase):
    """Base class that redirects CLAUDE_DIR / CLAUDE_JSON to a temp directory."""

    def setUp(self) -> None:
        self._tmpdir = tempfile.TemporaryDirectory()
        self.tmp = Path(self._tmpdir.name)
        self.claude_dir = self.tmp / ".claude"
        self.claude_json = self.tmp / ".claude.json"

        # Reload a fresh module so module-level state doesn't bleed between tests.
        self.m = _load_installer()
        self._patch_paths()

    def tearDown(self) -> None:
        self._tmpdir.cleanup()

    def _patch_paths(self) -> None:
        self.m.HOME = self.tmp
        self.m.CLAUDE_DIR = self.claude_dir
        self.m.CLAUDE_JSON = self.claude_json
        self.m.MANIFEST_PATH = self.claude_dir / ".team-harness.json"

    def write_claude_json(self, data: dict[str, Any]) -> None:
        self.claude_json.write_text(json.dumps(data, indent=2) + "\n", encoding="utf-8")

    def read_claude_json(self) -> dict[str, Any]:
        return json.loads(self.claude_json.read_text(encoding="utf-8"))

    def _ensure_claude_dir(self) -> None:
        self.claude_dir.mkdir(parents=True, exist_ok=True)


# ---------------------------------------------------------------------------
# Helper builders
# ---------------------------------------------------------------------------
def _memory_http(url: str = "http://localhost:8080/mcp") -> dict:
    return {"type": "http", "url": url}


def _memory_stdio(path: str = "/fake/.claude/knowledge-graph") -> dict:
    return {
        "type": "stdio",
        "command": "uv",
        "args": ["run", "--directory", path, "python", "-m", "server"],
        "env": {},
    }


def _context7_entry(key: str = "ctx7sk-real-key-12345") -> dict:
    return {
        "type": "http",
        "url": "https://mcp.context7.com/mcp",
        "headers": {"CONTEXT7_API_KEY": key},
    }


# ---------------------------------------------------------------------------
# Tests: _read_existing_mcp_servers
# ---------------------------------------------------------------------------
class TestReadExistingMcpServers(InstallerTestCase):
    def test_returns_empty_when_file_absent(self) -> None:
        result = self.m._read_existing_mcp_servers()
        self.assertEqual(result, {})

    def test_returns_mcp_servers_block(self) -> None:
        self.write_claude_json({"mcpServers": {"memory": _memory_http()}})
        result = self.m._read_existing_mcp_servers()
        self.assertEqual(result, {"memory": _memory_http()})

    def test_returns_empty_on_corrupt_json(self) -> None:
        self.claude_json.write_text("{not: valid json", encoding="utf-8")
        result = self.m._read_existing_mcp_servers()
        self.assertEqual(result, {})

    def test_returns_empty_when_no_mcp_servers_key(self) -> None:
        self.write_claude_json({"other": "stuff"})
        result = self.m._read_existing_mcp_servers()
        self.assertEqual(result, {})


# ---------------------------------------------------------------------------
# Tests: _looks_like_valid_memory_entry
# ---------------------------------------------------------------------------
class TestLooksLikeValidMemoryEntry(InstallerTestCase):
    def test_valid_http_entry(self) -> None:
        self.assertTrue(self.m._looks_like_valid_memory_entry(_memory_http()))

    def test_valid_http_entry_https(self) -> None:
        self.assertTrue(self.m._looks_like_valid_memory_entry(_memory_http("https://example.com/mcp")))

    def test_valid_stdio_entry(self) -> None:
        self.assertTrue(self.m._looks_like_valid_memory_entry(_memory_stdio()))

    def test_empty_dict_is_invalid(self) -> None:
        self.assertFalse(self.m._looks_like_valid_memory_entry({}))

    def test_http_without_url_is_invalid(self) -> None:
        self.assertFalse(self.m._looks_like_valid_memory_entry({"type": "http"}))

    def test_stdio_without_command_is_invalid(self) -> None:
        self.assertFalse(self.m._looks_like_valid_memory_entry({"type": "stdio", "command": ""}))

    def test_unknown_type_is_invalid(self) -> None:
        self.assertFalse(self.m._looks_like_valid_memory_entry({"type": "grpc", "url": "http://x"}))


# ---------------------------------------------------------------------------
# Tests: _is_valid_context7_key
# ---------------------------------------------------------------------------
class TestIsValidContext7Key(InstallerTestCase):
    def test_valid_real_key(self) -> None:
        self.assertTrue(self.m._is_valid_context7_key("ctx7sk-real-key-12345"))

    def test_empty_string_is_invalid(self) -> None:
        self.assertFalse(self.m._is_valid_context7_key(""))

    def test_fake_test_key_is_invalid(self) -> None:
        # The literal placeholder used in PR-5 manual test instructions must never
        # be treated as a real key — it would silently persist a bad value.
        self.assertFalse(self.m._is_valid_context7_key("ctx7sk-fake-test-key"))

    def test_too_short_is_invalid(self) -> None:
        # Must be at least 12 chars including the prefix.
        self.assertFalse(self.m._is_valid_context7_key("ctx7sk-ab"))

    def test_wrong_prefix_is_invalid(self) -> None:
        self.assertFalse(self.m._is_valid_context7_key("sk-real-key-12345"))


# ---------------------------------------------------------------------------
# Tests: prompt_kg_backend — preservation path
# ---------------------------------------------------------------------------
class TestPromptKgBackendPreservation(InstallerTestCase):
    def test_preserves_existing_memory_http(self) -> None:
        """Existing http entry → returned as-is, preserved=True, no env-var query."""
        url = "http://localhost:8080/mcp"
        self.write_claude_json({"mcpServers": {"memory": _memory_http(url)}})
        self.m.force_flag = False

        choice = self.m.prompt_kg_backend()

        self.assertTrue(choice.preserved)
        self.assertFalse(choice.skipped)
        self.assertEqual(choice.backend, "context-harness")
        self.assertEqual(choice.url, url)

    def test_preserves_existing_memory_stdio(self) -> None:
        """Existing stdio entry → returned as-is, preserved=True."""
        self.write_claude_json({"mcpServers": {"memory": _memory_stdio()}})
        self.m.force_flag = False

        choice = self.m.prompt_kg_backend()

        self.assertTrue(choice.preserved)
        self.assertFalse(choice.skipped)
        self.assertEqual(choice.backend, "memory")
        self.assertIsNone(choice.url)

    def test_force_flag_bypasses_preservation(self) -> None:
        """--force means existing entries are ignored; falls through to env var."""
        url = "http://localhost:8080/mcp"
        self.write_claude_json({"mcpServers": {"memory": _memory_http(url)}})
        self.m.force_flag = True

        env_backup = os.environ.copy()
        os.environ["KG_BACKEND"] = "memory"
        try:
            choice = self.m.prompt_kg_backend()
        finally:
            # Restore environment exactly.
            for key in list(os.environ):
                if key not in env_backup:
                    del os.environ[key]
            os.environ.update(env_backup)

        self.assertFalse(choice.preserved)
        self.assertEqual(choice.backend, "memory")

    def test_invalid_existing_entry_falls_through(self) -> None:
        """A malformed entry (empty dict) is not preserved; falls through to env."""
        self.write_claude_json({"mcpServers": {"memory": {}}})
        self.m.force_flag = False

        env_backup = os.environ.copy()
        os.environ["KG_BACKEND"] = "memory"
        try:
            choice = self.m.prompt_kg_backend()
        finally:
            for key in list(os.environ):
                if key not in env_backup:
                    del os.environ[key]
            os.environ.update(env_backup)

        self.assertFalse(choice.preserved)
        self.assertEqual(choice.backend, "memory")

    def test_first_install_no_existing_entry_uses_env(self) -> None:
        """No ~/.claude.json at all → first install; env var is respected normally."""
        # No file written — fresh install scenario.
        self.m.force_flag = False

        env_backup = os.environ.copy()
        os.environ["KG_BACKEND"] = "memory"
        try:
            choice = self.m.prompt_kg_backend()
        finally:
            for key in list(os.environ):
                if key not in env_backup:
                    del os.environ[key]
            os.environ.update(env_backup)

        self.assertFalse(choice.preserved)
        self.assertEqual(choice.backend, "memory")


# ---------------------------------------------------------------------------
# Tests: get_context7_api_key — preservation path
# ---------------------------------------------------------------------------
class TestGetContext7ApiKeyPreservation(InstallerTestCase):
    def _set_env(self, **kwargs: str) -> None:
        for k, v in kwargs.items():
            os.environ[k] = v

    def _unset_env(self, *keys: str) -> None:
        for k in keys:
            os.environ.pop(k, None)

    def test_preserves_existing_real_key_when_no_env(self) -> None:
        """Valid stored key + no env var → preserved without prompting."""
        real_key = "ctx7sk-real-key-99999"
        self.write_claude_json({"mcpServers": {"context7": _context7_entry(real_key)}})
        self.m.force_flag = False
        self._unset_env("CONTEXT7_API_KEY")

        try:
            result = self.m.get_context7_api_key()
        finally:
            self._unset_env("CONTEXT7_API_KEY")

        self.assertEqual(result, real_key)

    def test_preserves_existing_key_when_env_matches(self) -> None:
        """Valid stored key + same env var → preserved (no-op)."""
        real_key = "ctx7sk-real-key-99999"
        self.write_claude_json({"mcpServers": {"context7": _context7_entry(real_key)}})
        self.m.force_flag = False
        self._set_env(CONTEXT7_API_KEY=real_key)

        try:
            result = self.m.get_context7_api_key()
        finally:
            self._unset_env("CONTEXT7_API_KEY")

        self.assertEqual(result, real_key)

    def test_rejects_fake_placeholder_key_and_uses_env(self) -> None:
        """Stored fake placeholder → NOT preserved; env var wins in non-interactive."""
        self.write_claude_json({
            "mcpServers": {"context7": _context7_entry("ctx7sk-fake-test-key")}
        })
        self.m.force_flag = False
        real_env_key = "ctx7sk-real-env-key-abc"
        self._set_env(CONTEXT7_API_KEY=real_env_key)

        try:
            result = self.m.get_context7_api_key()
        finally:
            self._unset_env("CONTEXT7_API_KEY")

        self.assertEqual(result, real_env_key)

    def test_first_install_uses_env_key(self) -> None:
        """No existing ~/.claude.json → env var is used (normal first-install flow)."""
        self.m.force_flag = False
        env_key = "ctx7sk-brand-new-key-xyz"
        self._set_env(CONTEXT7_API_KEY=env_key)

        try:
            result = self.m.get_context7_api_key()
        finally:
            self._unset_env("CONTEXT7_API_KEY")

        self.assertEqual(result, env_key)

    def test_force_flag_ignores_existing_key_and_uses_env(self) -> None:
        """--force: existing valid key is ignored; env var wins."""
        stored_key = "ctx7sk-stored-real-99999"
        env_key = "ctx7sk-new-override-12345"
        self.write_claude_json({"mcpServers": {"context7": _context7_entry(stored_key)}})
        self.m.force_flag = True
        self._set_env(CONTEXT7_API_KEY=env_key)

        try:
            result = self.m.get_context7_api_key()
        finally:
            self._unset_env("CONTEXT7_API_KEY")

        self.assertEqual(result, env_key)


# ---------------------------------------------------------------------------
# Tests: register_mcp_servers — no-write when nothing changed
# ---------------------------------------------------------------------------
class TestRegisterMcpServersNoWrite(InstallerTestCase):
    def _make_kg_choice(
        self,
        backend: str = "context-harness",
        url: str | None = "http://localhost:8080/mcp",
        skipped: bool = False,
        preserved: bool = False,
    ):
        return self.m.KGBackendChoice(
            backend=backend, url=url, skipped=skipped, preserved=preserved
        )

    def test_no_write_when_memory_and_context7_match(self) -> None:
        """register_mcp_servers() returns None and writes no backup when nothing changed."""
        url = "http://localhost:8080/mcp"
        key = "ctx7sk-real-key-99999"
        initial_data = {
            "mcpServers": {
                "memory": _memory_http(url),
                "context7": _context7_entry(key),
            }
        }
        self.write_claude_json(initial_data)
        # Take note of all backup files before the call.
        backups_before = list(self.tmp.glob(".claude.json.bak-*"))

        kg_choice = self._make_kg_choice(url=url, preserved=True)
        backup = self.m.register_mcp_servers(key, kg_choice)

        backups_after = list(self.tmp.glob(".claude.json.bak-*"))
        self.assertIsNone(backup, "Expected no backup when nothing changed")
        self.assertEqual(backups_before, backups_after, "No new backup file should be created")
        # File content must be unchanged.
        self.assertEqual(self.read_claude_json(), initial_data)

    def test_writes_when_memory_differs(self) -> None:
        """register_mcp_servers() DOES write when the memory entry changes."""
        key = "ctx7sk-real-key-99999"
        old_url = "http://old-host:8080/mcp"
        new_url = "http://new-host:9090/mcp"
        self.write_claude_json({
            "mcpServers": {
                "memory": _memory_http(old_url),
                "context7": _context7_entry(key),
            }
        })

        kg_choice = self._make_kg_choice(url=new_url, preserved=False)
        backup = self.m.register_mcp_servers(key, kg_choice)

        self.assertIsNotNone(backup, "Expected a backup when memory entry changes")
        result = self.read_claude_json()
        self.assertEqual(result["mcpServers"]["memory"]["url"], new_url)

    def test_writes_when_context7_key_changes(self) -> None:
        """register_mcp_servers() DOES write when the context7 key changes."""
        url = "http://localhost:8080/mcp"
        old_key = "ctx7sk-old-key-00000"
        new_key = "ctx7sk-new-key-99999"
        self.write_claude_json({
            "mcpServers": {
                "memory": _memory_http(url),
                "context7": _context7_entry(old_key),
            }
        })

        kg_choice = self._make_kg_choice(url=url, preserved=True)
        backup = self.m.register_mcp_servers(new_key, kg_choice)

        self.assertIsNotNone(backup, "Expected a backup when context7 key changes")
        result = self.read_claude_json()
        stored_key = result["mcpServers"]["context7"]["headers"]["CONTEXT7_API_KEY"]
        self.assertEqual(stored_key, new_key)

    def test_creates_file_on_first_install(self) -> None:
        """First install (no ~/.claude.json) writes the file and returns a backup path of None."""
        self.assertFalse(self.claude_json.exists())
        key = "ctx7sk-first-install-key"
        kg_choice = self._make_kg_choice(url="http://localhost:8080/mcp")

        backup = self.m.register_mcp_servers(key, kg_choice)

        # backup_claude_json() returns None when the file didn't exist yet.
        self.assertIsNone(backup)
        result = self.read_claude_json()
        self.assertIn("mcpServers", result)
        self.assertEqual(result["mcpServers"]["memory"]["url"], "http://localhost:8080/mcp")
        stored_key = result["mcpServers"]["context7"]["headers"]["CONTEXT7_API_KEY"]
        self.assertEqual(stored_key, key)


if __name__ == "__main__":
    unittest.main()
