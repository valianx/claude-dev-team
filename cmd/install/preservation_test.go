package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// testEnv sets up a temporary HOME-equivalent directory and patches the
// package-level path variables so all installer functions operate in isolation.
// Returns a cleanup function that restores the original state.
func testEnv(t *testing.T) (tmpDir string, cleanup func()) {
	t.Helper()
	tmp := t.TempDir()

	origClaudeDir := claudeDir
	origClaudeJSON := claudeJSON
	origForce := forceFlag

	claudeDir = filepath.Join(tmp, ".claude")
	claudeJSON = filepath.Join(tmp, ".claude.json")
	forceFlag = false
	manifest.Files = map[string]manifestEntry{}
	manifest.InstalledVersion = ""

	return tmp, func() {
		claudeDir = origClaudeDir
		claudeJSON = origClaudeJSON
		forceFlag = origForce
		manifest.Files = map[string]manifestEntry{}
		manifest.InstalledVersion = ""
	}
}

func writeClaudeJSON(t *testing.T, data map[string]interface{}) {
	t.Helper()
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatalf("writeClaudeJSON marshal: %v", err)
	}
	raw = append(raw, '\n')
	if err := os.WriteFile(claudeJSON, raw, 0o644); err != nil {
		t.Fatalf("writeClaudeJSON write: %v", err)
	}
}

func readClaudeJSON(t *testing.T) map[string]interface{} {
	t.Helper()
	data, err := os.ReadFile(claudeJSON)
	if err != nil {
		t.Fatalf("readClaudeJSON: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("readClaudeJSON unmarshal: %v", err)
	}
	return out
}

// Fixture builders.
func memoryHTTP(url string) map[string]interface{} {
	if url == "" {
		url = "http://localhost:8080/mcp"
	}
	return map[string]interface{}{"type": "http", "url": url}
}

func memoryStdio(path string) map[string]interface{} {
	if path == "" {
		path = "/fake/.claude/knowledge-graph"
	}
	return map[string]interface{}{
		"type":    "stdio",
		"command": "uv",
		"args":    []interface{}{"run", "--directory", path, "python", "-m", "server"},
		"env":     map[string]interface{}{},
	}
}

func context7Entry(key string) map[string]interface{} {
	if key == "" {
		key = "ctx7sk-real-key-12345"
	}
	return map[string]interface{}{
		"type": "http",
		"url":  "https://mcp.context7.com/mcp",
		"headers": map[string]interface{}{
			"CONTEXT7_API_KEY": key,
		},
	}
}

func memChoice(url string, preserved bool) MemoryMCPChoice {
	if url == "" {
		url = "http://localhost:8080/mcp"
	}
	return MemoryMCPChoice{URL: url, Preserved: preserved}
}

// ---------------------------------------------------------------------------
// Tests: readExistingMCPServers
// ---------------------------------------------------------------------------

func TestReadExistingMCPServers_AbsentFile(t *testing.T) {
	_, cleanup := testEnv(t)
	defer cleanup()

	result := readExistingMCPServers()
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}

func TestReadExistingMCPServers_ReturnsMCPServersBlock(t *testing.T) {
	_, cleanup := testEnv(t)
	defer cleanup()

	writeClaudeJSON(t, map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"memory": memoryHTTP(""),
		},
	})
	result := readExistingMCPServers()
	if len(result) == 0 {
		t.Fatal("expected non-empty map")
	}
	mem, ok := result["memory"].(map[string]interface{})
	if !ok || mem["type"] != "http" {
		t.Errorf("unexpected memory entry: %v", result["memory"])
	}
}

func TestReadExistingMCPServers_CorruptJSON(t *testing.T) {
	_, cleanup := testEnv(t)
	defer cleanup()

	if err := os.WriteFile(claudeJSON, []byte("{not: valid json"), 0o644); err != nil {
		t.Fatal(err)
	}
	result := readExistingMCPServers()
	if len(result) != 0 {
		t.Errorf("expected empty map for corrupt JSON, got %v", result)
	}
}

func TestReadExistingMCPServers_NoMCPServersKey(t *testing.T) {
	_, cleanup := testEnv(t)
	defer cleanup()

	writeClaudeJSON(t, map[string]interface{}{"other": "stuff"})
	result := readExistingMCPServers()
	if len(result) != 0 {
		t.Errorf("expected empty map when no mcpServers key, got %v", result)
	}
}

// ---------------------------------------------------------------------------
// Tests: looksLikeValidMemoryEntry
// ---------------------------------------------------------------------------

func TestLooksLikeValidMemoryEntry_ValidHTTP(t *testing.T) {
	if !looksLikeValidMemoryEntry(memoryHTTP("")) {
		t.Error("expected true for valid http entry")
	}
}

func TestLooksLikeValidMemoryEntry_ValidHTTPS(t *testing.T) {
	if !looksLikeValidMemoryEntry(memoryHTTP("https://example.com/mcp")) {
		t.Error("expected true for https entry")
	}
}

func TestLooksLikeValidMemoryEntry_ValidStdio(t *testing.T) {
	if !looksLikeValidMemoryEntry(memoryStdio("")) {
		t.Error("expected true for valid stdio entry (legacy shape preserved)")
	}
}

func TestLooksLikeValidMemoryEntry_EmptyDict(t *testing.T) {
	if looksLikeValidMemoryEntry(map[string]interface{}{}) {
		t.Error("expected false for empty dict")
	}
}

func TestLooksLikeValidMemoryEntry_HTTPWithoutURL(t *testing.T) {
	if looksLikeValidMemoryEntry(map[string]interface{}{"type": "http"}) {
		t.Error("expected false for http without url")
	}
}

func TestLooksLikeValidMemoryEntry_StdioWithoutCommand(t *testing.T) {
	if looksLikeValidMemoryEntry(map[string]interface{}{"type": "stdio", "command": ""}) {
		t.Error("expected false for stdio with empty command")
	}
}

func TestLooksLikeValidMemoryEntry_UnknownType(t *testing.T) {
	if looksLikeValidMemoryEntry(map[string]interface{}{"type": "grpc", "url": "http://x"}) {
		t.Error("expected false for unknown type")
	}
}

// ---------------------------------------------------------------------------
// Tests: isValidContext7Key
// ---------------------------------------------------------------------------

func TestIsValidContext7Key_ValidRealKey(t *testing.T) {
	if !isValidContext7Key("ctx7sk-real-key-12345") {
		t.Error("expected true for valid real key")
	}
}

func TestIsValidContext7Key_EmptyString(t *testing.T) {
	if isValidContext7Key("") {
		t.Error("expected false for empty string")
	}
}

func TestIsValidContext7Key_FakeTestKey(t *testing.T) {
	if isValidContext7Key("ctx7sk-fake-test-key") {
		t.Error("expected false for fake test placeholder key")
	}
}

func TestIsValidContext7Key_TooShort(t *testing.T) {
	// "ctx7sk-ab" is 9 chars — below the 12-char minimum.
	if isValidContext7Key("ctx7sk-ab") {
		t.Error("expected false for key shorter than 12 chars")
	}
}

func TestIsValidContext7Key_WrongPrefix(t *testing.T) {
	if isValidContext7Key("sk-real-key-12345") {
		t.Error("expected false for key with wrong prefix")
	}
}

// ---------------------------------------------------------------------------
// Tests: promptMemoryMCPURL — preservation path
// ---------------------------------------------------------------------------

func TestPromptMemoryMCPURL_PreservesExistingHTTPEntry(t *testing.T) {
	_, cleanup := testEnv(t)
	defer cleanup()

	url := "https://my-mcp.example.com/mcp"
	writeClaudeJSON(t, map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"memory": memoryHTTP(url),
		},
	})
	forceFlag = false

	choice := promptMemoryMCPURL()

	if !choice.Preserved {
		t.Error("expected Preserved=true for existing valid http entry")
	}
	if choice.URL != url {
		t.Errorf("expected URL=%s, got %s", url, choice.URL)
	}
}

func TestPromptMemoryMCPURL_PreservesExistingStdioEntry(t *testing.T) {
	// Legacy stdio shape from pre-cutover installs must be preserved as-is.
	_, cleanup := testEnv(t)
	defer cleanup()

	writeClaudeJSON(t, map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"memory": memoryStdio(""),
		},
	})
	forceFlag = false

	choice := promptMemoryMCPURL()

	if !choice.Preserved {
		t.Error("expected Preserved=true for existing valid stdio entry")
	}
	// Stdio entries have no URL; urlFromEntry returns "".
	if choice.URL != "" {
		t.Errorf("expected empty URL for stdio entry, got %s", choice.URL)
	}
}

func TestPromptMemoryMCPURL_ForceFlagBypassesPreservation(t *testing.T) {
	_, cleanup := testEnv(t)
	defer cleanup()

	writeClaudeJSON(t, map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"memory": memoryHTTP("http://old-host/mcp"),
		},
	})
	forceFlag = true
	t.Setenv("MEMORY_MCP_URL", "https://new-host.example.com/mcp")

	choice := promptMemoryMCPURL()

	if choice.Preserved {
		t.Error("expected Preserved=false with --force")
	}
	if choice.URL != "https://new-host.example.com/mcp" {
		t.Errorf("expected env URL, got %s", choice.URL)
	}
}

func TestPromptMemoryMCPURL_EnvVarHighestPriority(t *testing.T) {
	_, cleanup := testEnv(t)
	defer cleanup()

	// No existing entry → env var is the source.
	forceFlag = false
	envURL := "https://railway.app/mcp"
	t.Setenv("MEMORY_MCP_URL", envURL)

	choice := promptMemoryMCPURL()

	if choice.Preserved {
		t.Error("expected Preserved=false when using env var")
	}
	if choice.URL != envURL {
		t.Errorf("expected URL=%s, got %s", envURL, choice.URL)
	}
}

func TestPromptMemoryMCPURL_NonInteractiveDefault(t *testing.T) {
	// Stdin is not a TTY in test execution, so non-interactive path runs.
	_, cleanup := testEnv(t)
	defer cleanup()

	forceFlag = false
	t.Setenv("MEMORY_MCP_URL", "")

	choice := promptMemoryMCPURL()

	if choice.Preserved {
		t.Error("expected Preserved=false for fresh install non-interactive")
	}
	if choice.URL != defaultMemoryMCPURL {
		t.Errorf("expected default URL=%s, got %s", defaultMemoryMCPURL, choice.URL)
	}
}

func TestPromptMemoryMCPURL_InvalidURLRejected(t *testing.T) {
	// validateMCPURL is the guard. Test the validator directly.
	cases := []struct {
		name string
		url  string
		want bool // true = valid, false = invalid
	}{
		{"empty", "", false},
		{"bare word", "memory", false},
		{"ftp scheme", "ftp://example.com/mcp", false},
		{"http valid", "http://localhost:7654/mcp", true},
		{"https valid", "https://my-service.com/mcp", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateMCPURL(tc.url)
			if tc.want && err != nil {
				t.Errorf("expected valid URL, got error: %v", err)
			}
			if !tc.want && err == nil {
				t.Error("expected invalid URL to return error")
			}
		})
	}
}

func TestPromptMemoryMCPURL_TrimsWhitespace(t *testing.T) {
	_, cleanup := testEnv(t)
	defer cleanup()

	forceFlag = false
	t.Setenv("MEMORY_MCP_URL", "  https://trimmed.example.com/mcp  ")

	choice := promptMemoryMCPURL()

	if choice.URL != "https://trimmed.example.com/mcp" {
		t.Errorf("expected trimmed URL, got %q", choice.URL)
	}
}

// ---------------------------------------------------------------------------
// Tests: registerMCPServers — no-write when nothing changed
// ---------------------------------------------------------------------------

func TestRegisterMCPServers_NoWriteWhenNothingChanged(t *testing.T) {
	_, cleanup := testEnv(t)
	defer cleanup()

	url := "http://localhost:8080/mcp"
	key := "ctx7sk-real-key-99999"
	initialData := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"memory":   memoryHTTP(url),
			"context7": context7Entry(key),
		},
	}
	writeClaudeJSON(t, initialData)

	backupsBefore, _ := filepath.Glob(claudeJSON + ".bak-*")

	mc := memChoice(url, true)
	backup := registerMCPServers(key, mc)

	backupsAfter, _ := filepath.Glob(claudeJSON + ".bak-*")
	if backup != "" {
		t.Errorf("expected no backup when nothing changed, got %s", backup)
	}
	if len(backupsAfter) != len(backupsBefore) {
		t.Errorf("unexpected backup file created: before=%d after=%d", len(backupsBefore), len(backupsAfter))
	}
}

func TestRegisterMCPServers_WritesWhenMemoryDiffers(t *testing.T) {
	_, cleanup := testEnv(t)
	defer cleanup()

	key := "ctx7sk-real-key-99999"
	oldURL := "http://old-host:8080/mcp"
	newURL := "http://new-host:9090/mcp"
	writeClaudeJSON(t, map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"memory":   memoryHTTP(oldURL),
			"context7": context7Entry(key),
		},
	})

	mc := memChoice(newURL, false)
	backup := registerMCPServers(key, mc)

	if backup == "" {
		t.Error("expected a backup when memory entry changes")
	}
	result := readClaudeJSON(t)
	mcpServers, _ := result["mcpServers"].(map[string]interface{})
	mem, _ := mcpServers["memory"].(map[string]interface{})
	if mem["url"] != newURL {
		t.Errorf("expected url=%s, got %v", newURL, mem["url"])
	}
}

func TestRegisterMCPServers_WritesWhenContext7KeyChanges(t *testing.T) {
	_, cleanup := testEnv(t)
	defer cleanup()

	url := "http://localhost:8080/mcp"
	oldKey := "ctx7sk-old-key-00000"
	newKey := "ctx7sk-new-key-99999"
	writeClaudeJSON(t, map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"memory":   memoryHTTP(url),
			"context7": context7Entry(oldKey),
		},
	})

	mc := memChoice(url, true)
	backup := registerMCPServers(newKey, mc)

	if backup == "" {
		t.Error("expected a backup when context7 key changes")
	}
	result := readClaudeJSON(t)
	mcpServers, _ := result["mcpServers"].(map[string]interface{})
	c7, _ := mcpServers["context7"].(map[string]interface{})
	headers, _ := c7["headers"].(map[string]interface{})
	storedKey, _ := headers["CONTEXT7_API_KEY"].(string)
	if storedKey != newKey {
		t.Errorf("expected context7 key=%s, got %s", newKey, storedKey)
	}
}

func TestRegisterMCPServers_CreatesFileOnFirstInstall(t *testing.T) {
	_, cleanup := testEnv(t)
	defer cleanup()

	if _, err := os.Stat(claudeJSON); !os.IsNotExist(err) {
		t.Fatal("expected claudeJSON to not exist at start of test")
	}
	key := "ctx7sk-first-install-key"
	mc := memChoice("http://localhost:8080/mcp", false)

	backup := registerMCPServers(key, mc)

	// backupClaudeJSON returns "" when the file didn't exist yet.
	if backup != "" {
		t.Errorf("expected no backup for first install (file didn't exist), got %s", backup)
	}
	result := readClaudeJSON(t)
	mcpServers, ok := result["mcpServers"].(map[string]interface{})
	if !ok {
		t.Fatal("expected mcpServers key in result")
	}
	mem, _ := mcpServers["memory"].(map[string]interface{})
	if mem["url"] != "http://localhost:8080/mcp" {
		t.Errorf("unexpected memory url: %v", mem["url"])
	}
	c7, _ := mcpServers["context7"].(map[string]interface{})
	headers, _ := c7["headers"].(map[string]interface{})
	storedKey, _ := headers["CONTEXT7_API_KEY"].(string)
	if storedKey != key {
		t.Errorf("expected context7 key=%s, got %s", key, storedKey)
	}
}

// ---------------------------------------------------------------------------
// Tests: getContext7APIKey — preservation path
// ---------------------------------------------------------------------------

func TestGetContext7APIKey_PreservesExistingRealKeyWhenNoEnv(t *testing.T) {
	_, cleanup := testEnv(t)
	defer cleanup()

	realKey := "ctx7sk-real-key-99999"
	writeClaudeJSON(t, map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"context7": context7Entry(realKey),
		},
	})
	forceFlag = false
	t.Setenv("CONTEXT7_API_KEY", "")

	result := getContext7APIKey()
	if result != realKey {
		t.Errorf("expected %s, got %s", realKey, result)
	}
}

func TestGetContext7APIKey_PreservesExistingKeyWhenEnvMatches(t *testing.T) {
	_, cleanup := testEnv(t)
	defer cleanup()

	realKey := "ctx7sk-real-key-99999"
	writeClaudeJSON(t, map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"context7": context7Entry(realKey),
		},
	})
	forceFlag = false
	t.Setenv("CONTEXT7_API_KEY", realKey)

	result := getContext7APIKey()
	if result != realKey {
		t.Errorf("expected %s, got %s", realKey, result)
	}
}

func TestGetContext7APIKey_RejectsFakePlaceholderAndUsesEnv(t *testing.T) {
	_, cleanup := testEnv(t)
	defer cleanup()

	writeClaudeJSON(t, map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"context7": context7Entry("ctx7sk-fake-test-key"),
		},
	})
	forceFlag = false
	realEnvKey := "ctx7sk-real-env-key-abc"
	t.Setenv("CONTEXT7_API_KEY", realEnvKey)

	result := getContext7APIKey()
	if result != realEnvKey {
		t.Errorf("expected %s, got %s", realEnvKey, result)
	}
}

func TestGetContext7APIKey_FirstInstallUsesEnvKey(t *testing.T) {
	_, cleanup := testEnv(t)
	defer cleanup()

	// No existing ~/.claude.json.
	forceFlag = false
	envKey := "ctx7sk-brand-new-key-xyz"
	t.Setenv("CONTEXT7_API_KEY", envKey)

	result := getContext7APIKey()
	if result != envKey {
		t.Errorf("expected %s, got %s", envKey, result)
	}
}

func TestGetContext7APIKey_ForceFlagIgnoresExistingKeyAndUsesEnv(t *testing.T) {
	_, cleanup := testEnv(t)
	defer cleanup()

	storedKey := "ctx7sk-stored-real-99999"
	envKey := "ctx7sk-new-override-12345"
	writeClaudeJSON(t, map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"context7": context7Entry(storedKey),
		},
	})
	forceFlag = true
	t.Setenv("CONTEXT7_API_KEY", envKey)

	result := getContext7APIKey()
	if result != envKey {
		t.Errorf("expected %s, got %s", envKey, result)
	}
}
