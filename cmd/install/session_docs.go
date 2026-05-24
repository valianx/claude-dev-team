package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// promptLogsMode determines the work-logs output mode. Decision priority:
//
//  1. Existing manifest.LogsMode (loaded from .team-harness.json) → always preserved.
//  2. LOGS_MODE env var (first-time or --force installs).
//  3. Interactive TTY → prompt with [l] local / [o] obsidian menu (first-time only).
//  4. No env var, no existing config, no TTY → default to "local" silently.
//
// Once set, the installer never modifies logs config. To change it, the
// operator edits ~/.claude/.team-harness.json directly.
func promptLogsMode() {
	if manifest.LogsMode != "" {
		promptLogsModePreserveOrChange()
		return
	}

	if env := strings.TrimSpace(os.Getenv("LOGS_MODE")); env != "" {
		promptLogsModeFromEnv(env)
		return
	}

	promptLogsModeInteractive()
}

// promptLogsModeFromEnv sets manifest fields from the LOGS_MODE env var.
// Valid values: "local", "obsidian". Any other value exits 1.
// When mode is "obsidian", LOGS_PATH is also required.
func promptLogsModeFromEnv(env string) {
	switch env {
	case "local":
		fmt.Printf("  Work-logs mode: local (loaded from LOGS_MODE env var)\n")
		manifest.LogsMode = "local"
	case "obsidian":
		logsPath := strings.TrimSpace(os.Getenv("LOGS_PATH"))
		if logsPath == "" {
			fmt.Fprintln(os.Stderr, "Error: LOGS_PATH is required when LOGS_MODE=obsidian")
			os.Exit(1)
		}
		fmt.Printf("  Work-logs mode: obsidian → %s (loaded from LOGS_MODE/LOGS_PATH env vars)\n", colorValue(logsPath))
		manifest.LogsMode = "obsidian"
		manifest.LogsPath = logsPath
		manifest.LogsSubfolder = "work-logs"
	default:
		fmt.Fprintf(os.Stderr, "Error: LOGS_MODE=%q is invalid. Accepted values: local, obsidian\n", env)
		os.Exit(1)
	}
}

// promptLogsModePreserveOrChange handles the case where an existing
// manifest.LogsMode was loaded from disk. Always preserves — the installer
// never modifies user-configured logs settings. To change logs-mode, the
// operator edits ~/.claude/.team-harness.json directly.
func promptLogsModePreserveOrChange() {
	displayPath := manifest.LogsMode
	if manifest.LogsMode == "obsidian" && manifest.LogsPath != "" {
		displayPath = fmt.Sprintf("obsidian → %s", manifest.LogsPath)
	}
	fmt.Printf("  Work-logs mode: %s (preserved)\n", displayPath)
}

// promptLogsModeInteractive shows the [l]/[o] menu and, when obsidian is
// selected, prompts for the vault path. Falls back to "local" silently
// when no TTY is available (backward compatibility).
func promptLogsModeInteractive() {
	input := openInteractiveInput()
	if input == nil {
		// Non-interactive with no env var and no existing config: default to local.
		manifest.LogsMode = "local"
		return
	}
	defer input.Close()

	scan := bufio.NewScanner(input)
	fmt.Println("  [l] local     — ./session-docs/{date}_{feature}/ relative to each project (default)")
	fmt.Println("  [o] obsidian  — writes to work-logs/ in an Obsidian vault with metadata")
	fmt.Println()
	choice := promptMenuWith("  Work-logs output [l/o]? [l]: ",
		map[string]bool{"l": true, "o": true}, "l", scan)

	if choice == "l" {
		manifest.LogsMode = "local"
		return
	}

	// Obsidian selected: prompt for vault path.
	fmt.Println()
	fmt.Print("  Absolute path to your Obsidian vault (folder containing .obsidian/): ")
	path := strings.TrimSpace(readLineFrom(scan))
	if path == "" {
		fmt.Fprintln(os.Stderr, "Error: Obsidian vault path cannot be empty.")
		os.Exit(1)
	}
	manifest.LogsMode = "obsidian"
	manifest.LogsPath = path
	manifest.LogsSubfolder = "work-logs"
}
