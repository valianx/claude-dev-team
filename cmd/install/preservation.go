package main

import (
	"strings"
)

// _FAKE_CONTEXT7_KEY is the placeholder used in manual test instructions.
// It must never be treated as a real key.
const _FAKE_CONTEXT7_KEY = "ctx7sk-fake-test-key"

// looksLikeValidMemoryEntry returns true ONLY for entries that are preservable
// under the v2 http-only model: type=http with a non-empty url that starts
// with http:// or https://.
//
// stdio entries (the v1 shape, pointing at the removed bundled Python KG
// server) deliberately return false here even when they have a non-empty
// command. Reason: the v2 installer always writes http entries; if we
// "preserved" a stdio entry, urlFromEntry would return "" (stdio has no url
// field) and the downstream write would produce {type:"http", url:""}, a
// broken entry. Forcing stdio to fall through to env-var/prompt is the only
// safe migration path. Issue #11 — regression in v2.0.0.
func looksLikeValidMemoryEntry(entry map[string]interface{}) bool {
	if len(entry) == 0 {
		return false
	}
	kind, _ := entry["type"].(string)
	if kind != "http" {
		return false
	}
	url, _ := entry["url"].(string)
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

// isLegacyStdioMemoryEntry returns true if the entry is the v1 stdio shape
// (type=stdio, non-empty command). Used by the prompt to print a migration
// notice — the entry will NOT be preserved (looksLikeValidMemoryEntry
// rejects it) and will be replaced with an http entry derived from env var
// or interactive input.
func isLegacyStdioMemoryEntry(entry map[string]interface{}) bool {
	if len(entry) == 0 {
		return false
	}
	kind, _ := entry["type"].(string)
	if kind != "stdio" {
		return false
	}
	cmd, _ := entry["command"].(string)
	return strings.TrimSpace(cmd) != ""
}

// isValidContext7Key returns true if the key looks like a real context7 key:
//   - non-empty
//   - starts with "ctx7sk-"
//   - length >= 12
//   - is not the fake test placeholder
func isValidContext7Key(key string) bool {
	return key != "" &&
		strings.HasPrefix(key, "ctx7sk-") &&
		len(key) >= 12 &&
		key != _FAKE_CONTEXT7_KEY
}
