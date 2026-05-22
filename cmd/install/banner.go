package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ansiSupported returns true when the terminal is likely to render ANSI colour
// escape sequences correctly. It checks for Windows legacy cmd (no ANSI) by
// examining the TERM and COLORTERM env vars and the isTerminal guard.
//
// Criteria (conservative — prefer plain over garbled):
//   - stdin must be an interactive terminal (isTerminal())
//   - TERM must not be "dumb"
//   - On Windows, TERM or COLORTERM must be set (Windows Terminal / Git Bash /
//     VS Code integrated terminal all set one of these; legacy cmd.exe does not)
func ansiSupported() bool {
	if !isTerminal() {
		return false
	}
	term := os.Getenv("TERM")
	if term == "dumb" {
		return false
	}
	// On Windows the native console (cmd.exe / old PowerShell) does not set TERM.
	// Windows Terminal, VS Code, and Git Bash do. Accept COLORTERM as a fallback.
	if isWindowsRuntime() {
		if term == "" && os.Getenv("COLORTERM") == "" {
			return false
		}
	}
	return true
}

// printWelcomeBanner prints a stylised ASCII banner to stdout. It is called
// once at the very start of main(), before any prompt fires. It is never
// called on --version or --help paths.
//
// Design goals:
//   - Cross-platform terminal safe: ASCII only, no emoji, no high-bit chars.
//   - <= 15 lines tall, <= 80 cols wide.
//   - Optional muted ANSI orange tint on the wordmark when the terminal supports
//     colour; plain ASCII fallback otherwise (legacy cmd.exe, CI piped output).
//   - Evokes the orbital hub motif: a central name surrounded by radial lines.
func printWelcomeBanner() {
	if ansiSupported() {
		printBannerColor()
	} else {
		printBannerPlain()
	}
}

// orange is the ANSI SGR sequence for a warm orange (256-colour palette index
// 208). We gate on ansiSupported() before using it, so this never reaches a
// terminal that cannot render it.
const (
	ansiOrange = "\033[38;5;208m"
	ansiReset  = "\033[0m"
	ansiDim    = "\033[2m"
)

func printBannerColor() {
	lines := colorBannerLines()
	for _, l := range lines {
		fmt.Println(l)
	}
}

func printBannerPlain() {
	lines := plainBannerLines()
	for _, l := range lines {
		fmt.Println(l)
	}
}

// colorBannerLines returns the banner lines with ANSI colour applied.
func colorBannerLines() []string {
	plain := plainBannerLines()
	out := make([]string, len(plain))
	for i, l := range plain {
		// Colour the wordmark lines (lines that contain "team-harness").
		if strings.Contains(l, "team-harness") {
			out[i] = ansiOrange + l + ansiReset
		} else if strings.HasPrefix(strings.TrimSpace(l), "//") || strings.HasPrefix(strings.TrimSpace(l), "*") {
			// Dim the orbital ring lines.
			out[i] = ansiDim + l + ansiReset
		} else {
			out[i] = l
		}
	}
	return out
}

// plainBannerLines returns the raw ASCII banner lines (no escape sequences).
// Width: 60 cols max. Height: 13 lines.
//
//	         *   *   *
//	      *           *
//	    *    team-      *
//	    *    harness    *
//	      *           *
//	         *   *   *
//	  orchestrated subagents for Claude Code
func plainBannerLines() []string {
	return []string{
		"",
		"         *   *   *",
		"      *           *",
		"    *                *",
		"    *  team-harness  *",
		"    *                *",
		"      *           *",
		"         *   *   *",
		"",
		"  orchestrated subagents for Claude Code",
		"",
	}
}

// pressEnterToExit pauses with a "Press Enter to exit..." prompt when stdin is
// an interactive terminal (i.e. the install was run by a human double-clicking
// the binary or running it from a shell). In non-interactive mode (CI / script
// / piped input) it returns immediately so automation is never blocked.
//
// Call this at the end of a successful install path, after printSummary.
func pressEnterToExit() {
	if !isTerminal() {
		return
	}
	fmt.Println()
	fmt.Print("Press Enter to exit...")
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')
}
